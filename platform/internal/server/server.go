package server

import (
	"fmt"
	"net"

	pb "github.com/yourusername/agent-platform/proto"
	grpcHandler "github.com/yourusername/agent-platform/platform/internal/grpc"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Server struct {
	addr       string
	grpcServer *grpc.Server
	listener   net.Listener
	db         *gorm.DB
}

func NewServer(addr string, db *gorm.DB) *Server {
	s := &Server{
		addr:       addr,
		grpcServer: grpc.NewServer(),
		db:         db,
	}

	handler := grpcHandler.NewAgentServiceHandler(db)
	pb.RegisterAgentServiceServer(s.grpcServer, handler)

	return s
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.listener = lis

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}
