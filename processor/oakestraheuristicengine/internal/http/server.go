package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraheuristicengine/internal/domain"
	"go.uber.org/zap"
)

// Server represents the HTTP server for policy evaluation
type Server struct {
	host   string
	port   int
	router *http.ServeMux
	logger *zap.Logger

	// policies
	policies map[string]domain.Policy

	// metric store
	metricStore domain.MetricStore

	// HTTP server instance
	server *http.Server
}

// ServerConfig contains configuration for the HTTP server
type ServerConfig struct {
	Host string
	Port int
}

// NewServer creates a new policy evaluation HTTP server
func NewServer(config ServerConfig, logger *zap.Logger, policies map[string]domain.Policy, metricStore domain.MetricStore) *Server {
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
		s.logger.Info("HTTP server was nil, nothing to shut down")
		return nil
	}

	s.logger.Info("Shutting down HTTP server", zap.String("address", s.server.Addr))

	// Create a context with timeout to ensure we don't hang forever
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := s.server.Shutdown(shutdownCtx)
	if err != nil {
		if err == context.DeadlineExceeded {
			s.logger.Warn("HTTP server shutdown timed out, forcing close", zap.Error(err))
			// Force close if timeout occurs
			return s.server.Close()
		}
		s.logger.Error("Error during HTTP server shutdown", zap.Error(err))
		return err
	}

	s.logger.Info("HTTP server shutdown completed successfully")
	return nil
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	s.router.HandleFunc("/policy/", s.handlePolicy)
}

// handlePolicy processes requests to the /policy/<policyName>/<processorName> endpoint
func (s *Server) handlePolicy(w http.ResponseWriter, r *http.Request) {
	// Parse appName from request body

	var requestBody domain.Job

	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Failed to parse request body: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	jobName := requestBody.JobName
	instances := requestBody.ServiceInstanceList

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

	// Get the policy from the policy store
	policy, ok := s.policies[policyName]
	if !ok {
		http.Error(w, "Policy not found", http.StatusNotFound)
		return
	}

	// Get the processor from the policy
	values := s.metricStore.GetValueMapByString()
	fmt.Println("values", values)
	processors := policy.HeuristicEngine().Processors()

	// Check if a specific processor was requested
	var processor domain.Processor
	var exists bool

	if processorName != "" {
		// Check if specified processor exists
		processor, exists = processors[processorName]
		if !exists {
			http.Error(w, fmt.Sprintf("Processor '%s' not found", processorName), http.StatusNotFound)
			return
		}
	} else {
		// Default to "routing" processor if none specified
		processor, exists = processors["default"]
		if !exists {
			http.Error(w, "Processor not found", http.StatusNotFound)
			return
		}
	}

	for i := range instances {
		instances[i].Priority = processor.Evaluator().Evaluate(1, values) // TODO: incorporate job name for evaluation
	}

	err := policy.Enforce(processorName, jobName, values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"policy":    policy.Name(),
		"processor": processorName,
		"result":    instances,
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
