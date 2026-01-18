#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 配置
INSTALL_DIR="/opt/agent-platform"
CONFIG_FILE="$INSTALL_DIR/config.yaml"
SYSTEMD_DIR="/etc/systemd/system"

echo -e "${GREEN}开始部署 Agent 管理平台...${NC}"

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}请使用 root 权限运行此脚本${NC}"
    exit 1
fi

# 创建安装目录
echo -e "${YELLOW}创建安装目录...${NC}"
mkdir -p $INSTALL_DIR/bin
mkdir -p $INSTALL_DIR/logs
mkdir -p /var/log/agent-platform

# 构建应用
echo -e "${YELLOW}构建应用...${NC}"
go build -o $INSTALL_DIR/bin/platform-server platform/cmd/server/main.go
go build -o $INSTALL_DIR/bin/agent agent/cmd/main.go

# 复制配置文件
echo -e "${YELLOW}配置应用...${NC}"
if [ ! -f "$CONFIG_FILE" ]; then
    cp platform/config.example.yaml $CONFIG_FILE
    echo -e "${YELLOW}请编辑配置文件: $CONFIG_FILE${NC}"
fi

# 创建用户
echo -e "${YELLOW}创建系统用户...${NC}"
if ! id -u agent > /dev/null 2>&1; then
    useradd -r -s /bin/false agent
fi

# 设置权限
chown -R agent:agent $INSTALL_DIR
chown -R agent:agent /var/log/agent-platform

# 安装 systemd 服务
echo -e "${YELLOW}安装 systemd 服务...${NC}"

# 平台服务
cat > $SYSTEMD_DIR/agent-platform.service <<EOF
[Unit]
Description=Agent Management Platform
After=network.target postgresql.service

[Service]
Type=simple
User=agent
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/bin/platform-server -config $CONFIG_FILE
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

# Agent 服务模板
cat > $SYSTEMD_DIR/agent@.service <<EOF
[Unit]
Description=Agent %i
After=network.target agent-platform.service

[Service]
Type=simple
User=agent
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/bin/agent -server localhost:50051 -id %i
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

# 重载 systemd
systemctl daemon-reload

echo -e "${GREEN}部署完成！${NC}"
echo -e "${YELLOW}下一步操作：${NC}"
echo "1. 编辑配置文件: $CONFIG_FILE"
echo "2. 启动平台服务: systemctl start agent-platform"
echo "3. 启用开机自启: systemctl enable agent-platform"
echo "4. 启动 Agent: systemctl start agent@agent-001"
echo "5. 查看状态: systemctl status agent-platform"
