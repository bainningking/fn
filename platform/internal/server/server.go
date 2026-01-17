package server

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	addr       string
	grpcServer *grpc.Server
	listener   net.Listener
}

func NewServer(addr string) *Server {
	return &Server{
		addr:       addr,
		grpcServer: grpc.NewServer(),
	}
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
