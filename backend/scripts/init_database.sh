#!/bin/bash

# YUNAI 数据库初始化脚本
# 作者: 小云
# 版本: v1.0
# 日期: 2025-08-26

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 配置变量
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-5201314hdz}
DB_NAME=${DB_NAME:-yunai_dev}
ADMIN_DB=${ADMIN_DB:-postgres}

# 检查 PostgreSQL 是否运行
check_postgres() {
    log_info "检查 PostgreSQL 服务状态..."
    if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER > /dev/null 2>&1; then
        log_error "PostgreSQL 服务未运行或连接失败"
        log_info "请确保 PostgreSQL 已启动并且连接参数正确"
        exit 1
    fi
    log_success "PostgreSQL 服务运行正常"
}

# 创建数据库
create_database() {
    log_info "创建数据库 $DB_NAME..."
    
    # 检查数据库是否已存在
    DB_EXISTS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $ADMIN_DB -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'")
    
    if [ "$DB_EXISTS" = "1" ]; then
        log_warning "数据库 $DB_NAME 已存在"
    else
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $ADMIN_DB -c "CREATE DATABASE $DB_NAME;"
        log_success "数据库 $DB_NAME 创建成功"
    fi
}

# 创建扩展
create_extensions() {
    log_info "创建数据库扩展..."
    
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << EOF
-- 创建 UUID 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建 pgcrypto 扩展（用于密码哈希）
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 创建 pg_trgm 扩展（用于模糊搜索）
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 创建 btree_gin 扩展（用于复合索引）
CREATE EXTENSION IF NOT EXISTS "btree_gin";
EOF
    
    log_success "数据库扩展创建完成"
}

# 运行迁移
run_migrations() {
    log_info "运行数据库迁移..."
    
    # 检查迁移文件是否存在
    if [ ! -d "./migrations" ]; then
        log_error "迁移目录不存在: ./migrations"
        exit 1
    fi
    
    # 按顺序执行迁移文件
    for migration_file in ./migrations/*.up.sql; do
        if [ -f "$migration_file" ]; then
            log_info "执行迁移: $(basename $migration_file)"
            PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$migration_file"
        fi
    done
    
    log_success "数据库迁移完成"
}

# 创建数据库用户（如果需要）
create_app_user() {
    local APP_USER=${APP_USER:-yunai_app}
    local APP_PASSWORD=${APP_PASSWORD:-yunai_app_password}
    
    log_info "创建应用数据库用户..."
    
    # 检查用户是否已存在
    USER_EXISTS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $ADMIN_DB -tAc "SELECT 1 FROM pg_roles WHERE rolname='$APP_USER'")
    
    if [ "$USER_EXISTS" = "1" ]; then
        log_warning "用户 $APP_USER 已存在"
    else
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $ADMIN_DB << EOF
CREATE USER $APP_USER WITH PASSWORD '$APP_PASSWORD';
GRANT CONNECT ON DATABASE $DB_NAME TO $APP_USER;
GRANT USAGE ON SCHEMA public TO $APP_USER;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO $APP_USER;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO $APP_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO $APP_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO $APP_USER;
EOF
        log_success "应用用户 $APP_USER 创建成功"
    fi
}

# 验证数据库
verify_database() {
    log_info "验证数据库结构..."
    
    # 检查关键表是否存在
    TABLES=("users" "wallets" "gift_cards" "feature_flags" "ai_models")
    
    for table in "${TABLES[@]}"; do
        TABLE_EXISTS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT 1 FROM information_schema.tables WHERE table_name='$table'")
        if [ "$TABLE_EXISTS" = "1" ]; then
            log_success "表 $table 存在"
        else
            log_error "表 $table 不存在"
            exit 1
        fi
    done
    
    # 检查默认管理员用户
    ADMIN_EXISTS=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "SELECT 1 FROM users WHERE username='admin'")
    if [ "$ADMIN_EXISTS" = "1" ]; then
        log_success "默认管理员用户存在"
    else
        log_error "默认管理员用户不存在"
        exit 1
    fi
    
    log_success "数据库验证通过"
}

# 显示数据库信息
show_database_info() {
    log_info "数据库信息:"
    echo "  主机: $DB_HOST:$DB_PORT"
    echo "  数据库: $DB_NAME"
    echo "  用户: $DB_USER"
    echo ""
    
    log_info "默认管理员账号:"
    echo "  用户名: admin"
    echo "  邮箱: admin@yunai.com"
    echo "  密码: admin123"
    echo ""
    
    log_info "测试卡密:"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT card_code, card_type, description FROM gift_cards WHERE card_code LIKE '%TEST%';"
}

# 主函数
main() {
    log_info "开始初始化 YUNAI 数据库..."
    echo "========================================"
    
    check_postgres
    create_database
    create_extensions
    run_migrations
    
    # 如果设置了创建应用用户的环境变量
    if [ "$CREATE_APP_USER" = "true" ]; then
        create_app_user
    fi
    
    verify_database
    show_database_info
    
    echo "========================================"
    log_success "YUNAI 数据库初始化完成！"
}

# 帮助信息
show_help() {
    echo "YUNAI 数据库初始化脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "环境变量:"
    echo "  DB_HOST          数据库主机 (默认: localhost)"
    echo "  DB_PORT          数据库端口 (默认: 5432)"
    echo "  DB_USER          数据库用户 (默认: postgres)"
    echo "  DB_PASSWORD      数据库密码 (默认: password)"
    echo "  DB_NAME          数据库名称 (默认: yunai_dev)"
    echo "  CREATE_APP_USER  是否创建应用用户 (默认: false)"
    echo "  APP_USER         应用用户名 (默认: yunai_app)"
    echo "  APP_PASSWORD     应用用户密码 (默认: yunai_app_password)"
    echo ""
    echo "选项:"
    echo "  -h, --help       显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0"
    echo "  DB_NAME=yunai_prod CREATE_APP_USER=true $0"
}

# 解析命令行参数
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    *)
        main
        ;;
esac
