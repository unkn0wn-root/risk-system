#!/bin/bash

set -e

echo "🚀 Setting up User Risk Management System..."

echo "📋 Checking prerequisites..."

if ! command -v go >/dev/null 2>&1; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
    echo "❌ Docker is not installed. Please install Docker."
    exit 1
fi

if ! command -v docker-compose >/dev/null 2>&1; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose."
    exit 1
fi

echo "✅ Prerequisites check passed"

echo "📦 Setting up Go module..."
if [ ! -f "go.mod" ]; then
    go mod init user-risk-system
fi

if ! command -v protoc >/dev/null 2>&1; then
    echo "⚠️  protoc not found. Installing..."
    case "$(uname -s)" in
        Darwin*)
            if command -v brew >/dev/null 2>&1; then
                brew install protobuf
            else
                echo "❌ Please install Homebrew and run: brew install protobuf"
                exit 1
            fi
            ;;
        Linux*)
            sudo apt-get update && sudo apt-get install -y protobuf-compiler
            ;;
        *)
            echo "❌ Please install protoc manually"
            exit 1
            ;;
    esac
fi

echo "🔧 Running development setup..."
make dev-setup

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "🚀 Next steps:"
echo "1. Start all services: make run"
echo "2. Test the system: make test-system"
echo "3. Load test data: make load-test-data"
echo "4. View logs: make logs"
echo ""
echo "📚 Useful URLs:"
echo "• API Gateway: http://localhost:8080"
echo "• RabbitMQ Management: http://localhost:15672 (guest/guest)"
echo "• API Documentation: http://localhost:8080/api/docs (swagger docs)"
