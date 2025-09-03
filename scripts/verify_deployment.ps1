# YUNAI部署验证脚本

param(
    [string]$ServerURL = "http://localhost:8081"
)

Write-Host "🔍 YUNAI部署验证" -ForegroundColor Green
Write-Host "===============================================================================" -ForegroundColor Cyan

# 验证结果
$verificationResults = @{}

# 1. 服务器可访问性验证
Write-Host "`n1. 🌐 服务器可访问性验证..." -ForegroundColor Yellow

try {
    $healthResponse = Invoke-WebRequest -Uri "$ServerURL/health" -Method GET -TimeoutSec 10
    if ($healthResponse.StatusCode -eq 200) {
        $healthData = $healthResponse.Content | ConvertFrom-Json
        Write-Host "   ✅ 服务器运行正常" -ForegroundColor Green
        Write-Host "   📊 服务器状态: $($healthData.status)" -ForegroundColor Gray
        Write-Host "   🕐 运行时间: $($healthData.uptime)" -ForegroundColor Gray
        $verificationResults["服务器可访问性"] = $true
    } else {
        Write-Host "   ❌ 服务器响应异常" -ForegroundColor Red
        $verificationResults["服务器可访问性"] = $false
    }
} catch {
    Write-Host "   ❌ 服务器无法访问: $($_.Exception.Message)" -ForegroundColor Red
    $verificationResults["服务器可访问性"] = $false
}

# 2. 核心API端点验证
Write-Host "`n2. 🎯 核心API端点验证..." -ForegroundColor Yellow

$coreEndpoints = @{
    "用户注册" = "/api/v1/auth/register"
    "角色管理" = "/api/v1/characters"
    "身份识别" = "/api/v1/identity/extract"
    "全局提示词" = "/api/v1/global-prompt/generate"
    "朋友圈" = "/api/v1/moments"
    "钱包系统" = "/api/v1/wallet/balance"
    "世界观设定" = "/api/v1/world-setting/create"
    "语音通话" = "/api/v1/voice-call/start"
}

$workingEndpoints = 0
foreach ($endpoint in $coreEndpoints.GetEnumerator()) {
    try {
        $response = Invoke-WebRequest -Uri "$ServerURL$($endpoint.Value)" -Method HEAD -TimeoutSec 5 -ErrorAction SilentlyContinue
        if ($response.StatusCode -lt 500) {
            Write-Host "   ✅ $($endpoint.Key): 可访问" -ForegroundColor Green
            $workingEndpoints++
        } else {
            Write-Host "   ❌ $($endpoint.Key): 服务器错误" -ForegroundColor Red
        }
    } catch {
        Write-Host "   ❌ $($endpoint.Key): 无法访问" -ForegroundColor Red
    }
}

$endpointSuccessRate = [math]::Round($workingEndpoints / $coreEndpoints.Count * 100, 1)
Write-Host "   📊 API端点可用率: $workingEndpoints/$($coreEndpoints.Count) ($endpointSuccessRate%)" -ForegroundColor Blue

$verificationResults["API端点"] = $workingEndpoints -ge ($coreEndpoints.Count * 0.8)

# 3. 数据库连接验证
Write-Host "`n3. 💾 数据库连接验证..." -ForegroundColor Yellow

$env:PGPASSWORD = "5201314hdz"
try {
    $dbStatus = psql -h localhost -U postgres -d yunai -c "SELECT 'connected' as status;" -t 2>$null
    if ($LASTEXITCODE -eq 0 -and $dbStatus -match "connected") {
        Write-Host "   ✅ 数据库连接正常" -ForegroundColor Green
        $verificationResults["数据库连接"] = $true
        
        # 检查核心表
        $coreTableCount = psql -h localhost -U postgres -d yunai -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name LIKE 'core_%';" -t 2>$null
        Write-Host "   📊 核心表数量: $coreTableCount" -ForegroundColor Gray
        
    } else {
        Write-Host "   ❌ 数据库连接失败" -ForegroundColor Red
        $verificationResults["数据库连接"] = $false
    }
} catch {
    Write-Host "   ❌ 数据库连接异常: $($_.Exception.Message)" -ForegroundColor Red
    $verificationResults["数据库连接"] = $false
}

# 4. 功能完整性验证
Write-Host "`n4. 🧪 功能完整性验证..." -ForegroundColor Yellow

