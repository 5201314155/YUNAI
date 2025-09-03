#!/bin/bash
# ===============================================================================
# YUNAI 数据库一键部署脚本 (Linux/Mac)
# ===============================================================================
# 版本: v2.0
# 创建时间: 2025-08-29
# 作者: YUNAI开发团队
# 说明: 一键创建YUNAI数据库和所有表结构
# ===============================================================================

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 数据库连接参数
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-5201314hdz}
DB_NAME=${DB_NAME:-yunai}

# 设置环境变量避免密码提示
export PGPASSWORD=$DB_PASSWORD

echo
echo "==============================================================================="
echo -e "${BLUE}🚀 YUNAI 数据库一键部署脚本${NC}"
echo "==============================================================================="
echo

echo -e "${BLUE}📋 部署配置:${NC}"
echo "   数据库主机: $DB_HOST:$DB_PORT"
echo "   数据库用户: $DB_USER"
echo "   数据库名称: $DB_NAME"
echo

# 检查PostgreSQL是否运行
echo -e "${BLUE}🔍 检查PostgreSQL服务状态...${NC}"
if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER >/dev/null 2>&1; then
    echo -e "${RED}❌ PostgreSQL服务未运行或连接失败！${NC}"
    echo "   请确保PostgreSQL服务已启动并且连接参数正确。"
    exit 1
fi
echo -e "${GREEN}✅ PostgreSQL服务运行正常${NC}"

# 检查数据库是否存在
echo
echo -e "${BLUE}🔍 检查数据库是否存在...${NC}"
if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
    echo -e "${YELLOW}⚠️  数据库 '$DB_NAME' 已存在！${NC}"
    read -p "是否删除现有数据库并重新创建？(y/N): " choice
    case "$choice" in 
        y|Y ) 
            echo -e "${YELLOW}🗑️  删除现有数据库...${NC}"
            # 终止所有连接到该数据库的会话
            psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();" >/dev/null 2>&1
            
            # 删除数据库
            if ! dropdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME; then
                echo -e "${RED}❌ 删除数据库失败！${NC}"
                exit 1
            fi
            echo -e "${GREEN}✅ 数据库删除成功${NC}"
            ;;
        * ) 
            echo -e "${RED}🚫 取消部署${NC}"
            exit 0
            ;;
    esac
fi

# 创建新数据库
echo
echo -e "${BLUE}🏗️  创建数据库 '$DB_NAME'...${NC}"
if ! createdb -h $DB_HOST -p $DB_PORT -U $DB_USER -E UTF8 $DB_NAME; then
    echo -e "${RED}❌ 创建数据库失败！${NC}"
    exit 1
fi
echo -e "${GREEN}✅ 数据库创建成功${NC}"

# 执行建表脚本
echo
echo -e "${BLUE}📊 执行建表脚本...${NC}"
if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f yunai_complete_database_setup.sql; then
    echo -e "${RED}❌ 建表脚本执行失败！${NC}"
    exit 1
fi

# 验证表创建
echo
echo -e "${BLUE}🔍 验证表创建结果...${NC}"
table_count=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")
echo -e "${GREEN}✅ 成功创建 $table_count 个数据表${NC}"

echo
echo "==============================================================================="
echo -e "${GREEN}🎉 YUNAI 数据库部署完成！${NC}"
echo "==============================================================================="
echo
echo -e "${BLUE}📊 数据库信息:${NC}"
echo "   数据库名称: $DB_NAME"
echo "   连接地址: $DB_HOST:$DB_PORT"
echo "   用户名: $DB_USER"
echo
echo -e "${BLUE}🚀 下一步:${NC}"
echo "   1. 启动YUNAI后端服务器: go run yunai_core_server.go"
echo "   2. 访问API文档: http://localhost:8081/swagger/index.html"
echo "   3. 开始功能测试"
echo

# 清理环境变量
unset PGPASSWORD
