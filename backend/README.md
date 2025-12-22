# Ecosistema Imob - Backend API

Multi-tenant real estate management system backend built with Go, Gin, and Firebase.

## Architecture

- **Framework**: Gin Web Framework
- **Database**: Cloud Firestore
- **Authentication**: Firebase Authentication
- **Storage**: Google Cloud Storage
- **Language**: Go 1.25+

## Project Structure

```
backend/
├── cmd/
│   └── server/          # Application entry point
│       └── main.go      # Main server initialization
├── internal/
│   ├── config/          # Configuration management
│   ├── middleware/      # HTTP middleware
│   │   ├── auth.go      # Firebase authentication
│   │   ├── cors.go      # CORS handling
│   │   ├── logging.go   # Request logging
│   │   ├── error.go     # Error recovery
│   │   └── tenant.go    # Tenant validation
│   ├── models/          # Domain models
│   ├── repositories/    # Data access layer
│   ├── services/        # Business logic layer
│   ├── handlers/        # HTTP handlers
│   └── utils/           # Utility functions
├── config/              # Configuration files
│   └── firebase-adminsdk.json  # Firebase credentials
├── .env                 # Environment variables
├── go.mod               # Go module definition
├── go.sum               # Go dependencies
└── Makefile             # Build automation

```

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Firebase project with Firestore enabled
- Firebase Admin SDK credentials

### Installation

1. Clone the repository
2. Navigate to the backend directory
3. Install dependencies:

```bash
make install
```

### Configuration

1. Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

2. Update `.env` with your configuration:

```env
# Firebase Configuration
GOOGLE_APPLICATION_CREDENTIALS=./config/firebase-adminsdk.json
FIREBASE_PROJECT_ID=your-project-id

# Server Configuration
PORT=8080
GIN_MODE=debug
ENVIRONMENT=development

# CORS Configuration
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,http://localhost:3002

# Cloud Storage
GCS_BUCKET_NAME=your-bucket-name

# Logging
LOG_LEVEL=info
```

3. Place your Firebase Admin SDK credentials at `config/firebase-adminsdk.json`

### Running the Server

#### Development Mode (with auto-reload)

```bash
make dev
```

#### Production Mode

```bash
make build
make run
```

Or directly:

```bash
go run ./cmd/server/main.go
```

### Building

Build the server binary:

```bash
make build
```

The binary will be created at `bin/server.exe`

## API Routes

### Public Endpoints

- `GET /health` - Health check
- `GET /metrics` - Basic metrics
- `POST /tenants` - Create new tenant (public registration)

### Protected Endpoints (Require Authentication)

All API routes are prefixed with `/api` and require authentication via Firebase ID token in the `Authorization` header:

```
Authorization: Bearer <firebase-id-token>
```

#### Tenant-Scoped Routes

All tenant-scoped routes require the `tenant_id` path parameter and validate tenant existence and active status.

- **Brokers**: `/api/:tenant_id/brokers`
- **Owners**: `/api/:tenant_id/owners`
- **Properties**: `/api/:tenant_id/properties`
- **Listings**: `/api/:tenant_id/listings`
- **Property Broker Roles**: `/api/:tenant_id/property-broker-roles`
- **Leads**: `/api/:tenant_id/leads`
- **Activity Logs**: `/api/:tenant_id/activity-logs`

Each resource supports standard CRUD operations (GET, POST, PUT, DELETE).

## Middleware

### Authentication Middleware

- **AuthRequired()** - Validates Firebase ID tokens and extracts user information
- **OptionalAuth()** - Optionally extracts user info if token is present (for public endpoints)

### Tenant Middleware

- **ValidateTenant()** - Validates tenant exists and is active, sets tenant in context

### CORS Middleware

- Configurable allowed origins, methods, and headers
- Supports wildcard subdomains
- Credentials support

### Logging Middleware

- Logs all HTTP requests with structured data
- Includes request ID for tracing
- Logs method, path, status code, latency, user info

### Error Recovery Middleware

- Recovers from panics
- Logs stack traces
- Returns proper error responses

## Configuration

Configuration is managed through environment variables and the `config` package:

- **Firebase**: Project ID, credentials path, bucket name
- **Server**: Port, host, environment, Gin mode
- **CORS**: Allowed origins
- **Logging**: Log level

## Development

### Project Layout

The project follows the standard Go project layout with clean architecture principles:

- **cmd/**: Application entry points
- **internal/**: Private application code
- **config/**: External configuration files

### Dependency Injection

The application uses constructor-based dependency injection:

1. Initialize Firebase (App, Auth, Firestore)
2. Initialize Repositories (with Firestore client)
3. Initialize Services (with Repositories)
4. Initialize Handlers (with Services)
5. Setup Router (with Handlers and Middleware)

### Adding New Routes

1. Create handler in `internal/handlers/`
2. Register routes in handler's `RegisterRoutes()` method
3. Add handler initialization in `main.go`
4. Register handler routes in router setup

## Testing

Run tests:

```bash
make test
```

## Building for Production

1. Set environment to production:
```env
ENVIRONMENT=production
GIN_MODE=release
```

2. Build the binary:
```bash
make build
```

3. Run the server:
```bash
./bin/server.exe
```

## Graceful Shutdown

The server supports graceful shutdown with a 30-second timeout. It will:

1. Stop accepting new requests
2. Wait for existing requests to complete (up to 30 seconds)
3. Close database connections
4. Exit cleanly

Trigger shutdown with `Ctrl+C` or `SIGTERM`.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `FIREBASE_PROJECT_ID` | Firebase project ID | Required |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to Firebase credentials | `./config/firebase-adminsdk.json` |
| `GCS_BUCKET_NAME` | Cloud Storage bucket name | Required |
| `PORT` | Server port | `8080` |
| `HOST` | Server host | `0.0.0.0` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |
| `ALLOWED_ORIGINS` | CORS allowed origins (comma-separated) | `http://localhost:3000,...` |
| `LOG_LEVEL` | Logging level | `info` |

## License

Copyright (c) 2024 Altatech Systems
