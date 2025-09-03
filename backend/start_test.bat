@echo off
chcp 65001 >nul

echo 🚀 YUNAI 项目全功能测试启动器
echo ================================================================================
echo.

REM 检查Go环境
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ Go 未安装，请先安装 Go 环境
    pause
    exit /b 1
)

echo ✅ Go 环境检查通过

REM 检查服务器是否运行
echo 🔍 检查服务器状态...
curl -s http://localhost:8092/health >nul 2>nul
if %errorlevel% equ 0 (
    echo ✅ 服务器已运行
    set SERVER_RUNNING=true
) else (
    echo 📡 服务器未运行，正在启动...
    
    REM 启动服务器
    echo 🔧 启动 YUNAI 服务器...
    start /b go run cmd/monolith/main.go
    
    echo ⏳ 等待服务器启动...
    timeout /t 10 /nobreak >nul
    
    REM 再次检查服务器状态
    curl -s http://localhost:8092/health >nul 2>nul
    if %errorlevel% equ 0 (
        echo ✅ 服务器启动成功
        set SERVER_STARTED=true
    ) else (
        echo ❌ 服务器启动失败
        pause
        exit /b 1
    )
)

echo.
echo 🧪 开始执行全功能测试...
echo.

REM 运行测试
go run test_all_features.go

echo.
echo 📋 测试完成！

REM 如果启动了服务器，询问是否关闭
if defined SERVER_STARTED (
    echo.
    set /p CLOSE_SERVER="是否关闭测试服务器? (y/N): "
    if /i "%CLOSE_SERVER%"=="y" (
        echo 🔴 关闭服务器...
        taskkill /f /im go.exe >nul 2>nul
        echo ✅ 服务器已关闭
    ) else (
        echo ℹ️  服务器继续运行
        echo    使用任务管理器手动关闭 go.exe 进程
    )
)

echo.
echo 🎉 测试流程完成！
pause
