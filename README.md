# Api-Go — Backend REST API

API REST desarrollada en **Go** con el framework [Gin](https://github.com/gin-gonic/gin) para el sistema de gestión de horarios estudiantiles. Se encarga de exponer los endpoints que consume el frontend, conectándose a una base de datos **MySQL** y autenticando usuarios mediante **LDAP** con tokens **JWT**.

---

## Tabla de contenidos

- [Tecnologías](#tecnologías)
- [Estructura del proyecto](#estructura-del-proyecto)
- [Variables de entorno](#variables-de-entorno)
- [Instalación y ejecución](#instalación-y-ejecución)
- [Docker](#docker)
- [CI/CD](#cicd)
- [Endpoints de la API](#endpoints-de-la-api)
  - [Actividades oficiales](#actividades-oficiales)
  - [Comentarios](#comentarios)
  - [Actividades personales](#actividades-personales)
  - [Etiquetas](#etiquetas)
  - [Recordatorios](#recordatorios)
  - [Notificaciones y correos](#notificaciones-y-correos)
  - [Importar horario](#importar-horario)
  - [Usuario](#usuario)
  - [Autenticación LDAP](#autenticación-ldap)
- [Autenticación](#autenticación)
- [Arquitectura del código](#arquitectura-del-código)

---

## Tecnologías

| Tecnología | Versión | Uso |
|---|---|---|
| Go | 1.25+ | Lenguaje principal |
| Gin | 1.11 | Framework HTTP |
| MySQL | — | Base de datos relacional |
| LDAP (Active Directory) | — | Proveedor de autenticación |
| JWT (HS256) | — | Tokens de sesión |
| Docker | Multi-stage | Contenedorización |
| GitHub Actions | — | CI/CD |

---

## Estructura del proyecto

```
Api-go/
├── main.go                  # Punto de entrada, configuración DB y rutas
├── models.go                # Structs y tipos de datos
├── middleware.go             # Middleware de autenticación por API Key
├── modulo_official.go        # Handlers — Actividades oficiales
├── modulo_comment.go         # Handlers — Comentarios de actividades
├── modulo_personal.go        # Handlers — Actividades personales y tipos de curso
├── modulo_tag.go             # Handlers — Etiquetas
├── modulo_reminder.go        # Handlers — Recordatorios (ToDo List)
├── modulo_notification.go    # Handlers — Notificaciones y correos
├── modulo_import.go          # Handlers — Importación de horario
├── modulo_user.go            # Handlers — Información del usuario
├── modulo_ldap.go            # Autenticación LDAP, JWT y gestión de usuarios
├── internal/
│   └── auth/
│       ├── service.go        # Servicio de autenticación (interfaz Provider)
│       └── types.go          # Tipos y errores del dominio auth
├── Dockerfile                # Build multi-stage para producción
├── go.mod                    # Dependencias del módulo
├── go.sum                    # Checksums de dependencias
├── .github/
│   └── workflows/
│       └── CI.yml            # Pipeline CI/CD → GitHub Container Registry
└── README.md
```

---

## Variables de entorno

Crea un archivo `.env` en la raíz del proyecto con las siguientes variables:

```env
# Base de datos MySQL
DB_USER=tu_usuario
DB_PASS=tu_contraseña
DB_ADDR=direccion_del_host
DB_ADDR_PORT=3306
DB_NAME=nombre_de_la_bd

# API Key para proteger los endpoints
API_KEY=tu_api_key_secreta

# LDAP / Active Directory
LDAP_ADDR=direccion_ldap
LDAP_PORT=389

# JWT
JWT_SECRET=tu_secreto_jwt

# Admin LDAP (para creación de usuarios)
ADMIN_LDAP_ADMIN=usuario_admin
ADMIN_LDAP_PASS=password_admin
```

> **Importante:** El archivo `.env` no debe subirse al repositorio.

---

## Instalación y ejecución

### Requisitos previos

- Go 1.25 o superior
- Acceso a una instancia MySQL
- (Opcional) Acceso a un servidor LDAP/Active Directory

### Pasos

```bash
# 1. Clonar el repositorio
git clone https://github.com/Lu2ky/Api-go.git
cd Api-go

# 2. Instalar dependencias
go mod tidy

# 3. Crear el archivo .env (ver sección Variables de entorno)

# 4. Ejecutar
go run .
```

La API estará disponible en `http://localhost:8080`.

---

## Docker

El proyecto usa un build **multi-stage** para generar una imagen ligera basada en Alpine.

```bash
# Construir la imagen
docker build -t api-go .

# Ejecutar el contenedor
docker run -p 8080:8080 --env-file .env api-go
```

### Dockerfile

```dockerfile
FROM golang:1.26.0-alpine3.23 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .

FROM alpine:latest
COPY --from=builder /app/main .
RUN chmod u+x main
RUN touch .env
EXPOSE 3913
CMD ["./main"]
```

> **Nota:** El contenedor expone el puerto `3913` en el Dockerfile pero la aplicación escucha en el puerto `8080`. Ajustar según sea necesario.

---

## CI/CD

El repositorio cuenta con un pipeline de **GitHub Actions** (`.github/workflows/CI.yml`) que se ejecuta en cada push a la rama `Testing`:

1. Hace checkout del código
2. Inicia sesión en **GitHub Container Registry** (`ghcr.io`)
3. Construye y publica la imagen Docker como `ghcr.io/lu2ky/apigohe:latest`

El runner es **self-hosted**.

---

## Endpoints de la API

Todos los endpoints requieren el header `X-API-Key` con una API Key válida.

### Actividades oficiales

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/GetOfficialScheduleByUserId/:id` | Obtener horario oficial por código de usuario |

### Comentarios

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/GetPersonalComments/:id` | Obtener todos los comentarios del usuario |
| `GET` | `/GetPersonalCourseComments/:id/:idCourse` | Comentarios por usuario y curso |
| `POST` | `/addPersonalComment` | Agregar un comentario |
| `POST` | `/updatePersonalComment` | Editar un comentario |
| `POST` | `/deletePersonalComment` | Eliminar un comentario |

### Actividades personales

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/GetPersonalScheduleByUserId/:id` | Obtener actividades personales del usuario |
| `POST` | `/addPersonalActivity` | Crear una actividad personal |
| `POST` | `/updatePersonalScheduleByIdCourse` | Editar una actividad personal |
| `POST` | `/deleteOrRecoveryPersonalScheduleByIdCourse` | Eliminar/recuperar actividad personal |
| `GET` | `/GetTiposCurso` | Obtener tipos de curso disponibles |

### Etiquetas

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/GetTagsByUserId/:id` | Obtener etiquetas del usuario |
| `GET` | `/GetTagsByUserIdAndReminderId/:id/:reminderId` | Etiquetas por usuario y recordatorio |
| `POST` | `/deleteTag` | Eliminar una etiqueta |

### Recordatorios

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/GetReminders/:id` | Obtener recordatorios del usuario |
| `GET` | `/GetRemindersTags/:id` | Recordatorios con etiquetas asociadas |
| `POST` | `/addReminder` | Crear recordatorio (con hasta 5 etiquetas) |
| `POST` | `/updateReminder` | Editar recordatorio |
| `POST` | `/deleteOrRecoverReminder` | Eliminar/recuperar recordatorio |

### Notificaciones y correos

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/GetNotifications/:id` | Obtener notificaciones del usuario |
| `POST` | `/addNotification` | Crear una notificación |
| `POST` | `/muteNotification` | Configurar/silenciar notificaciones |
| `POST` | `/addCorreo` | Crear un correo |

### Importar horario

| Método | Ruta | Descripción |
|---|---|---|
| `POST` | `/importSchedule` | Importar un horario oficial desde datos externos |

### Usuario

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/GetUserInfo/:id` | Obtener información del perfil del usuario |

### Autenticación LDAP

| Método | Ruta | Descripción |
|---|---|---|
| `POST` | `/auth` | Autenticarse con credenciales LDAP (devuelve JWT) |
| `POST` | `/addauthuser` | Crear usuario en LDAP |
| `POST` | `/addadmin` | Crear usuario administrador en LDAP |

---

## Autenticación

La API implementa dos capas de seguridad:

### 1. API Key (middleware global)
Todas las peticiones deben incluir el header:
```
X-API-Key: tu_api_key
```
Si la key no es válida o no se envía, la petición es rechazada con `401`/`403`.

### 2. LDAP + JWT (autenticación de usuarios)
El endpoint `/auth` recibe usuario y contraseña, los valida contra un servidor **Active Directory** (LDAP), y retorna:
- Un **token JWT** (HS256) con duración de 24 horas
- La información del usuario con sus roles/grupos

**Payload del JWT:**
```json
{
  "sub": "codigo_usuario",
  "roles": ["Usuario", "OtroGrupo"],
  "iss": "horario_estudiantes",
  "exp": 1741305600
}
```

---

## Arquitectura del código

```
Petición HTTP
    │
    ▼
┌──────────────┐
│  middleware   │  ← Validación de API Key
└──────┬───────┘
       │
       ▼
┌──────────────┐
│    main.go   │  ← Router (Gin) + Configuración DB
└──────┬───────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│              Módulos (handlers)               │
├──────────────┬──────────────┬────────────────┤
│  official    │  personal    │  reminder      │
│  comment     │  tag         │  notification  │
│  import      │  user        │  ldap          │
└──────┬───────┴──────┬───────┴────────┬───────┘
       │              │                │
       ▼              ▼                ▼
┌──────────────┐  ┌────────┐  ┌──────────────┐
│    MySQL     │  │ models │  │  LDAP / JWT  │
│  (database)  │  │ (.go)  │  │  (auth)      │
└──────────────┘  └────────┘  └──────────────┘
```

Cada módulo (`modulo_*.go`) agrupa las funciones handler de un dominio específico. Todos comparten la conexión global `db` y los tipos definidos en `models.go`.