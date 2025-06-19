# User Risk Management System

A microservices-based system for user management with automated risk detection and multi-channel notifications, built with Go, gRPC, and RabbitMQ.

## Architecture

**Event-driven microservices** with dual communication patterns:
- **Synchronous**: gRPC for real-time service communication
- **Asynchronous**: RabbitMQ for event processing and notifications

```
┌─────────────────────────────────────────────────────────────────┐
│                     User Risk Management System                 │
└─────────────────────────────────────────────────────────────────┘

    ┌─────────────────┐
    │   API Gateway   │ ← HTTP/REST (Port 8080)
    │    (main.go)    │
    └─────────┬───────┘
              │ gRPC
    ┌─────────┴───────────────────────────────────────────────────┐
    │                                                             │
    v                           v                           v     v
┌─────────┐               ┌─────────┐               ┌─────────────┐
│  User   │               │  Risk   │               │Notification │
│Service  │◄──────────────┤ Engine  │──────────────►│   Service   │
│(50051)  │   Risk Check  │(50052)  │  Risk Events │   (50053)   │
└────┬────┘               └─────────┘               └─────┬───────┘
     │                                                    │
     │ PostgreSQL                            RabbitMQ ◄───┘
     v                                           │
┌─────────┐                                     │
│Database │                         ┌───────────v──────────┐
│(5432)   │                         │   External APIs      │
└─────────┘                         │ ┌─────────────────┐  │
                                    │ │ SendGrid(Email) │  │
                                    │ │ Twilio (SMS)    │  │
                                    │ └─────────────────┘  │
                                    └──────────────────────┘

Project Structure:
├── api-gateway/          # HTTP REST API & routing
├── cmd/                  # Service executables
│   ├── user/             # User management service
│   ├── risk-engine/      # Risk detection & analytics
│   └── notification/     # Multi-channel notifications
├── pkg/                  # Shared libraries
│   ├── auth/             # JWT authentication
│   ├── proto/            # gRPC definitions
│   └── messaging/        # RabbitMQ client
└── scripts/              # Database initialization
```

### Services
- **API Gateway** - REST endpoints, authentication, request routing
- **User Service** - User management, workflow orchestration, PostgreSQL
- **Risk Engine** - Fraud detection, configurable risk rules, analytics
- **Notification Service** - Multi-channel notifications (email, SMS, push)

## Quick Start

**Prerequisites**: Docker, Docker Compose, Go 1.21+

```bash
./setup.sh
make run
```

**Access the API:**
- **API Documentation**: http://localhost:8080/api/docs (Swagger UI)
- **OpenAPI Spec**: http://localhost:8080/api/docs/openapi.json
- **Health Check**: http://localhost:8080/api/v1/health

**Create a user:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "johnny_not_fake_at_all@fakeasfake.com", "first_name": "Johnny", "last_name": "Fake", "password": "HereIsMyFakePassword"}'
```

## Development

| Command | Description |
|---------|-------------|
| `make proto` | Generate Protocol Buffer files |
| `make test-all-endpoints` | Run comprehensive tests |
| `make logs` | View service logs |
| `make status` | Check service status |

## Key Features

- **OpenAPI 3.0 Documentation** - Interactive Swagger UI with API documentation
- **Automated Risk Detection** - Real-time fraud screening with configurable rules
- **Multi-Channel Notifications** - Email, SMS, and push notification delivery
- **Event-Driven Architecture** - Asynchronous processing with RabbitMQ
- **JWT Authentication** - Role-based access control with Bearer token support

## API Documentation

The API is fully documented using **OpenAPI 3.0.3** specification with interactive Swagger UI:

- **Interactive Documentation**: http://localhost:8080/api/docs
- **OpenAPI Specification**: http://localhost:8080/api/docs/openapi.json

### Key Endpoints

**Authentication**
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login with JWT tokens
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `GET /api/v1/profile` - Get authenticated user profile

**User Management** (Role-based access)
- `GET /api/v1/users` - List users (Admin only)
- `POST /api/v1/users` - Create user (Admin only)
- `GET /api/v1/users/{id}` - Get user details
- `PUT /api/v1/users/{id}` - Update user

**Risk Assessment**
- `POST /api/v1/risk/check` - Perform risk assessment

**Risk Management** (Admin only)
- `GET /api/v1/risk/rules` - List risk rules
- `POST /api/v1/risk/rules` - Create risk rule
- `PUT /api/v1/risk/rules/{id}` - Update risk rule
- `DELETE /api/v1/risk/rules/{id}` - Delete risk rule

**System**
- `GET /api/v1/health` - Health check

## Services & Ports

| Service | Port | Technology |
|---------|------|------------|
| API Gateway | 8080 | HTTP/REST |
| User Service | 50051 | gRPC, PostgreSQL |
| Risk Engine | 50052 | gRPC |
| Notification Service | 50053 | gRPC, RabbitMQ |
| RabbitMQ Management | 15672 | Web UI (guest/guest) |
| PostgreSQL | 5432 | Database |

## Testing

```bash
# Run all tests
make test-all-endpoints

# Load sample data
make load-test-data

# Performance testing
make test-performance
```

## Technology Stack

- **Backend**: Go, gRPC, Protocol Buffers
- **API Documentation**: OpenAPI 3.0.3, Swagger UI
- **Database**: PostgreSQL with GORM
- **Messaging**: RabbitMQ
- **Authentication**: JWT tokens with Bearer authentication
- **Deployment**: Docker, Docker Compose
- **External APIs**: SendGrid (email), Twilio (SMS)
