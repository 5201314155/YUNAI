@echo off
REM ===============================================================================
REM YUNAI 数据库一键部署脚本 (Windows)
REM ===============================================================================
REM 版本: v2.0
REM 创建时间: 2025-08-29
REM 作者: YUNAI开发团队
REM 说明: 一键创建YUNAI数据库和所有表结构
REM ===============================================================================

echo.
echo ===============================================================================
echo 🚀 YUNAI 数据库一键部署脚本
echo ===============================================================================
echo.

REM 设置数据库连接参数
set DB_HOST=localhost
set DB_PORT=5432
set DB_USER=postgres
set DB_PASSWORD=5201314hdz
set DB_NAME=yunai

REM 设置环境变量避免密码提示
set PGPASSWORD=%DB_PASSWORD%

echo 📋 部署配置:
echo    数据库主机: %DB_HOST%:%DB_PORT%
echo    数据库用户: %DB_USER%
echo    数据库名称: %DB_NAME%
echo.

REM 检查PostgreSQL是否运行
echo 🔍 检查PostgreSQL服务状态...
pg_isready -h %DB_HOST% -p %DB_PORT% -U %DB_USER% >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ PostgreSQL服务未运行或连接失败！
    echo    请确保PostgreSQL服务已启动并且连接参数正确。
    pause
    exit /b 1
)
echo ✅ PostgreSQL服务运行正常

REM 检查数据库是否存在，如果存在则询问是否删除
echo.
echo 🔍 检查数据库是否存在...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -lqt | findstr /C:"%DB_NAME%" >nul 2>&1
if %errorlevel% equ 0 (
    echo ⚠️  数据库 '%DB_NAME%' 已存在！
    set /p choice="是否删除现有数据库并重新创建？(y/N): "
    if /i "%choice%" neq "y" (
        echo 🚫 取消部署
        pause
        exit /b 0
    )
    
    echo 🗑️  删除现有数据库...
    REM 终止所有连接到该数据库的会话
    psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%DB_NAME%' AND pid <> pg_backend_pid();" >nul 2>&1
    
    REM 删除数据库
    dropdb -h %DB_HOST% -p %DB_PORT% -U %DB_USER% %DB_NAME%
    if %errorlevel% neq 0 (
        echo ❌ 删除数据库失败！
        pause
        exit /b 1
    )
    echo ✅ 数据库删除成功
)

REM 创建新数据库
echo.
echo 🏗️  创建数据库 '%DB_NAME%'...
createdb -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -E UTF8 %DB_NAME%
if %errorlevel% neq 0 (
    echo ❌ 创建数据库失败！
    pause
    exit /b 1
)
echo ✅ 数据库创建成功

REM 执行建表脚本
echo.
echo 📊 执行建表脚本...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -f yunai_complete_database_setup.sql
if %errorlevel% neq 0 (
    echo ❌ 建表脚本执行失败！
    pause
    exit /b 1
)

REM 验证表创建
echo.
echo 🔍 验证表创建结果...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "SELECT COUNT(*) as table_count FROM information_schema.tables WHERE table_schema = 'public';"

echo.
echo ===============================================================================
echo 🎉 YUNAI 数据库部署完成！
echo ===============================================================================
echo.
echo 📊 数据库信息:
echo    数据库名称: %DB_NAME%
echo    连接地址: %DB_HOST%:%DB_PORT%
echo    用户名: %DB_USER%
echo.
echo 🚀 下一步:
echo    1. 启动YUNAI后端服务器: go run yunai_core_server.go
echo    2. 访问API文档: http://localhost:8081/swagger/index.html
echo    3. 开始功能测试
echo.

pause
