package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/agent-platform/agent/internal/client"
	"github.com/yourusername/agent-platform/agent/internal/config"
)

func main() {
	configPath := flag.String("config", "agent/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建客户端
	c := client.NewClient(cfg.Server.Address, cfg.Server.TLS, cfg.Agent.ID)

	// 连接到服务器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	log.Printf("Agent %s started", cfg.Agent.ID)

	// 运行客户端
	go func() {
		if err := c.Run(context.Background()); err != nil {
			log.Printf("Client error: %v", err)
		}
	}()

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Agent shutting down")
}
