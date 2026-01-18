package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/agent-platform/platform/internal/api"
	"github.com/yourusername/agent-platform/platform/internal/config"
	"github.com/yourusername/agent-platform/platform/internal/database"
	grpcserver "github.com/yourusername/agent-platform/platform/internal/grpc"
	"github.com/yourusername/agent-platform/platform/internal/monitor"
)

func main() {
	configPath := flag.String("config", "platform/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
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

	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	log.Println("Database connected successfully")

	// 启动监控
	monitor.StartMonitoring()

	// 启动 gRPC 服务器
	grpcServer := grpcserver.NewServer(cfg.Server.GRPCPort, db)
	go func() {
		log.Printf("Starting gRPC server on %s", cfg.Server.GRPCPort)
		if err := grpcServer.Start(); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// 启动 HTTP API 服务器
	router := api.SetupRouter(db)
	go func() {
		log.Printf("Starting HTTP server on %s", cfg.Server.HTTPPort)
		if err := router.Run(cfg.Server.HTTPPort); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Server shutting down")
	grpcServer.Stop()
}