if (Test-Path "test_all_systems.go") {
    Write-Host "   🔄 运行功能测试..." -ForegroundColor Blue
    
    try {
        $testOutput = go run test_all_systems.go 2>&1
        
        # 分析测试输出
        if ($testOutput -match "(\d+)/(\d+) 功能测试通过") {
            $passedTests = $matches[1]
            $totalTests = $matches[2]
            $successRate = [math]::Round([int]$passedTests / [int]$totalTests * 100, 1)
            
            Write-Host "   📊 功能测试结果: $passedTests/$totalTests 通过 ($successRate%)" -ForegroundColor Blue
            
            if ($successRate -ge 80) {
                Write-Host "   ✅ 功能测试通过" -ForegroundColor Green
                $verificationResults["功能完整性"] = $true
            } else {
                Write-Host "   ⚠️ 功能测试部分通过" -ForegroundColor Yellow
                $verificationResults["功能完整性"] = $false
            }
        } else {
            Write-Host "   ⚠️ 无法解析测试结果" -ForegroundColor Yellow
            $verificationResults["功能完整性"] = $false
        }
    } catch {
        Write-Host "   ❌ 功能测试执行失败: $($_.Exception.Message)" -ForegroundColor Red
        $verificationResults["功能完整性"] = $false
    }
} else {
    Write-Host "   ⚠️ 测试文件不存在，跳过功能验证" -ForegroundColor Yellow
    $verificationResults["功能完整性"] = $false
}

# 5. 系统资源验证
Write-Host "`n5. 💻 系统资源验证..." -ForegroundColor Yellow

# 检查Go进程
$goProcesses = Get-Process | Where-Object {$_.ProcessName -eq "go"}
if ($goProcesses) {
    Write-Host "   ✅ YUNAI服务器进程运行中 (PID: $($goProcesses[0].Id))" -ForegroundColor Green
    
    # 检查内存使用
    $memoryMB = [math]::Round($goProcesses[0].WorkingSet64 / 1MB, 1)
    Write-Host "   📊 内存使用: $memoryMB MB" -ForegroundColor Gray
    
    $verificationResults["系统资源"] = $true
} else {
    Write-Host "   ❌ 未发现YUNAI服务器进程" -ForegroundColor Red
    $verificationResults["系统资源"] = $false
}

# 输出验证结果
Write-Host "`n===============================================================================" -ForegroundColor Cyan
Write-Host "📋 YUNAI部署验证结果" -ForegroundColor Green
Write-Host "===============================================================================" -ForegroundColor Cyan

$passedVerifications = 0
$totalVerifications = $verificationResults.Count

foreach ($verification in $verificationResults.GetEnumerator()) {
    $status = if ($verification.Value) { "✅ 通过" } else { "❌ 失败" }
    $color = if ($verification.Value) { "Green" } else { "Red" }
    
    Write-Host ("{0,-15}: {1}" -f $verification.Key, $status) -ForegroundColor $color
    
    if ($verification.Value) {
        $passedVerifications++
    }
}

$overallSuccessRate = [math]::Round($passedVerifications / $totalVerifications * 100, 1)
Write-Host "`n📊 验证总结: $passedVerifications/$totalVerifications 项验证通过 ($overallSuccessRate%)" -ForegroundColor White

# 最终评估
if ($passedVerifications -eq $totalVerifications) {
    Write-Host "`n🎉 YUNAI系统部署验证完全通过！" -ForegroundColor Green
    Write-Host "系统已准备好投入使用。" -ForegroundColor Green
} elseif ($passedVerifications -ge $totalVerifications * 0.8) {
    Write-Host "`n⚠️ YUNAI系统基本部署成功！" -ForegroundColor Yellow
    Write-Host "系统可以使用，但建议修复剩余问题。" -ForegroundColor Yellow
} else {
    Write-Host "`n🚨 YUNAI系统部署存在严重问题！" -ForegroundColor Red
    Write-Host "请检查并修复问题后重新验证。" -ForegroundColor Red
}

# 提供下一步建议
Write-Host "`n💡 下一步建议:" -ForegroundColor Yellow

if ($verificationResults["服务器可访问性"] -eq $false) {
    Write-Host "   🚀 启动服务器: .\scripts\start_server.ps1" -ForegroundColor Cyan
}

if ($verificationResults["数据库连接"] -eq $false) {
    Write-Host "   💾 检查数据库配置和连接" -ForegroundColor Cyan
}

if ($verificationResults["功能完整性"] -eq $false) {
    Write-Host "   🧪 运行详细测试: go run test_all_systems.go" -ForegroundColor Cyan
}

Write-Host "`n✨ 验证脚本执行完成！" -ForegroundColor Green
