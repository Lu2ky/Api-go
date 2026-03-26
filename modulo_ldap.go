package main

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	"unicode/utf16"

	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
	"github.com/golang-jwt/jwt/v5"
)

//	------------------------ FUNCIONALIDADES DEL LDAP ------------------------ //

func auth(c *gin.Context) {
	var User UserAuth
	err := c.BindJSON(&User)
	if err != nil {
		c.JSON(400, gin.H{"error": "formato invalido de json"})
		return
	}
	token, userU, err := ConnectLDAP(User.User, User.Pass, JWTManager{
		Secret: []byte(os.Getenv("JWT_SECRET")),
		TTL:    24 * time.Hour,
		Issuer: "horario_estudiantes",
	})
	if err != nil {
		log.Printf("ldap error: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	var userID int
	err = db.QueryRow(
		"SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario = ?",
		User.User,
	).Scan(&userID)

	if err != nil {
		log.Println("Error obteniendo usuario para log:", err)
		userID = 0
	}


	insertarLog(userID,"LOGIN","El usuario inició sesión",)
	c.JSON(200, gin.H{
		"Token":    token,
		"UserAuth": userU,
	})
}

func (j JWTManager) Generate(u *User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: u.Username,
		Roles:  u.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-clockSkewTolerance)),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.TTL)),
			Subject:   u.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.Secret)
}

func (j JWTManager) Validate(tokenStr string) (*Claims, error) {
	if tokenStr == "" {
		return nil, errors.New("sin token")
	}

	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("erro en el metodo de inicio")
		}
		return j.Secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("token invalido")
	}

	if j.Issuer != "" && claims.Issuer != j.Issuer {
		return nil, errors.New("que?")
	}

	return claims, nil
}

func dialLDAPS() (*ldap.Conn, error) {
	return ldap.DialURL("ldaps://"+os.Getenv("LDAP_ADDR")+":636",
		ldap.DialWithTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
		}),
	)
}

func ConnectLDAP(user string, pass string, j JWTManager) (string, *User, error) {
	l, err := dialLDAPS()
	if err != nil {
		return "", nil, err
	}
	defer l.Close()

	l.SetTimeout(5 * time.Second)

	err = l.Bind(user+"@upbplanner.local", pass)
	if err != nil {
		return "", nil, err
	}

	searchRequest := ldap.NewSearchRequest(
		"DC=upbplanner,DC=local",
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(sAMAccountName=%s)", user),
		[]string{"memberOf", "displayName"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return "", nil, err
	}

	if len(sr.Entries) == 0 {
		return "", nil, errors.New("usuario no encontrado en LDAP")
	}

	entry := sr.Entries[0]

	var roles []string
	for _, groupDN := range entry.GetAttributeValues("memberOf") {
		dn, err := ldap.ParseDN(groupDN)
		if err == nil && len(dn.RDNs) > 0 {
			cn := dn.RDNs[0].Attributes[0].Value
			roles = append(roles, cn)
		}
	}

	u := &User{
		Username: user,
		Roles:    roles,
	}

	token, err := j.Generate(u)
	if err != nil {
		return "", nil, err
	}

	return token, u, nil
}

func createUser(c *gin.Context) {
	var req UserAuth

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}

	err := CreateLDAPUser(
		os.Getenv("ADMIN_LDAP_ADMIN"),
		os.Getenv("ADMIN_LDAP_PASS"),
		req.User,
		req.Pass,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	insertarLog(0,"CREAR_USUARIO","Se creó el usuario: "+req.User,)

	c.JSON(200, gin.H{"message": "Usuario creado correctamente"})
}

func CreateLDAPUser(adminUser, adminPass, username, password string) error {
	l, err := dialLDAPS()
	if err != nil {
		return err
	}
	defer l.Close()

	err = l.Bind(adminUser+"@upbplanner.local", adminPass)
	if err != nil {
		return err
	}

	userDN := fmt.Sprintf("CN=%s,CN=Usuarios,DC=upbplanner,DC=local", username)

	addReq := ldap.NewAddRequest(userDN, nil)

	addReq.Attribute("objectClass", []string{
		"top",
		"person",
		"organizationalPerson",
		"user",
	})

	addReq.Attribute("cn", []string{username})
	addReq.Attribute("sAMAccountName", []string{username})
	addReq.Attribute("userPrincipalName", []string{username + "@upbplanner.local"})
	addReq.Attribute("displayName", []string{username})
	addReq.Attribute("userAccountControl", []string{"544"})

	err = l.Add(addReq)
	if err != nil {
		return err
	}

	quotedPwd := fmt.Sprintf("\"%s\"", password)
	utf16Pwd := utf16.Encode([]rune(quotedPwd))

	pwdBytes := make([]byte, len(utf16Pwd)*2)
	for i, v := range utf16Pwd {
		binary.LittleEndian.PutUint16(pwdBytes[i*2:], v)
	}

	modPwd := ldap.NewModifyRequest(userDN, nil)
	modPwd.Replace("unicodePwd", []string{string(pwdBytes)})

	err = l.Modify(modPwd)
	if err != nil {
		return fmt.Errorf("error seteando password: %v", err)
	}

	modEnable := ldap.NewModifyRequest(userDN, nil)
	modEnable.Replace("userAccountControl", []string{"512"})

	err = l.Modify(modEnable)
	if err != nil {
		return fmt.Errorf("error habilitando usuario: %v", err)
	}

	groupDN := "CN=Usuarios,CN=Users,DC=upbplanner,DC=local"

	modGroup := ldap.NewModifyRequest(groupDN, nil)
	modGroup.Add("member", []string{userDN})

	err = l.Modify(modGroup)
	if err != nil {
		return fmt.Errorf("error agregando al grupo Usuario: %v", err)
	}

	return nil
}

func createAdmin(c *gin.Context) {
	var req UserAuth

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}

	err := CreateLDAPAdminUser(
		os.Getenv("ADMIN_LDAP_ADMIN"),
		os.Getenv("ADMIN_LDAP_PASS"),
		req.User,
		req.Pass,
	)
	insertarLog(0,"CREAR_ADMIN","Se creó un administrador: "+req.User,)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Admin creado correctamente"})
}

