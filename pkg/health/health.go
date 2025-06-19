package health

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type ServiceHealth struct {
	Name   string
	Status grpc_health_v1.HealthCheckResponse_ServingStatus
}

type Config struct {
	// OverallStatus sets the overall server health
	OverallStatus grpc_health_v1.HealthCheckResponse_ServingStatus
	// Services is a list of services with their individual health status
	Services []ServiceHealth
}

// HealthServer wraps the gRPC health server to provide control methods
type HealthServer struct {
	server *health.Server
}

// RegisterHealthService registers gRPC health check service with the given server.
func RegisterHealthService(server *grpc.Server, config Config) *HealthServer {
	healthServer := health.NewServer()

	healthServer.SetServingStatus("", config.OverallStatus)

	for _, service := range config.Services {
		healthServer.SetServingStatus(service.Name, service.Status)
	}

	grpc_health_v1.RegisterHealthServer(server, healthServer)

	return &HealthServer{server: healthServer}
}

// SetServingStatus allows runtime control of service health status
func (hs *HealthServer) SetServingStatus(service string, status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	hs.server.SetServingStatus(service, status)
}

// SetOverallStatus sets the overall server health status
func (hs *HealthServer) SetOverallStatus(status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	hs.server.SetServingStatus("", status)
}

// RegisterHealthServiceWithDefaults registers health service with default SERVING status
func RegisterHealthServiceWithDefaults(server *grpc.Server, serviceName string) *HealthServer {
	config := Config{
		OverallStatus: grpc_health_v1.HealthCheckResponse_SERVING,
		Services: []ServiceHealth{
			{Name: serviceName, Status: grpc_health_v1.HealthCheckResponse_SERVING},
		},
	}
	return RegisterHealthService(server, config)
}
