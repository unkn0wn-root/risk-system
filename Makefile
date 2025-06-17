.PHONY: build run stop clean proto test deps fix-imports test-auth test-users test-admin test-errors test-all-endpoints

deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download

proto:
	@echo "🔧 Generating protobuf files..."
	@if ! command -v protoc >/dev/null 2>&1; then \
		echo "❌ protoc is not installed. Please install Protocol Buffers compiler."; \
		echo "On macOS: brew install protobuf"; \
		echo "On Ubuntu: sudo apt install protobuf-compiler"; \
		exit 1; \
	fi
	@if ! command -v protoc-gen-go >/dev/null 2>&1; then \
		echo "Installing protoc-gen-go..."; \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	fi
	@if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then \
		echo "Installing protoc-gen-go-grpc..."; \
		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
	fi
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/proto/user/*.proto \
		pkg/proto/risk/*.proto \
		pkg/proto/notification/*.proto

build-local:
	@echo "🔨 Building services locally..."
	go build -o bin/api-gateway ./api-gateway/
	go build -o bin/user-service ./cmd/user/
	go build -o bin/risk-engine ./cmd/risk-engine/
	go build -o bin/notification-service ./cmd/notification/

build:
	@echo "🐳 Building Docker images..."
	docker-compose build --parallel

run:
	@echo "🚀 Starting all services..."
	docker-compose up -d --build
	@echo ""
	@echo "✅ Services started successfully!"
	@echo "🌐 API Gateway: http://localhost:8080"
	@echo "🐰 RabbitMQ Management: http://localhost:15672 (guest/guest)"
	@echo "🐘 PostgreSQL: localhost:5432 (postgres/postgres)"
	@echo ""
	@echo "📋 Quick test commands:"
	@echo "curl http://localhost:8080/api/v1/health"

stop:
	@echo "🛑 Stopping all services..."
	docker-compose down

clean:
	@echo "🧹 Cleaning up..."
	docker-compose down -v --rmi all --remove-orphans
	docker system prune -f

logs:
	docker-compose logs -f

status:
	docker-compose ps

dev-setup:
	@echo "🔧 Setting up development environment..."
	@echo "1. Installing Go dependencies..."
	go mod tidy
	@echo "2. Generating protobuf files..."
	make proto
	@echo "3. Building services locally to check for errors..."
	make build-local
	@echo "4. Starting infrastructure services..."
	docker-compose up -d postgres rabbitmq
	@echo ""
	@echo "✅ Development environment ready!"
	@echo "Run 'make run' to start all services with Docker"
	@echo "Or run services individually with:"
	@echo "  go run risk-service/main.go"
	@echo "  go run notification-service/main.go"
	@echo "  go run user-service/main.go"
	@echo "  go run api-gateway/main.go"

test-system:
	@echo "🧪 Testing basic system functionality..."
	@echo "Testing health endpoint..."
	curl -f http://localhost:8080/api/v1/health || (echo "❌ Health check failed" && exit 1)
	@echo "✅ Health check passed"

test-auth:
	@chmod +x tests/test-auth.sh
	@./tests/test-auth.sh

test-users:
	@chmod +x tests/test-users.sh
	@./tests/test-users.sh

test-errors:
	@chmod +x tests/test-errors.sh
	@./tests/test-errors.sh

test-risk-detection:
	@chmod +x tests/test-risk-detection.sh
	@./tests/test-risk-detection.sh

test-risk-management:
	@chmod +x tests/test-risk-management.sh
	@./tests/test-risk-management.sh

test-admin:
	@chmod +x tests/test-admin.sh
	@./tests/test-admin.sh


test-all-endpoints:
	@echo "🧪 Running Comprehensive Endpoint Tests..."
	@echo "==============================================="
	@echo ""

	@echo "🏥 Testing Health Endpoint..."
	@make --no-print-directory test-system
	@echo ""

	@echo "🔐 Testing Authentication..."
	@make --no-print-directory test-auth
	@echo ""

	@echo "👤 Testing User Management..."
	@make --no-print-directory test-users
	@echo ""

	@echo "🚫 Testing Error Handling..."
	@make --no-print-directory test-errors
	@echo ""

	@echo "🚨 Testing Risk Detection..."
	@make --no-print-directory test-risk-detection
	@echo ""

	@echo "📊 Testing Admin Functions..."
	@make --no-print-directory test-admin
	@echo ""

	@echo "🎉 All endpoint tests completed!"
	@echo "Check the logs above for any failures."
	@echo ""
	@echo "To view service logs: make logs"
	@echo "To view RabbitMQ management: http://localhost:15672"

test-performance:
	@echo "⚡ Running Performance Tests..."
	@echo ""

	@echo "1. Testing concurrent registrations..."
	@for i in $$(seq 1 10); do \
		curl -s -X POST http://localhost:8080/api/v1/auth/register \
			-H "Content-Type: application/json" \
			-d "{\"email\":\"perf$$i@example.com\",\"password\":\"perfpass123\",\"first_name\":\"Perf\",\"last_name\":\"User$$i\"}" & \
	done; \
	wait; \
	echo "✅ Concurrent registrations completed"
	@echo ""

	@echo "2. Testing concurrent logins..."
	@for i in $$(seq 1 5); do \
		curl -s -X POST http://localhost:8080/api/v1/auth/login \
			-H "Content-Type: application/json" \
			-d "{\"email\":\"perf$$i@example.com\",\"password\":\"perfpass123\"}" & \
	done; \
	wait; \
	echo "✅ Concurrent logins completed"
	@echo ""

load-test-data:
	@echo "📊 Loading comprehensive test data..."
	@echo ""

	@echo "1. Creating normal users..."
	@curl -s -X POST http://localhost:8080/api/v1/auth/register \
		-H "Content-Type: application/json" \
		-d '{"email":"alice@company.com","password":"password123","first_name":"Alice","last_name":"Johnson","phone":"+1555000123"}' > /dev/null
	@curl -s -X POST http://localhost:8080/api/v1/auth/register \
		-H "Content-Type: application/json" \
		-d '{"email":"bob@company.com","password":"password123","first_name":"Bob","last_name":"Smith","phone":"+1555000124"}' > /dev/null
	@echo "✅ Normal users created"
	@echo ""

	@echo "2. Creating high-risk users..."
	@curl -s -X POST http://localhost:8080/api/v1/auth/register \
		-H "Content-Type: application/json" \
		-d '{"email":"test@suspicious.com","password":"password123","first_name":"Suspicious","last_name":"User","phone":"+9876543210"}' > /dev/null
	@curl -s -X POST http://localhost:8080/api/v1/auth/register \
		-H "Content-Type: application/json" \
		-d '{"email":"fake@scammer.org","password":"password123","first_name":"Fake","last_name":"Name","phone":"+1234567890"}' > /dev/null
	@echo "✅ High-risk users created"
	@echo ""

	@echo "3. Creating admin user..."
	@curl -s -X POST http://localhost:8080/api/v1/auth/register \
		-H "Content-Type: application/json" \
		-d '{"email":"admin@company.com","password":"adminpass123","first_name":"System","last_name":"Administrator"}' > /dev/null
	@echo "✅ Admin user created"
	@echo ""

	@echo "📋 Test data summary:"
	@echo "Normal users: alice@company.com, bob@company.com"
	@echo "High-risk users: test@suspicious.com, fake@scammer.org"
	@echo "Admin user: admin@company.com (needs manual role assignment)"
	@echo ""
	@echo "To promote admin user, run:"
	@echo "docker-compose exec postgres psql -U postgres -d users -c \"UPDATE users SET roles = '[\\\"admin\\\"]' WHERE email = 'admin@company.com';\""

test-ci:
	@echo "🚀 Running CI/CD Quick Tests..."
	@sleep 10
	@curl -f http://localhost:8080/api/v1/health || exit 1
	@curl -s -X POST http://localhost:8080/api/v1/auth/register \
		-H "Content-Type: application/json" \
		-d '{"email":"ci@test.com","password":"citest123","first_name":"CI","last_name":"Test"}' | grep -q "access_token" || exit 1
	@echo "✅ CI tests passed"

test-summary:
	@echo "📋 Test Summary"
	@echo "==============="
	@echo ""
	@echo "Available test commands:"
	@echo "  make test-system        - Basic health and connectivity"
	@echo "  make test-auth          - Authentication endpoints"
	@echo "  make test-users         - User management endpoints"
	@echo "  make test-admin         - Admin-only endpoints"
	@echo "  make test-errors        - Error handling and security"
	@echo "  make test-risk-detection - Risk assessment functionality"
	@echo "  make test-performance   - Performance and load testing"
	@echo "  make test-all-endpoints - Complete comprehensive test suite"
	@echo "  make test-ci            - Quick CI/CD tests"
	@echo ""
	@echo "Data management:"
	@echo "  make load-test-data     - Load sample test data"
	@echo ""
	@echo "To run all tests: make test-all-endpoints"
