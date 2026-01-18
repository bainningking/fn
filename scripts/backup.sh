#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 配置
BACKUP_DIR="/var/backups/agent-platform"
DB_NAME="agent_platform"
DB_USER="agent_user"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo -e "${GREEN}开始备份 Agent 管理平台...${NC}"

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}请使用 root 权限运行此脚本${NC}"
    exit 1
fi

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份数据库
echo -e "${YELLOW}备份数据库...${NC}"
pg_dump -U $DB_USER $DB_NAME > $BACKUP_DIR/db_$TIMESTAMP.sql
gzip $BACKUP_DIR/db_$TIMESTAMP.sql

# 备份配置文件
echo -e "${YELLOW}备份配置文件...${NC}"
tar -czf $BACKUP_DIR/config_$TIMESTAMP.tar.gz /opt/agent-platform/config.yaml

# 清理旧备份（保留最近7天）
echo -e "${YELLOW}清理旧备份...${NC}"
find $BACKUP_DIR -name "db_*.sql.gz" -mtime +7 -delete
find $BACKUP_DIR -name "config_*.tar.gz" -mtime +7 -delete

echo -e "${GREEN}备份完成！${NC}"
echo "数据库备份: $BACKUP_DIR/db_$TIMESTAMP.sql.gz"
echo "配置备份: $BACKUP_DIR/config_$TIMESTAMP.tar.gz"
