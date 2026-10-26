package discovery

import (
	"fmt"
	"log/slog"

	"github.com/hashicorp/consul/api"
)

type Registry struct {
	client *api.Client
	logger *slog.Logger
}

func NewRegistry(address string, logger *slog.Logger) (*Registry, error) {
	cfg := api.DefaultConfig()
	cfg.Address = address

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &Registry{client: client, logger: logger}, nil
}

// RegisterService acts like the @EnableEurekaClient logic in Spring
func (r *Registry) RegisterService(id, name, host string, port int) error {
	registration := &api.AgentServiceRegistration{
		ID:      id,   // Unique instance id (e.g., users-service-1)
		Name:    name, // Service name (e.g., users-service)
		Address: host,
		Port:    port,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", host, port),
			Interval: "10s",
			Timeout:  "5s",
		},
	}

	r.logger.Info("registering service with consul", slog.String("name", name), slog.Int("port", port))
	return r.client.Agent().ServiceRegister(registration)
}
