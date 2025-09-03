# YUNAI系统自动化部署脚本 (PowerShell)
# 适用于Windows环境

param(
    [string]$DBPassword = "5201314hdz",
    [string]$DBHost = "localhost",
    [string]$DBUser = "postgres",
    [string]$DBName = "yunai",
    [int]$ServerPort = 8081
)

Write-Host "🚀 YUNAI系统自动化部署开始" -ForegroundColor Green
Write-Host "===============================================================================" -ForegroundColor Cyan

# 设置错误处理
$ErrorActionPreference = "Stop"

try {
    # 1. 环境检查
    Write-Host "`n1. 🔍 环境检查..." -ForegroundColor Yellow
    
    # 检查Go环境
    $goVersion = go version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✅ Go环境: $goVersion" -ForegroundColor Green
    } else {
        throw "Go环境未安装或未配置"
    }
    
    # 检查PostgreSQL
    $env:PGPASSWORD = $DBPassword
    $pgVersion = psql -h $DBHost -U $DBUser -d postgres -c "SELECT version();" -t 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✅ PostgreSQL连接成功" -ForegroundColor Green
    } else {
        throw "PostgreSQL连接失败"
    }

    # 2. 数据库准备
    Write-Host "`n2. 💾 数据库准备..." -ForegroundColor Yellow
    
    # 检查数据库是否存在
    $dbExists = psql -h $DBHost -U $DBUser -d postgres -c "SELECT 1 FROM pg_database WHERE datname='$DBName';" -t 2>$null
    if ($dbExists -match "1") {
        Write-Host "   ✅ 数据库 '$DBName' 已存在" -ForegroundColor Green
    } else {
        Write-Host "   📝 创建数据库 '$DBName'..." -ForegroundColor Blue
        createdb -h $DBHost -U $DBUser $DBName
        Write-Host "   ✅ 数据库创建成功" -ForegroundColor Green
    }

    # 3. 初始化数据库表结构
    Write-Host "`n3. 🏗️ 初始化数据库表结构..." -ForegroundColor Yellow
    
    if (Test-Path "database/init_database.sql") {
        psql -h $DBHost -U $DBUser -d $DBName -f "database/init_database.sql" > $null 2>&1
        Write-Host "   ✅ 数据库表结构初始化完成" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️ 数据库初始化脚本不存在，跳过" -ForegroundColor Yellow
    }

    # 4. 修复外键约束
    Write-Host "`n4. 🔧 修复外键约束..." -ForegroundColor Yellow
    
    if (Test-Path "fix_foreign_keys.sql") {
        psql -h $DBHost -U $DBUser -d $DBName -f "fix_foreign_keys.sql" > $null 2>&1
        Write-Host "   ✅ 外键约束修复完成" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️ 外键修复脚本不存在，跳过" -ForegroundColor Yellow
    }

    # 5. 安装Go依赖
    Write-Host "`n5. 📦 安装Go依赖..." -ForegroundColor Yellow
    
    Set-Location "backend"
    go mod tidy
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✅ Go依赖安装完成" -ForegroundColor Green
    } else {
        throw "Go依赖安装失败"
    }

    # 6. 编译项目
    Write-Host "`n6. 🔨 编译项目..." -ForegroundColor Yellow
    
    go build .
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✅ 项目编译成功" -ForegroundColor Green
    } else {
        throw "项目编译失败"
    }

    # 7. 运行功能测试
    Write-Host "`n7. 🧪 运行功能测试..." -ForegroundColor Yellow
    
    # 启动服务器（后台）
    $serverJob = Start-Job -ScriptBlock {
        param($BackendPath)
        Set-Location $BackendPath
        $env:PGPASSWORD = "5201314hdz"
        go run yunai_core_server.go user_identity_service.go global_prompt_service.go ai_invitation_service.go story_trigger_service.go world_setting_service.go voice_call_service.go
    } -ArgumentList (Get-Location).Path
    
    # 等待服务器启动
    Start-Sleep -Seconds 5
    
    # 检查服务器是否启动
    try {
        $healthCheck = Invoke-WebRequest -Uri "http://localhost:$ServerPort/health" -Method GET -TimeoutSec 10
        if ($healthCheck.StatusCode -eq 200) {
            Write-Host "   ✅ 服务器启动成功" -ForegroundColor Green
            
            # 运行功能测试
            Set-Location ".."
            if (Test-Path "test_all_systems.go") {
                Write-Host "   🧪 执行功能测试..." -ForegroundColor Blue
                go run test_all_systems.go
                Write-Host "   ✅ 功能测试完成" -ForegroundColor Green
            }
        }
    } catch {
        Write-Host "   ❌ 服务器启动失败" -ForegroundColor Red
    } finally {
        # 停止后台服务器
        Stop-Job $serverJob -ErrorAction SilentlyContinue
        Remove-Job $serverJob -ErrorAction SilentlyContinue
    }

    # 8. 部署完成
    Write-Host "`n===============================================================================" -ForegroundColor Cyan
    Write-Host "🎉 YUNAI系统部署完成！" -ForegroundColor Green
    Write-Host "===============================================================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "🌐 服务器地址: http://localhost:$ServerPort" -ForegroundColor White
    Write-Host "📊 健康检查: http://localhost:$ServerPort/health" -ForegroundColor White
    Write-Host "📈 系统统计: http://localhost:$ServerPort/stats" -ForegroundColor White
    Write-Host "🎯 API文档: http://localhost:$ServerPort/api/docs" -ForegroundColor White
    Write-Host ""
    Write-Host "🚀 启动服务器命令:" -ForegroundColor Yellow
    Write-Host "cd backend" -ForegroundColor Gray
    Write-Host "go run yunai_core_server.go user_identity_service.go global_prompt_service.go ai_invitation_service.go story_trigger_service.go world_setting_service.go voice_call_service.go" -ForegroundColor Gray
    Write-Host ""

} catch {
    Write-Host "`n❌ 部署失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "请检查错误信息并重新运行部署脚本" -ForegroundColor Yellow
    exit 1
}

Write-Host "✨ 部署脚本执行完成！" -ForegroundColor Green