func CreateLDAPAdminUser(adminUser, adminPass, username, password string) error {
	l, err := dialLDAPS()
	if err != nil {
		return err
	}
	defer l.Close()

	err = l.Bind(adminUser+"@upbplanner.local", adminPass)
	if err != nil {
		return err
	}

	userDN := fmt.Sprintf("CN=%s,CN=Users,DC=upbplanner,DC=local", username)

	addReq := ldap.NewAddRequest(userDN, nil)

	addReq.Attribute("objectClass", []string{
		"top",
		"person",
		"organizationalPerson",
		"user",
	})

	addReq.Attribute("cn", []string{username})
	addReq.Attribute("sAMAccountName", []string{username})
	addReq.Attribute("userPrincipalName", []string{username + "@upbplanner.local"})
	addReq.Attribute("displayName", []string{username})
	addReq.Attribute("userAccountControl", []string{"544"})

	err = l.Add(addReq)
	if err != nil {
		return err
	}

	quotedPwd := fmt.Sprintf("\"%s\"", password)
	utf16Pwd := utf16.Encode([]rune(quotedPwd))

	pwdBytes := make([]byte, len(utf16Pwd)*2)
	for i, v := range utf16Pwd {
		binary.LittleEndian.PutUint16(pwdBytes[i*2:], v)
	}

	modPwd := ldap.NewModifyRequest(userDN, nil)
	modPwd.Replace("unicodePwd", []string{string(pwdBytes)})

	err = l.Modify(modPwd)
	if err != nil {
		return fmt.Errorf("error seteando password: %v", err)
	}

	modEnable := ldap.NewModifyRequest(userDN, nil)
	modEnable.Replace("userAccountControl", []string{"512"})

	err = l.Modify(modEnable)
	if err != nil {
		return fmt.Errorf("error habilitando usuario: %v", err)
	}

	groupDN :="CN=admin_upb_planner,CN=Users,DC=upbplanner,DC=local"

	modGroup := ldap.NewModifyRequest(groupDN, nil)
	modGroup.Add("member", []string{userDN})

	err = l.Modify(modGroup)
	if err != nil {
		return fmt.Errorf("error agregando al grupo admin_upb_planner: %v", err)
	}

	return nil
}

func changeusrpasswd(c *gin.Context) {
	var req UserAuth

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	err := ChangeUserPassword(
		os.Getenv("ADMIN_LDAP_ADMIN"),
		os.Getenv("ADMIN_LDAP_PASS"),
		req.User,
		req.Pass,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	var userID int
	err = db.QueryRow("SELECT N_idUsuario FROM Usuarios WHERE T_codUsuario = ?",req.User,).Scan(&userID)

	if err != nil {log.Println("Error obteniendo usuario para log:", err)
		userID = 0
	}
	insertarLog(userID,"CAMBIAR_CONTRASEÑA","El usuario cambió su contraseña",)
	c.JSON(200, gin.H{"message": "Contraseña cambiada correctamente"})
}

func ChangeUserPassword(adminUser, adminPass, username, newPassword string) error {
	l, err := dialLDAPS()
	if err != nil {
		return err
	}
	defer l.Close()

	err = l.Bind(adminUser+"@upbplanner.local", adminPass)
	if err != nil {
		return err
	}

	userDN := fmt.Sprintf("CN=%s,CN=Users,DC=upbplanner,DC=local", username)
	quotedPwd := fmt.Sprintf("\"%s\"", newPassword)
	utf16Pwd := utf16.Encode([]rune(quotedPwd))
	pwdBytes := make([]byte, len(utf16Pwd)*2)
	for i, v := range utf16Pwd {
		binary.LittleEndian.PutUint16(pwdBytes[i*2:], v)
	}

	modPwd := ldap.NewModifyRequest(userDN, nil)
	modPwd.Replace("unicodePwd", []string{string(pwdBytes)})

	err = l.Modify(modPwd)
	if err != nil {
		return fmt.Errorf("error cambiando password: %v", err)
	}
	return nil
}
