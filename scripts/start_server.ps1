# YUNAI服务器快速启动脚本 (PowerShell)

param(
    [string]$DBPassword = "5201314hdz",
    [int]$Port = 8081,
    [switch]$Test = $false
)

Write-Host "🚀 启动YUNAI服务器" -ForegroundColor Green
Write-Host "===============================================================================" -ForegroundColor Cyan

# 设置环境变量
$env:PGPASSWORD = $DBPassword

# 检查端口是否被占用
$portInUse = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue
if ($portInUse) {
    Write-Host "⚠️ 端口 $Port 已被占用，正在尝试停止现有服务..." -ForegroundColor Yellow
    
    # 尝试停止占用端口的进程
    $processes = Get-Process | Where-Object {$_.ProcessName -eq "go"}
    if ($processes) {
        $processes | Stop-Process -Force
        Write-Host "✅ 已停止现有Go进程" -ForegroundColor Green
        Start-Sleep -Seconds 2
    }
}

# 切换到backend目录
if (Test-Path "backend") {
    Set-Location "backend"
} elseif (Test-Path "../backend") {
    Set-Location "../backend"
} else {
    Write-Host "❌ 找不到backend目录" -ForegroundColor Red
    exit 1
}

# 检查必要文件
$requiredFiles = @(
    "yunai_core_server.go",
    "user_identity_service.go", 
    "global_prompt_service.go",
    "ai_invitation_service.go",
    "story_trigger_service.go",
    "world_setting_service.go",
    "voice_call_service.go"
)

$missingFiles = @()
foreach ($file in $requiredFiles) {
    if (!(Test-Path $file)) {
        $missingFiles += $file
    }
}

if ($missingFiles.Count -gt 0) {
    Write-Host "❌ 缺少必要文件:" -ForegroundColor Red
    $missingFiles | ForEach-Object { Write-Host "   - $_" -ForegroundColor Red }
    exit 1
}

Write-Host "✅ 所有必要文件检查通过" -ForegroundColor Green

# 启动服务器
Write-Host "`n🔄 启动YUNAI核心服务器..." -ForegroundColor Blue
Write-Host "端口: $Port" -ForegroundColor Gray
Write-Host "数据库: yunai@localhost" -ForegroundColor Gray

if ($Test) {
    Write-Host "🧪 测试模式：服务器将在测试完成后自动停止" -ForegroundColor Yellow
    
    # 后台启动服务器
    $serverJob = Start-Job -ScriptBlock {
        param($BackendPath, $Password)
        Set-Location $BackendPath
        $env:PGPASSWORD = $Password
        go run yunai_core_server.go user_identity_service.go global_prompt_service.go ai_invitation_service.go story_trigger_service.go world_setting_service.go voice_call_service.go
    } -ArgumentList (Get-Location).Path, $DBPassword
    
    # 等待服务器启动
    Start-Sleep -Seconds 8
    
    # 检查服务器状态
    try {
        $healthCheck = Invoke-WebRequest -Uri "http://localhost:$Port/health" -Method GET -TimeoutSec 10
        if ($healthCheck.StatusCode -eq 200) {
            Write-Host "✅ 服务器启动成功，开始功能测试..." -ForegroundColor Green
            
            # 运行功能测试
            Set-Location ".."
            if (Test-Path "test_all_systems.go") {
                go run test_all_systems.go
            } else {
                Write-Host "⚠️ 测试文件不存在" -ForegroundColor Yellow
            }
        }
    } catch {
        Write-Host "❌ 服务器启动失败: $($_.Exception.Message)" -ForegroundColor Red
    } finally {
        # 停止服务器
        Write-Host "`n🛑 停止测试服务器..." -ForegroundColor Yellow
        Stop-Job $serverJob -ErrorAction SilentlyContinue
        Remove-Job $serverJob -ErrorAction SilentlyContinue
    }
    
} else {
    Write-Host "🔄 正常模式：服务器将持续运行" -ForegroundColor Blue
    Write-Host "按 Ctrl+C 停止服务器" -ForegroundColor Gray
    Write-Host ""
    
    # 直接启动服务器
    go run yunai_core_server.go user_identity_service.go global_prompt_service.go ai_invitation_service.go story_trigger_service.go world_setting_service.go voice_call_service.go
}

Write-Host "`n✨ 启动脚本执行完成！" -ForegroundColor Green
