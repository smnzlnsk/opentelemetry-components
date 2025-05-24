package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/smnzlnsk/opentelemetry-components/internal/shared/job"
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

	// HTTP server instance
	server *http.Server

	// metrics service
	// needed to retrieve service metrics for evaluation
	metricsService domain.MetricsService
}

// ServerConfig contains configuration for the HTTP server
type ServerConfig struct {
	Host string
	Port int
}

// NewServer creates a new policy evaluation HTTP server
func NewServer(config ServerConfig, logger *zap.Logger, policies map[string]domain.Policy, metricsService domain.MetricsService) *Server {
	return &Server{
		host:           config.Host,
		port:           config.Port,
		router:         http.NewServeMux(),
		logger:         logger,
		policies:       policies,
		metricsService: metricsService,
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

	var requestBody job.Request

	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Failed to parse request body: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

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
	processors := policy.HeuristicEntity().Processors()

	var processorsToEnforce []string

	if processorName != "" {
		// Check if specified processor exists
		_, exists := processors[processorName]
		if !exists {
			http.Error(w, fmt.Sprintf("Processor '%s' not found", processorName), http.StatusNotFound)
			return
		}
		processorsToEnforce = []string{processorName}
	} else {
		// Default to processing all processors, if none specified
		for name := range processors {
			processorsToEnforce = append(processorsToEnforce, name)
		}
	}

	// Enforce policies on all selected processors
	for _, name := range processorsToEnforce {
		if err := policy.Enforce(name, requestBody); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	/*
		for i := range instances {
			values, err := s.metricsService.GetJobMetricsAsMap(
				context.Background(),
				jobName,
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			instanceValues := values.InstanceMetricsForEvaluation(
				fmt.Sprintf("%s.instance.%d", jobName, instances[i].InstanceNumber),
			)
			instances[i].Priority = processor.Evaluator().Evaluate(1, instanceValues) // TODO: incorporate job name for evaluation
		}
	*/

	http.Error(w, "ok", http.StatusOK)
}
