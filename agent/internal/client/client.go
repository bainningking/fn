package client

import (
	"context"
	"fmt"
	"io"
	"log"

	pb "github.com/yourusername/agent-platform/proto"
	"github.com/yourusername/agent-platform/agent/internal/executor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	serverAddr string
	useTLS     bool
	agentID    string
	conn       *grpc.ClientConn
	executor   *executor.Executor
}

func NewClient(serverAddr string, useTLS bool, agentID string) *Client {
	return &Client{
		serverAddr: serverAddr,
		useTLS:     useTLS,
		agentID:    agentID,
		executor:   executor.NewExecutor(),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	var opts []grpc.DialOption
	if !c.useTLS {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(ctx, c.serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	return nil
}

func (c *Client) Run(ctx context.Context) error {
	client := pb.NewAgentServiceClient(c.conn)
	stream, err := client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	// 发送注册消息
	if err := stream.Send(&pb.AgentMessage{
		Message: &pb.AgentMessage_Register{
			Register: &pb.AgentRegister{
				AgentId: c.agentID,
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}

	// 接收服务器消息
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to receive: %w", err)
		}

		switch m := msg.Message.(type) {
		case *pb.ServerMessage_TaskRequest:
			go c.handleTask(ctx, stream, m.TaskRequest)
		case *pb.ServerMessage_RegisterResponse:
			log.Printf("Registered successfully")
		case *pb.ServerMessage_HeartbeatAck:
			log.Printf("Heartbeat acknowledged")
		}
	}
}

func (c *Client) handleTask(ctx context.Context, stream pb.AgentService_ConnectClient, task *pb.TaskRequest) {
	log.Printf("Received task: %s", task.TaskId)

	// 将 TaskType 枚举转换为字符串
	var scriptType string
	switch task.Type {
	case pb.TaskType_TASK_TYPE_SHELL:
		scriptType = "shell"
	case pb.TaskType_TASK_TYPE_PYTHON:
		scriptType = "python"
	default:
		log.Printf("Unknown task type: %v", task.Type)
		return
	}

	result, err := c.executor.Execute(ctx, scriptType, task.Script, int(task.Timeout))

	taskResult := &pb.TaskResult{
		TaskId: task.TaskId,
	}

	if err != nil {
		taskResult.ExitCode = -1
		taskResult.Stderr = err.Error()
	} else {
		taskResult.ExitCode = int32(result.ExitCode)
		taskResult.Stdout = result.Stdout
		taskResult.Stderr = result.Stderr
	}

	if err := stream.Send(&pb.AgentMessage{
		Message: &pb.AgentMessage_TaskResult{
			TaskResult: taskResult,
		},
	}); err != nil {
		log.Printf("Failed to send task result: %v", err)
	}
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
