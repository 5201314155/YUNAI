#!/bin/bash

# YUNAI系统自动化部署脚本 (Bash)
# 适用于Linux/macOS环境

set -e  # 遇到错误立即退出

# 配置参数
DB_PASSWORD=${DB_PASSWORD:-"5201314hdz"}
DB_HOST=${DB_HOST:-"localhost"}
DB_USER=${DB_USER:-"postgres"}
DB_NAME=${DB_NAME:-"yunai"}
SERVER_PORT=${SERVER_PORT:-8081}

echo "🚀 YUNAI系统自动化部署开始"
echo "==============================================================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 1. 环境检查
echo -e "\n1. 🔍 环境检查..."

# 检查Go环境
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    echo -e "   ${GREEN}✅ Go环境: $GO_VERSION${NC}"
else
    echo -e "   ${RED}❌ Go环境未安装${NC}"
    exit 1
fi

# 检查PostgreSQL
export PGPASSWORD=$DB_PASSWORD
if psql -h $DB_HOST -U $DB_USER -d postgres -c "SELECT version();" &> /dev/null; then
    echo -e "   ${GREEN}✅ PostgreSQL连接成功${NC}"
else
    echo -e "   ${RED}❌ PostgreSQL连接失败${NC}"
    exit 1
fi

# 2. 数据库准备
echo -e "\n2. 💾 数据库准备..."

# 检查数据库是否存在
DB_EXISTS=$(psql -h $DB_HOST -U $DB_USER -d postgres -t -c "SELECT 1 FROM pg_database WHERE datname='$DB_NAME';" 2>/dev/null | xargs)
if [ "$DB_EXISTS" = "1" ]; then
    echo -e "   ${GREEN}✅ 数据库 '$DB_NAME' 已存在${NC}"
else
    echo -e "   ${BLUE}📝 创建数据库 '$DB_NAME'...${NC}"
    createdb -h $DB_HOST -U $DB_USER $DB_NAME
    echo -e "   ${GREEN}✅ 数据库创建成功${NC}"
fi

# 3. 初始化数据库表结构
echo -e "\n3. 🏗️ 初始化数据库表结构..."

if [ -f "database/init_database.sql" ]; then
    psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f "database/init_database.sql" > /dev/null 2>&1
    echo -e "   ${GREEN}✅ 数据库表结构初始化完成${NC}"
else
    echo -e "   ${YELLOW}⚠️ 数据库初始化脚本不存在，跳过${NC}"
fi

# 4. 修复外键约束
echo -e "\n4. 🔧 修复外键约束..."

if [ -f "fix_foreign_keys.sql" ]; then
    psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f "fix_foreign_keys.sql" > /dev/null 2>&1
    echo -e "   ${GREEN}✅ 外键约束修复完成${NC}"
else
    echo -e "   ${YELLOW}⚠️ 外键修复脚本不存在，跳过${NC}"
fi

# 5. 安装Go依赖
echo -e "\n5. 📦 安装Go依赖..."

cd backend
go mod tidy
if [ $? -eq 0 ]; then
    echo -e "   ${GREEN}✅ Go依赖安装完成${NC}"
else
    echo -e "   ${RED}❌ Go依赖安装失败${NC}"
    exit 1
fi

# 6. 编译项目
echo -e "\n6. 🔨 编译项目..."

go build .
if [ $? -eq 0 ]; then
    echo -e "   ${GREEN}✅ 项目编译成功${NC}"
else
    echo -e "   ${RED}❌ 项目编译失败${NC}"
    exit 1
fi

# 7. 运行功能测试
echo -e "\n7. 🧪 运行功能测试..."

# 启动服务器（后台）
export PGPASSWORD=$DB_PASSWORD
go run yunai_core_server.go user_identity_service.go global_prompt_service.go ai_invitation_service.go story_trigger_service.go world_setting_service.go voice_call_service.go &
SERVER_PID=$!

# 等待服务器启动
sleep 5

# 检查服务器是否启动
if curl -s "http://localhost:$SERVER_PORT/health" > /dev/null; then
    echo -e "   ${GREEN}✅ 服务器启动成功${NC}"
    
    # 运行功能测试
    cd ..
    if [ -f "test_all_systems.go" ]; then
        echo -e "   ${BLUE}🧪 执行功能测试...${NC}"
        go run test_all_systems.go
        echo -e "   ${GREEN}✅ 功能测试完成${NC}"
    fi
else
    echo -e "   ${RED}❌ 服务器启动失败${NC}"
fi

# 停止后台服务器
kill $SERVER_PID 2>/dev/null || true

# 8. 部署完成
echo -e "\n==============================================================================="
echo -e "${GREEN}🎉 YUNAI系统部署完成！${NC}"
echo -e "==============================================================================="
echo ""
echo -e "🌐 服务器地址: http://localhost:$SERVER_PORT"
echo -e "📊 健康检查: http://localhost:$SERVER_PORT/health"
echo -e "📈 系统统计: http://localhost:$SERVER_PORT/stats"
echo -e "🎯 API文档: http://localhost:$SERVER_PORT/api/docs"
echo ""
echo -e "${YELLOW}🚀 启动服务器命令:${NC}"
echo -e "${CYAN}cd backend${NC}"
echo -e "${CYAN}go run yunai_core_server.go user_identity_service.go global_prompt_service.go ai_invitation_service.go story_trigger_service.go world_setting_service.go voice_call_service.go${NC}"
echo ""

echo -e "✨ 部署脚本执行完成！"

# 使用说明
cat << 'EOF'

📖 使用说明:
===============================================================================

1. 基本使用:
   ./scripts/deploy.sh

2. 自定义配置:
   DB_PASSWORD="your_password" ./scripts/deploy.sh

3. 完整配置:
   DB_PASSWORD="your_password" DB_HOST="your_host" DB_USER="your_user" ./scripts/deploy.sh

4. 手动启动服务器:
   cd backend
   export PGPASSWORD="your_password"
   go run yunai_core_server.go user_identity_service.go global_prompt_service.go ai_invitation_service.go story_trigger_service.go world_setting_service.go voice_call_service.go

5. 运行功能测试:
   go run test_all_systems.go

===============================================================================
EOF
