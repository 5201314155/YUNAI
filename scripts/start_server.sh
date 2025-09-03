#!/bin/bash

# YUNAI服务器快速启动脚本 (Bash)

set -e

# 配置参数
DB_PASSWORD=${DB_PASSWORD:-"5201314hdz"}
PORT=${PORT:-8081}
TEST_MODE=${TEST_MODE:-false}

echo "🚀 启动YUNAI服务器"
echo "==============================================================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 设置环境变量
export PGPASSWORD=$DB_PASSWORD

# 检查端口是否被占用
if lsof -Pi :$PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️ 端口 $PORT 已被占用，正在尝试停止现有服务...${NC}"
    
    # 尝试停止占用端口的进程
    PID=$(lsof -Pi :$PORT -sTCP:LISTEN -t)
    if [ ! -z "$PID" ]; then
        kill -9 $PID 2>/dev/null || true
        echo -e "${GREEN}✅ 已停止现有进程 (PID: $PID)${NC}"
        sleep 2
    fi
fi

# 切换到backend目录
if [ -d "backend" ]; then
    cd backend
elif [ -d "../backend" ]; then
    cd ../backend
else
    echo -e "${RED}❌ 找不到backend目录${NC}"
    exit 1
fi

# 检查必要文件
REQUIRED_FILES=(
    "yunai_core_server.go"
    "user_identity_service.go"
    "global_prompt_service.go"
    "ai_invitation_service.go"
    "story_trigger_service.go"
    "world_setting_service.go"
    "voice_call_service.go"
)

MISSING_FILES=()
for file in "${REQUIRED_FILES[@]}"; do
    if [ ! -f "$file" ]; then
        MISSING_FILES+=("$file")
    fi
done

if [ ${#MISSING_FILES[@]} -gt 0 ]; then
    echo -e "${RED}❌ 缺少必要文件:${NC}"
    for file in "${MISSING_FILES[@]}"; do
        echo -e "   - $file"
    done
    exit 1
fi

echo -e "${GREEN}✅ 所有必要文件检查通过${NC}"

# 启动服务器
echo -e "\n${BLUE}🔄 启动YUNAI核心服务器...${NC}"
echo -e "端口: $PORT"
echo -e "数据库: yunai@localhost"

if [ "$TEST_MODE" = "true" ]; then
    echo -e "${YELLOW}🧪 测试模式：服务器将在测试完成后自动停止${NC}"
    
    # 后台启动服务器
    go run yunai_core_server.go user_identity_service.go global_prompt_service.go ai_invitation_service.go story_trigger_service.go world_setting_service.go voice_call_service.go &
    SERVER_PID=$!
    
    # 等待服务器启动
    sleep 8
    
    # 检查服务器状态
    if curl -s "http://localhost:$PORT/health" > /dev/null; then
        echo -e "${GREEN}✅ 服务器启动成功，开始功能测试...${NC}"
        
        # 运行功能测试
        cd ..
        if [ -f "test_all_systems.go" ]; then
            go run test_all_systems.go
        else
            echo -e "${YELLOW}⚠️ 测试文件不存在${NC}"
        fi
    else
        echo -e "${RED}❌ 服务器启动失败${NC}"
    fi
    
    # 停止服务器
    echo -e "\n${YELLOW}🛑 停止测试服务器...${NC}"
    kill $SERVER_PID 2>/dev/null || true
    
else
    echo -e "${BLUE}🔄 正常模式：服务器将持续运行${NC}"
    echo -e "按 Ctrl+C 停止服务器"
    echo ""
    
    # 直接启动服务器
    go run yunai_core_server.go user_identity_service.go global_prompt_service.go ai_invitation_service.go story_trigger_service.go world_setting_service.go voice_call_service.go
fi

echo -e "\n${GREEN}✨ 启动脚本执行完成！${NC}"
