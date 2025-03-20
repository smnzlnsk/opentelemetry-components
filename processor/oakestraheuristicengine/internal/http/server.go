package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/common/interfaces"
	"go.uber.org/zap"
)

// Server represents the HTTP server for tablequery
type Server struct {
	host   string
	port   int
	router *http.ServeMux
	logger *zap.Logger

	// policies
	policies map[string]interfaces.Policy

	// metric store
	metricStore interfaces.MetricStore

	// HTTP server instance
	server *http.Server
}

// ServerConfig contains configuration for the HTTP server
type ServerConfig struct {
	Host string
	Port int
}

// NewServer creates a new tablequery HTTP server
func NewServer(config ServerConfig, logger *zap.Logger, policies map[string]interfaces.Policy, metricStore interfaces.MetricStore) *Server {
	return &Server{
		host:        config.Host,
		port:        config.Port,
		router:      http.NewServeMux(),
		logger:      logger,
		policies:    policies,
		metricStore: metricStore,
	}
}

// Start initializes and starts the HTTP server
func (s *Server) Start() error {
	s.setupRoutes()

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Start the server in a new goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	s.logger.Info("HTTP server started", zap.String("address", addr))

	return nil
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	s.router.HandleFunc("/policy/", s.handlePolicy)
}

// handleTableQuery processes requests to the /tablequery/<routingPolicy> endpoint
func (s *Server) handlePolicy(w http.ResponseWriter, r *http.Request) {
	var success bool
	// Extract the routing policy from the URL path
	path := strings.TrimPrefix(r.URL.Path, "/policy/")
	if path == "" {
		http.Error(w, "Policy not specified", http.StatusBadRequest)
		return
	}

	// Split the path to get policyName and processorName
	pathParts := strings.Split(path, "/")
	if len(pathParts) < 1 {
		http.Error(w, "Policy not specified", http.StatusBadRequest)
		return
	}

	policyName := pathParts[0]
	processorName := ""
	if len(pathParts) > 1 {
		processorName = pathParts[1]
	}

	// Process the request based on the routing policy

	// get the policy from the policy store
	policy, ok := s.policies[policyName]
	if !ok {
		http.Error(w, "Policy not found", http.StatusNotFound)
		return
	}

	// get the processor from the policy
	values := s.metricStore.GetValueMapByString()
	processors := policy.HeuristicEngine().Processors()

	// Check if a specific processor was requested
	var processorResult float64
	if processorName != "" {
		processor, exists := processors[processorName]
		if !exists {
			http.Error(w, fmt.Sprintf("Processor '%s' not found", processorName), http.StatusNotFound)
			return
		}
		processorResult = processor.Evaluator().Evaluate(1, values)
		success = true
	} else {
		// Default to "routing" processor if none specified
		if routingProcessor, exists := processors["default"]; exists {
			processorResult = routingProcessor.Evaluator().Evaluate(1, values)
			success = true
		} else {
			http.Error(w, "Processor not found", http.StatusNotFound)
			return
		}
	}

	response := map[string]interface{}{
		"status":    map[bool]string{true: "success", false: "failed"}[success],
		"policy":    policy.Name(),
		"processor": processorName,
		"result":    processorResult,
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Example usage:
// func main() {
//     config := ServerConfig{
//         Host: "0.0.0.0",
//         Port: 8080,
//     }
//     server := NewServer(config)
//     if err := server.Start(); err != nil {
//         log.Fatalf("Failed to start server: %v", err)
//     }
// }
