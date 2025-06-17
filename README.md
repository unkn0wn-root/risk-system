# User Risk Management System

A microservices-based system for user management with automated risk detection and multi-channel notifications, built with Go, gRPC, and RabbitMQ.

## Architecture

**Event-driven microservices** with dual communication patterns:
- **Synchronous**: gRPC for real-time service communication
- **Asynchronous**: RabbitMQ for event processing and notifications

### Services
- **API Gateway** - REST endpoints, authentication, request routing
- **User Service** - User management, workflow orchestration, PostgreSQL persistence
- **Risk Engine** - Fraud detection, configurable risk rules, analytics
- **Notification Service** - Multi-channel notifications (email, SMS, push)

## Quick Start

**Prerequisites**: Docker, Docker Compose, Go 1.21+

```bash
./setup.sh
make run
```

**Test the system:**
```bash
curl http://localhost:8080/api/v1/health
```

**Create a user:**
```bash
curl -X POST http://localhost:8080/api/v1/users \
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

- **Automated Risk Detection** - Real-time fraud screening with configurable rules
- **Multi-Channel Notifications** - Email, SMS, and push notification delivery
- **Event-Driven Architecture** - Asynchronous processing with RabbitMQ
- **JWT Authentication** - Role-based access control

## API Endpoints

- **Health**: `GET /api/v1/health`
- **User Management**: `POST /api/v1/users`, `GET /api/v1/users/{id}`
- **Authentication**: `POST /api/v1/auth/register`, `POST /api/v1/auth/login`

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
- **Database**: PostgreSQL with GORM
- **Messaging**: RabbitMQ
- **Authentication**: JWT tokens
- **Deployment**: Docker, Docker Compose
- **External APIs**: SendGrid (email), Twilio (SMS)
