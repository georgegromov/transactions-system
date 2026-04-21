package consul

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
)

type Config struct {
	// ConsulHTTPAddr e.g. http://consul:8500 — if empty, registration is skipped.
	ConsulHTTPAddr string
	// ServiceID unique instance id (default: service-name-port).
	ServiceID string
	// ServiceName logical name in Consul (e.g. order-service).
	ServiceName string
	// ServiceHost hostname reachable from Consul for health checks (Docker service name).
	ServiceHost string
	Port        int
	// HealthPath defaults to /health
	HealthPath string
}

// RegisterMaybe registers the service when ConsulHTTPAddr and ServiceHost are set.
// The returned cleanup deregisters the service (safe to call multiple times).
func RegisterMaybe(cfg Config) (func(), error) {
	noop := func() {}

	if strings.TrimSpace(cfg.ConsulHTTPAddr) == "" || strings.TrimSpace(cfg.ServiceHost) == "" {
		return noop, nil
	}

	if cfg.HealthPath == "" {
		cfg.HealthPath = "/health"
	}
	if !strings.HasPrefix(cfg.HealthPath, "/") {
		cfg.HealthPath = "/" + cfg.HealthPath
	}
	if cfg.ServiceID == "" {
		cfg.ServiceID = fmt.Sprintf("%s-%d", cfg.ServiceName, cfg.Port)
	}

	consulAddr := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(cfg.ConsulHTTPAddr), "http://"), "https://")
	consulAddr = strings.TrimSuffix(consulAddr, "/")

	client, err := api.NewClient(&api.Config{Address: consulAddr})
	if err != nil {
		return nil, fmt.Errorf("consul client: %w", err)
	}

	checkURL := fmt.Sprintf("http://%s:%d%s", cfg.ServiceHost, cfg.Port, cfg.HealthPath)
	reg := &api.AgentServiceRegistration{
		ID:      cfg.ServiceID,
		Name:    cfg.ServiceName,
		Address: cfg.ServiceHost,
		Port:    cfg.Port,
		Check: &api.AgentServiceCheck{
			HTTP:                           checkURL,
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "1m",
		},
	}

	if err := client.Agent().ServiceRegister(reg); err != nil {
		return nil, fmt.Errorf("consul register: %w", err)
	}
	log.Printf("consul: registered %s (%s)", cfg.ServiceID, checkURL)

	return func() {
		if err := client.Agent().ServiceDeregister(cfg.ServiceID); err != nil {
			log.Printf("consul: deregister %s: %v", cfg.ServiceID, err)
			return
		}
		log.Printf("consul: deregistered %s", cfg.ServiceID)
	}, nil
}

// ListenPort parses TCP listen address like ":8081" or "0.0.0.0:8081".
func ListenPort(addr string) int {
	_, portStr, err := net.SplitHostPort(addr)
	if err == nil && portStr != "" {
		p, _ := strconv.Atoi(portStr)
		if p > 0 {
			return p
		}
	}
	if strings.HasPrefix(addr, ":") {
		p, _ := strconv.Atoi(strings.TrimPrefix(addr, ":"))
		if p > 0 {
			return p
		}
	}
	return 0
}
