package executor

import (
	"context"
	"testing"
	"time"
)

func TestExecuteShellScript(t *testing.T) {
	executor := NewExecutor()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, "shell", "echo 'hello world'", 10)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	if result.Stdout != "hello world\n" {
		t.Errorf("expected stdout 'hello world\\n', got %q", result.Stdout)
	}
}

func TestExecuteWithTimeout(t *testing.T) {
	executor := NewExecutor()

	ctx := context.Background()

	// 执行一个会超时的脚本（超时设置为1秒，但脚本需要10秒）
	_, err := executor.Execute(ctx, "shell", "sleep 10", 1)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
