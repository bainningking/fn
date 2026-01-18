package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/agent-platform/platform/internal/config"
	"github.com/yourusername/agent-platform/platform/internal/database"
	"github.com/yourusername/agent-platform/platform/internal/server"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("platform/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	dbCfg := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
	}

	_, err = database.Connect(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	log.Println("Database connected successfully")

	// 启动 gRPC 服务器
	grpcAddr := fmt.Sprintf(":%d", cfg.Server.GRPCPort)
	grpcServer := server.NewServer(grpcAddr)

	go func() {
		log.Printf("Starting gRPC server on %s", grpcAddr)
		if err := grpcServer.Start(); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Server shutting down")
	grpcServer.Stop()
}
