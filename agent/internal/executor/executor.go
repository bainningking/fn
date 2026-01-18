package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type ExecutionResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

type Executor struct {
	maxConcurrent int
	semaphore     chan struct{}
}

func NewExecutor() *Executor {
	maxConcurrent := 5
	return &Executor{
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
	}
}

func (e *Executor) Execute(ctx context.Context, scriptType, script string, timeoutSeconds int) (*ExecutionResult, error) {
	// 获取信号量，限制并发执行数
	select {
	case e.semaphore <- struct{}{}:
		defer func() { <-e.semaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 创建超时上下文
	if timeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
		defer cancel()
	}

	var cmd *exec.Cmd
	switch scriptType {
	case "shell":
		cmd = exec.CommandContext(ctx, "sh", "-c", script)
	case "python":
		cmd = exec.CommandContext(ctx, "python3", "-c", script)
	default:
		return nil, fmt.Errorf("unsupported script type: %s", scriptType)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &ExecutionResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			if exitCode == -1 {
				// 进程被信号终止（如超时），返回错误
				return nil, err
			}
			result.ExitCode = exitCode
			return result, nil
		}
		return nil, err
	}

	result.ExitCode = 0
	return result, nil
}
