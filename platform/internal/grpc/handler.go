package grpc

import (
	"io"
	"log"

	pb "github.com/yourusername/agent-platform/proto"
	"github.com/yourusername/agent-platform/platform/internal/models"
	"gorm.io/gorm"
)

type AgentServiceHandler struct {
	pb.UnimplementedAgentServiceServer
	db *gorm.DB
}

func NewAgentServiceHandler(db *gorm.DB) *AgentServiceHandler {
	return &AgentServiceHandler{
		db: db,
	}
}

func (h *AgentServiceHandler) Connect(stream pb.AgentService_ConnectServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch m := msg.Message.(type) {
		case *pb.AgentMessage_Register:
			if err := h.handleRegister(stream, m.Register); err != nil {
				log.Printf("Error handling register: %v", err)
			}
		case *pb.AgentMessage_Heartbeat:
			if err := h.handleHeartbeat(stream, m.Heartbeat); err != nil {
				log.Printf("Error handling heartbeat: %v", err)
			}
		case *pb.AgentMessage_TaskResult:
			if err := h.handleTaskResult(stream, m.TaskResult); err != nil {
				log.Printf("Error handling task result: %v", err)
			}
		case *pb.AgentMessage_TaskLog:
			if err := h.handleTaskLog(stream, m.TaskLog); err != nil {
				log.Printf("Error handling task log: %v", err)
			}
		}
	}
}

func (h *AgentServiceHandler) handleRegister(stream pb.AgentService_ConnectServer, register *pb.AgentRegister) error {
	// 处理 Agent 注册逻辑
	log.Printf("Agent registered: %s", register.AgentId)
	return stream.Send(&pb.ServerMessage{
		Message: &pb.ServerMessage_RegisterResponse{
			RegisterResponse: &pb.Response{
				Success: true,
			},
		},
	})
}

func (h *AgentServiceHandler) handleHeartbeat(stream pb.AgentService_ConnectServer, heartbeat *pb.Heartbeat) error {
	// 处理心跳逻辑
	return stream.Send(&pb.ServerMessage{
		Message: &pb.ServerMessage_HeartbeatAck{
			HeartbeatAck: &pb.Response{
				Success: true,
			},
		},
	})
}

func (h *AgentServiceHandler) handleTaskResult(stream pb.AgentService_ConnectServer, result *pb.TaskResult) error {
	// 更新任务结果
	taskResult := h.db.Model(&models.Task{}).
		Where("task_id = ?", result.TaskId).
		Updates(map[string]interface{}{
			"exit_code": result.ExitCode,
			"stdout":    result.Stdout,
			"stderr":    result.Stderr,
			"status":    "completed",
		})

	if taskResult.Error != nil {
		return taskResult.Error
	}

	return stream.Send(&pb.ServerMessage{
		Message: &pb.ServerMessage_RegisterResponse{
			RegisterResponse: &pb.Response{
				Success: true,
			},
		},
	})
}

func (h *AgentServiceHandler) handleTaskLog(stream pb.AgentService_ConnectServer, taskLog *pb.TaskLog) error {
	// 处理任务日志逻辑
	log.Printf("Task %s log: %s", taskLog.TaskId, taskLog.Output)
	return nil
}
