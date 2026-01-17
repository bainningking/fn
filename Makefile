.PHONY: proto build-agent build-platform test clean

# 生成 protobuf 代码
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto

# 构建 Agent
build-agent:
	go build -o bin/agent ./agent/cmd/agent

# 构建管理平台
build-platform:
	go build -o bin/server ./platform/cmd/server

# 运行测试
test:
	go test -v ./...

# 清理构建产物
clean:
	rm -rf bin/
