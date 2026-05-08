# Api-Go — Backend REST API

API REST desarrollada en **Go** con el framework [Gin](https://github.com/gin-gonic/gin) para el sistema de gestión de horarios estudiantiles. Se encarga de exponer los endpoints que consume el frontend, conectándose a una base de datos **MySQL**, cacheando datos en **Redis**, y autenticando usuarios mediante **LDAP** con tokens **JWT**.

---

## Tabla de contenidos

- [Tecnologías](#tecnologías)
- [Estructura del proyecto](#estructura-del-proyecto)
- [Instalación y configuración](#instalación-y-configuración)
  - [Variables de entorno](#variables-de-entorno)
  - [Requisitos](#requisitos)
- [Ejecución](#ejecución)
  - [Desarrollo local](#desarrollo-local)
  - [Docker](#docker)
  - [CI/CD](#cicd)
- [Endpoints de la API](#endpoints-de-la-api)
  - [Autenticación](#autenticación)
  - [Horarios oficiales](#horarios-oficiales)
  - [Horarios personales](#horarios-personales)
  - [Comentarios](#comentarios)
  - [Etiquetas](#etiquetas)
  - [Recordatorios (ToDo)](#recordatorios-todo)
  - [Notificaciones](#notificaciones)
  - [Importación de horarios](#importación-de-horarios)
  - [Períodos académicos](#períodos-académicos)
  - [Usuarios](#usuarios)
  - [Registros (Logs)](#registros-logs)
  - [Paleta de colores](#paleta-de-colores)
  - [Onboarding](#onboarding)
- [Autenticación y autorización](#autenticación-y-autorización)
  - [JWT](#jwt)
  - [Middleware](#middleware)
  - [Roles](#roles)
- [Arquitectura del código](#arquitectura-del-código)

---

## Tecnologías

| Tecnología | Versión | Uso |
|---|---|---|
| Go | 1.25+ | Lenguaje principal |
| Gin | 1.12.0 | Framework HTTP REST |
| MySQL | 5.7+ | Base de datos relacional |
| Redis | 7.0+ | Cache distribuido |
| LDAP (Active Directory) | — | Proveedor de autenticación |
| JWT (HS256) | — | Tokens de sesión seguros |
| Docker | Multi-stage | Contenedorización |
| GitHub Actions | — | CI/CD automatizado |

---

## Estructura del proyecto

```
Api-go/
├── main.go                      # Punto de entrada, configuración e inicialización de rutas
├── middleware.go                # Middleware de autenticación por API Key
├── models.go                    # Tipos/structs de datos (requests/responses)
│
├── internal/
│   └── auth/
│       ├── service.go          # Interfaz de servicio de autenticación (Provider pattern)
│       └── types.go            # Tipos de dominio para autenticación
│
├── modulo_ldap.go              # Autenticación LDAP, JWT, gestión de usuarios
├── modulo_logs.go              # Sistema de auditoria y logs
│
├── Handlers (módulos de negocio):
│   ├── modulo_official.go       # Horarios académicos oficiales
│   ├── modulo_personal.go       # Actividades personales y tipos de curso
│   ├── modulo_comment.go        # Comentarios sobre actividades
│   ├── modulo_tag.go            # Etiquetas para recordatorios
│   ├── modulo_reminder.go       # Recordatorios (ToDo List)
│   ├── modulo_notification.go   # Notificaciones y correos
│   ├── modulo_import.go         # Importación de horarios desde sistemas externos
│   └── modulo_user.go           # Información del usuario
│
├── Dockerfile                   # Build multi-stage para producción
├── go.mod                       # Definición de módulo y dependencias
├── go.sum                       # Checksums de dependencias (para reproducibilidad)
├── .github/
│   └── workflows/
│       └── CI.yml              # Pipeline CI/CD → GitHub Container Registry
└── README.md
```

**Patrón de organización**: Cada `modulo_*.go` contiene los handlers HTTP y lógica de negocio para su dominio específico. Los handlers internos de Gin están nombrados en minúsculas (ej: `getOfficialScheduleByUserId`).

---

## Instalación y configuración

### Requisitos

- Go 1.25+
- MySQL 5.7+
- Redis 7.0+
- LDAP/Active Directory disponible (para autenticación)

### Variables de entorno

Crea un archivo `.env` en la raíz del proyecto con las siguientes variables:

```env
# Base de datos MySQL
DB_USER=tu_usuario
DB_PASS=tu_contraseña
DB_ADDR=localhost
DB_ADDR_PORT=3306
DB_NAME=nombre_de_la_bd

# Redis (Cache)
DB_ADDR_REDIS=localhost
DB_ADDR_PORT_REDIS=6379
DB_PASS_REDIS=contraseña_redis

# API Key para proteger los endpoints
API_KEY=tu_api_key_secreta_fuerte

# LDAP / Active Directory
LDAP_ADDR=ldap.tudominio.com
LDAP_PORT=389

# JWT (Tokens de sesión)
JWT_SECRET=tu_secreto_jwt_muy_seguro
JWT_TTL=24h
JWT_ISSUER=api-go

# Admin LDAP (para creación de usuarios)
ADMIN_LDAP_ADMIN=usuario_admin_ldap
ADMIN_LDAP_PASS=password_admin

# Roles (para autorización)
ROLE_ADM=admin
ROLE_USER=user
```

---

## Ejecución

### Desarrollo local

```bash
# Descargar dependencias
go mod download

# Ejecutar la aplicación
go run main.go modulo_*.go middleware.go

# O más simple:
go run *.go
```

La API estará disponible en `http://localhost:8080`

### Docker

```bash
# Build de la imagen (multi-stage, optimizado para producción)
docker build -t api-go:latest .

# Ejecutar el contenedor
docker run -d \
  --name api-go \
  -p 8080:8080 \
  --env-file .env \
  api-go:latest
```

### CI/CD

El proyecto usa **GitHub Actions** (ver `.github/workflows/CI.yml`):

1. Al hacer push a `main`, se ejecutan tests y builds
2. Se construye automáticamente una imagen Docker
3. Se sube a **GitHub Container Registry** (ghcr.io)

---

## Endpoints de la API

**Base URL**: `http://localhost:8080/api/v1`

**Nota**: Todos los endpoints requieren el header `X-API-Key: tu_api_key_secreta`

Los endpoints protegidos requieren además un header `Authorization: Bearer <jwt_token>`

---

### Autenticación

#### Login (obtener JWT token)
```
POST /auth/login
Content-Type: application/json

{
  "user": "codigo_estudiante",
  "pass": "contraseña_ldap"
}

Response 200:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "userID": "codigo_estudiante",
  "name": "Nombre Completo",
  "roles": ["user"]
}
```

#### Registrar usuario
```
POST /auth/users
Content-Type: application/json

{
  "user": "codigo_nuevo",
  "pass": "contraseña"
}
```

#### Cambiar contraseña
```
POST /auth/change-password
Content-Type: application/json
Authorization: Bearer <token>

{
  "user": "codigo_usuario",
  "oldPass": "contraseña_anterior",
  "newPass": "contraseña_nueva"
}
```

#### Crear admin (solo admins)
```
POST /auth/admins
Content-Type: application/json
Authorization: Bearer <admin_token>

{
  "user": "codigo_admin",
  "pass": "contraseña"
}
```

---

### Horarios oficiales

#### Obtener horario oficial del usuario
```
GET /schedules/official/users/:id
Authorization: Bearer <token>

Response 200:
[
  {
    "N_idCurso": 1,
    "P_nombreCurso": "Programación I",
    "P_descripcion": "Introducción a la programación",
    "P_dia": 1,
    "P_horaInicio": "09:00",
    "P_horaFin": "11:00",
    "P_aula": "A101"
  }
]
```

#### Obtener tipos de curso
```
GET /course-types
Authorization: Bearer <token>

Response 200:
[
  {
    "id": 1,
    "nombre": "Teoría",
    "abreviacion": "TEO"
  }
]
```

#### Verificar colisiones de horarios
```
POST /schedules/activities/times
Content-Type: application/json
Authorization: Bearer <token>

{
  "idUsuario": 123,
  "dia": 1,
  "actividades": [...]
}
```

#### Obtener períodos académicos
```
GET /academic-periods
Authorization: Bearer <token>

Response 200:
[
  {
    "id": 1,
    "nombre": "Semestre 2025-1",
    "fechaInicio": "2025-02-01",
    "fechaFin": "2025-06-30"
  }
]
```

#### Crear período académico (solo admins)
```
POST /academic-periods/insert
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "nombre": "Semestre 2025-2",
  "fechaInicio": "2025-08-01",
  "fechaFin": "2025-12-31"
}
```

---

### Horarios personales

#### Obtener horario personal
```
GET /schedules/personal/users/:id
Authorization: Bearer <token>

Response 200:
[
  {
    "P_idCurso": 1,
    "P_nombreCurso": "Estudio adicional",
    "P_descripcion": "Preparación para examen",
    "P_dia": 1,
    "P_horaInicio": "14:00",
    "P_horaFin": "15:30"
  }
]
```

#### Crear actividad personal
```
POST /schedules/personal
Authorization: Bearer <token>
Content-Type: application/json

{
  "P_usuario": 123,
  "P_nombreCurso": "Estudio",
  "P_descripcion": "Estudio de matemáticas",
  "P_fechaInicio": "2025-02-01",
  "P_fechaFin": "2025-02-28",
  "P_dia": 2,
  "P_horaInicio": "15:00",
  "P_horaFin": "16:30",
  "codUsuario": "codigo_usuario"
}
```

#### Actualizar actividad personal
```
POST /schedules/personal/update
Authorization: Bearer <token>
Content-Type: application/json

{
  "P_idCurso": 1,
  "P_nombreCurso": "Estudio actualizado",
  "P_descripcion": "Nueva descripción",
  ...
}
```

#### Eliminar o recuperar actividad
```
POST /schedules/personal/delete-or-recover
Authorization: Bearer <token>
Content-Type: application/json

{
  "IdPersonalSchedule": 1,
  "codUsuario": "codigo_usuario"
}
```

---

### Comentarios

#### Obtener comentarios personales
```
GET /comments/personal/users/:id
Authorization: Bearer <token>

Response 200:
[
  {
    "N_idComentario": 1,
    "N_idUsuario": 123,
    "N_idCurso": 5,
    "P_comentario": "Buen curso",
    "Dt_fecha": "2025-02-15T10:30:00Z"
  }
]
```

#### Obtener comentarios por curso
```
GET /comments/personal/users/:id/courses/:idCourse
Authorization: Bearer <token>
```

#### Crear comentario
```
POST /comments/personal
Authorization: Bearer <token>
Content-Type: application/json

{
  "N_idUsuario": 123,
  "N_idCurso": 5,
  "P_comentario": "Excelente contenido",
  "codUsuario": "codigo_usuario"
}
```

#### Actualizar comentario
```
POST /comments/personal/update
Authorization: Bearer <token>
Content-Type: application/json

{
  "N_idComentario": 1,
  "P_comentario": "Comentario actualizado",
  "codUsuario": "codigo_usuario"
}
```

#### Eliminar comentario
```
POST /comments/personal/delete
Authorization: Bearer <token>
Content-Type: application/json

{
  "N_idComentario": 1,
  "codUsuario": "codigo_usuario"
}
```

---

### Etiquetas

#### Obtener etiquetas del usuario
```
GET /tags/users/:id
Authorization: Bearer <token>

Response 200:
[
  {
    "N_idEtiqueta": 1,
    "P_etiqueta": "urgente",
    "P_color": "#FF0000"
  }
]
```

#### Obtener etiquetas de recordatorio específico
```
GET /tags/users/:id/reminders/:reminderId
Authorization: Bearer <token>
```

#### Eliminar etiqueta
```
POST /tags/delete
Authorization: Bearer <token>
Content-Type: application/json

{
  "N_idEtiqueta": 1,
  "codUsuario": "codigo_usuario"
}
```

---

### Recordatorios (ToDo)

#### Obtener recordatorios del usuario
```
GET /reminders/users/:id
Authorization: Bearer <token>

Response 200:
[
  {
    "N_idRecordatorio": 1,
    "N_idUsuario": 123,
    "P_descripcion": "Estudiar capítulo 5",
    "P_completado": false,
    "Dt_fecha": "2025-02-20T18:00:00Z",
    "etiquetas": [1, 2]
  }
]
```

#### Obtener recordatorios con etiquetas
```
GET /reminders/users/:id/tags
Authorization: Bearer <token>
```

#### Crear recordatorio
```
POST /reminders
Authorization: Bearer <token>
Content-Type: application/json

{
  "N_idUsuario": 123,
  "P_descripcion": "Entregar proyecto",
  "Dt_fecha": "2025-02-28T23:59:59Z",
  "etiquetas": [1],
  "codUsuario": "codigo_usuario"
}
```

#### Actualizar recordatorio
```
POST /reminders/update
Authorization: Bearer <token>
Content-Type: application/json

{
  "N_idRecordatorio": 1,
  "P_descripcion": "Nuevo texto",
  "P_completado": true
}
```

#### Eliminar o recuperar recordatorio
```
POST /reminders/delete-or-recover
Authorization: Bearer <token>
Content-Type: application/json

{
  "N_idRecordatorio": 1,
  "codUsuario": "codigo_usuario"
}
```

#### Eliminar múltiples recordatorios
```
POST /reminders/delete/multiple
Authorization: Bearer <token>
Content-Type: application/json

{
  "recordatorios": [1, 2, 3],
  "codUsuario": "codigo_usuario"
}
```

---

### Notificaciones

#### Obtener notificaciones
```
GET /notifications/users/:id
Authorization: Bearer <token>

Response 200:
[
  {
    "N_idNotificacion": 1,
    "P_titulo": "Nuevo horario disponible",
    "P_descripcion": "Tu horario está listo",
    "P_leida": false,
    "Dt_fecha": "2025-02-15T09:00:00Z"
  }
]
```

#### Silenciar notificaciones
```
POST /notifications/mute
Authorization: Bearer <token>
Content-Type: application/json

{
  "N_idNotificacion": 1,
  "codUsuario": "codigo_usuario"
}
```

#### Crear notificación (interno/admin)
```
POST /notifications
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "N_idUsuario": 123,
  "P_titulo": "Notificación",
  "P_descripcion": "Contenido"
}
```

#### Eliminar notificaciones (interno/admin)
```
POST /notifications/delete
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "N_idNotificacion": 1
}
```

---

### Importación de horarios

#### Importar horario desde sistema externo
```
POST /schedules/import
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "archivo": "base64_contenido",
  "tipo": "excel",
  "periodo": 1
}

Response 201:
{
  "registros_importados": 150,
  "errores": 2,
  "detalles": [...]
}
```

---

### Usuarios

#### Obtener información del usuario
```
GET /users/:id
Authorization: Bearer <token>

Response 200:
{
  "N_idUsuario": 123,
  "T_codigo": "codigo_usuario",
  "T_nombre": "Nombre Completo",
  "T_email": "usuario@universidad.edu",
  "T_roles": ["user"]
}
```

---

### Registros (Logs)

#### Registrar evento (auditoria)
```
POST /logs
Authorization: Bearer <token>
Content-Type: application/json

{
  "codUsuario": "codigo_usuario",
  "accion": "VER_HORARIO",
  "descripcion": "Usuario consultó su horario oficial"
}
```

---

### Paleta de colores

#### Guardar paleta de colores del usuario
```
POST /palette
Authorization: Bearer <token>
Content-Type: application/json

{
  "usuario": 123,
  "paleta": {
    "primario": "#FF5733",
    "secundario": "#33FF57",
    "acento": "#3357FF"
  }
}
```

#### Obtener paleta de colores
```
POST /palette/get
Authorization: Bearer <token>
Content-Type: application/json

{
  "usuario": 123
}
```

---

### Onboarding

#### Registrar estado de onboarding
```
POST /onboarding
Authorization: Bearer <token>
Content-Type: application/json

{
  "usuario": 123,
  "paso": 2,
  "completado": false
}
```

#### Obtener estado de onboarding
```
POST /onboarding/get
Authorization: Bearer <token>
Content-Type: application/json

{
  "usuario": 123
}
```

---

## Autenticación y autorización

### JWT

Los tokens JWT se usan para mantener sesiones seguras. Se generan en el login y contienen:
- **sub** (subject): ID del usuario
- **name**: Nombre del usuario
- **roles**: Array de roles del usuario
- **exp**: Tiempo de expiración
- **iat**: Tiempo de emisión

**TTL por defecto**: 24 horas (configurable en `.env`)

### Middleware

#### `apiKeyAuth()`
Valida que todas las peticiones contengan el header `X-API-Key` correcto.

#### `AuthMiddleware()` (JWT)
Valida el token JWT en peticiones a `/api/v1/*`. El token se envía en el header `Authorization: Bearer <token>`

#### `UserGetMiddleware()`
Verifica que el usuario en la URL sea el usuario autenticado (previene acceso a datos de otros usuarios).

#### `RoleMiddleware(role string)`
Verifica que el usuario tenga el rol requerido para ejecutar la acción.

### Roles

- **user**: Usuario normal (estudiante)
- **admin**: Administrador del sistema

---

## Arquitectura del código

### Patrón usado

El proyecto sigue el patrón de **handlers por módulo**:

1. **main.go**: Inicialización, configuración e inyección de dependencias
2. **middleware.go**: Middleware compartido de autenticación
3. **models.go**: Tipos de datos (DTOs, requests, responses)
4. **modulo_*.go**: Lógica de negocio y handlers HTTP para cada dominio

### Flujo de una petición

1. Cliente envía petición HTTP con API Key
2. `apiKeyAuth()` valida la API Key
3. Si requiere JWT, `AuthMiddleware()` valida el token
4. Se aplican middlewares adicionales si es necesario (UserGetMiddleware, RoleMiddleware)
5. Se ejecuta el handler específico
6. El handler consulta la BD MySQL (con caché en Redis si aplica)
7. Se registra la acción en la tabla de Logs
8. Se retorna la respuesta

### Guía para agregar nuevos endpoints

1. Definir structs de request/response en `models.go`
2. Crear el handler (ej: `func myNewHandler(c *gin.Context) {}`) en `modulo_*.go`
3. Registrar la ruta en `registerV1Routes()` en `main.go`
4. Agregar middleware si es necesario (JWT, Role, UserGet)
5. Documentar en este README

### Convenciones

- **Rutas**: CamelCase en minúsculas (ej: `/schedules/official`)
- **Handlers HTTP**: minúsculas (ej: `getOfficialScheduleByUserId`)
- **Handlers exportados (usados desde main)**: MAYÚSCULA inicial (ej: `GetOfficialScheduleByUserId`)
- **Variables globales**: lowercase (ej: `db`, `rdb`, `ctx`)
- **Structs**: PascalCase (ej: `Claims`, `User`)
- **Campos JSON**: con tags (ej: `json:"id"`)

---

## Notas de mantenimiento

### Limpieza realizada (Versión actual)

✅ Rutas legacy eliminadas (33 rutas duplicadas removidas)
✅ Función vacía `method()` eliminada
✅ Struct `PersonalScheduleNewValue` sin usar eliminada
✅ Código reorganizado y centralizado en API v1
✅ README actualizado con documentación completa

### Próximas mejoras sugeridas

- [ ] Usar `internal/auth/` Service para refactorizar autenticación
- [ ] Migrar a structured logging (stdlib log/slog o slog)
- [ ] Agregar tests unitarios
- [ ] Agregar documentación OpenAPI/Swagger
- [ ] Implementar rate limiting
- [ ] Agregar observabilidad (traces, métricas)