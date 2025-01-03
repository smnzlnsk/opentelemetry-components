package oakestraprocessor

import (
	"context"
	"fmt"
	"net"
	"sync"

	pb "github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	server *grpc.Server
	logger *zap.Logger
	proc   *MultiProcessor
	port   int
	mu     sync.Mutex
}

type monitoringServer struct {
	pb.UnimplementedMonitoringServiceServer
	logger *zap.Logger
	proc   *MultiProcessor
}

func NewGRPCServer(proc *MultiProcessor, port int) *GRPCServer {
	return &GRPCServer{
		port:   port,
		logger: proc.logger,
		proc:   proc,
	}
}

func (g *GRPCServer) Start() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.server != nil {
		return fmt.Errorf("gRPC server already running")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	g.server = grpc.NewServer()
	pb.RegisterMonitoringServiceServer(g.server, &monitoringServer{
		logger: g.logger,
		proc:   g.proc,
	})

	go func() {
		if err := g.server.Serve(lis); err != nil {
			g.logger.Error("Failed to serve gRPC", zap.Error(err))
		}
	}()

	g.logger.Info("gRPC server started", zap.Int("port", g.port))
	return nil
}

func (g *GRPCServer) Stop() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.server != nil {
		g.server.GracefulStop()
		g.server = nil
		g.logger.Info("gRPC server stopped")
	}
}

func (s *monitoringServer) NotifyDeployment(ctx context.Context, req *pb.MonitoringDeploymentRequest) (*pb.MonitoringResponse, error) {
	s.logger.Info("Received deployment",
		zap.String("job_name", req.JobName),
		zap.String("job_hash", req.JobHash),
		zap.Int32("instance_number", req.InstanceNumber),
		zap.String("cpu", req.Resource.Cpu),
		zap.String("memory", req.Resource.Memory),
		zap.String("gpu", req.Resource.Gpu),
		zap.String("network_bandwidth_in", req.Resource.Network.BandwidthIn),
		zap.String("network_bandwidth_out", req.Resource.Network.BandwidthOut),
		zap.String("disk", req.Resource.Disk),
	)

	s.proc.registerService(req.JobName, req.InstanceNumber, req.Resource)

	return &pb.MonitoringResponse{
		Acknowledged: true,
		Message:      "Successfully processed deployment",
	}, nil
}

func (s *monitoringServer) NotifyDeletion(ctx context.Context, req *pb.MonitoringDeletionRequest) (*pb.MonitoringResponse, error) {
	s.logger.Info("Received deletion", zap.String("job_name", req.JobName), zap.Int32("instance_number", req.InstanceNumber))

	s.proc.deleteService(req.JobName, req.InstanceNumber)

	return &pb.MonitoringResponse{
		Acknowledged: true,
		Message:      "Successfully processed deletion",
	}, nil
}
