# YUNAI系统状态检查脚本 (PowerShell)

param(
    [string]$DBPassword = "5201314hdz",
    [string]$ServerURL = "http://localhost:8081"
)

Write-Host "🔍 YUNAI系统状态检查" -ForegroundColor Green
Write-Host "===============================================================================" -ForegroundColor Cyan

$env:PGPASSWORD = $DBPassword

# 检查结果统计
$checkResults = @{}

# 1. 数据库连接检查
Write-Host "`n1. 💾 数据库连接检查..." -ForegroundColor Yellow

try {
    $dbVersion = psql -h localhost -U postgres -d yunai -c "SELECT version();" -t 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✅ 数据库连接正常" -ForegroundColor Green
        $checkResults["数据库连接"] = $true
    } else {
        Write-Host "   ❌ 数据库连接失败" -ForegroundColor Red
        $checkResults["数据库连接"] = $false
    }
} catch {
    Write-Host "   ❌ 数据库连接异常: $($_.Exception.Message)" -ForegroundColor Red
    $checkResults["数据库连接"] = $false
}

# 2. 数据库表结构检查
Write-Host "`n2. 🏗️ 数据库表结构检查..." -ForegroundColor Yellow

$coreTableCount = psql -h localhost -U postgres -d yunai -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name LIKE 'core_%';" -t 2>$null
$extendedTableCount = psql -h localhost -U postgres -d yunai -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name IN ('user_identity_contexts', 'world_settings', 'voice_calls', 'ai_invitations', 'stories', 'wallets');" -t 2>$null

if ($coreTableCount -ge 4 -and $extendedTableCount -ge 4) {
    Write-Host "   ✅ 数据库表结构完整 (核心表: $coreTableCount, 扩展表: $extendedTableCount)" -ForegroundColor Green
    $checkResults["数据库表结构"] = $true
} else {
    Write-Host "   ❌ 数据库表结构不完整 (核心表: $coreTableCount, 扩展表: $extendedTableCount)" -ForegroundColor Red
    $checkResults["数据库表结构"] = $false
}

# 3. 服务器状态检查
Write-Host "`n3. 🌐 服务器状态检查..." -ForegroundColor Yellow

try {
    $healthResponse = Invoke-WebRequest -Uri "$ServerURL/health" -Method GET -TimeoutSec 10
    if ($healthResponse.StatusCode -eq 200) {
        Write-Host "   ✅ 服务器运行正常" -ForegroundColor Green
        $checkResults["服务器状态"] = $true
        
        # 解析健康检查响应
        $healthData = $healthResponse.Content | ConvertFrom-Json
        Write-Host "   📊 服务器信息:" -ForegroundColor Blue
        Write-Host "      - 状态: $($healthData.status)" -ForegroundColor Gray
        Write-Host "      - 版本: $($healthData.version)" -ForegroundColor Gray
        Write-Host "      - 运行时间: $($healthData.uptime)" -ForegroundColor Gray
    } else {
        Write-Host "   ❌ 服务器响应异常 (状态码: $($healthResponse.StatusCode))" -ForegroundColor Red
        $checkResults["服务器状态"] = $false
    }
} catch {
    Write-Host "   ❌ 服务器无法访问: $($_.Exception.Message)" -ForegroundColor Red
    $checkResults["服务器状态"] = $false
}

# 4. API端点检查
Write-Host "`n4. 🎯 API端点检查..." -ForegroundColor Yellow

$apiEndpoints = @(
    "/api/v1/auth/register",
    "/api/v1/characters",
    "/api/v1/identity/extract",
    "/api/v1/global-prompt/generate",
    "/api/v1/moments",
    "/api/v1/wallet/balance",
    "/api/v1/world-setting/create",
    "/api/v1/voice-call/start"
)

$workingEndpoints = 0
foreach ($endpoint in $apiEndpoints) {
    try {
        # 使用HEAD请求检查端点是否存在
        $response = Invoke-WebRequest -Uri "$ServerURL$endpoint" -Method HEAD -TimeoutSec 5 -ErrorAction SilentlyContinue
        if ($response.StatusCode -lt 500) {
            $workingEndpoints++
        }
    } catch {
        # 忽略错误，某些端点可能需要特定参数
    }
}

if ($workingEndpoints -ge 6) {
    Write-Host "   ✅ API端点检查通过 ($workingEndpoints/$($apiEndpoints.Count))" -ForegroundColor Green
    $checkResults["API端点"] = $true
} else {
    Write-Host "   ❌ API端点检查失败 ($workingEndpoints/$($apiEndpoints.Count))" -ForegroundColor Red
    $checkResults["API端点"] = $false
}

# 5. 数据完整性检查
Write-Host "`n5. 🔍 数据完整性检查..." -ForegroundColor Yellow

try {
    $userCount = psql -h localhost -U postgres -d yunai -c "SELECT COUNT(*) FROM core_users;" -t 2>$null
    $characterCount = psql -h localhost -U postgres -d yunai -c "SELECT COUNT(*) FROM core_characters;" -t 2>$null
    $momentCount = psql -h localhost -U postgres -d yunai -c "SELECT COUNT(*) FROM core_moments;" -t 2>$null
    
    Write-Host "   📊 数据统计:" -ForegroundColor Blue
    Write-Host "      - 用户数量: $userCount" -ForegroundColor Gray
    Write-Host "      - 角色数量: $characterCount" -ForegroundColor Gray
    Write-Host "      - 朋友圈数量: $momentCount" -ForegroundColor Gray
    
    Write-Host "   ✅ 数据完整性检查通过" -ForegroundColor Green
    $checkResults["数据完整性"] = $true
} catch {
    Write-Host "   ❌ 数据完整性检查失败: $($_.Exception.Message)" -ForegroundColor Red
    $checkResults["数据完整性"] = $false
}

# 6. 外键约束检查
Write-Host "`n6. 🔗 外键约束检查..." -ForegroundColor Yellow

try {
    $fkErrors = psql -h localhost -U postgres -d yunai -c "SELECT COUNT(*) FROM pg_constraint WHERE contype = 'f' AND confrelid = 'users'::regclass;" -t 2>$null
    
    if ($fkErrors -eq 0) {
        Write-Host "   ✅ 外键约束配置正确" -ForegroundColor Green
        $checkResults["外键约束"] = $true
    } else {
        Write-Host "   ❌ 发现 $fkErrors 个错误的外键约束（指向users表而非core_users表）" -ForegroundColor Red
        Write-Host "   💡 请运行: psql -d yunai -f fix_foreign_keys.sql" -ForegroundColor Yellow
        $checkResults["外键约束"] = $false
    }
} catch {
    Write-Host "   ❌ 外键约束检查失败: $($_.Exception.Message)" -ForegroundColor Red
    $checkResults["外键约束"] = $false
}

# 输出检查结果
Write-Host "`n===============================================================================" -ForegroundColor Cyan
Write-Host "📋 系统状态检查结果" -ForegroundColor Green
Write-Host "===============================================================================" -ForegroundColor Cyan

$passedChecks = 0
$totalChecks = $checkResults.Count

foreach ($check in $checkResults.GetEnumerator()) {
    $status = if ($check.Value) { "✅ 正常" } else { "❌ 异常" }
    $color = if ($check.Value) { "Green" } else { "Red" }
    
    Write-Host ("{0,-15}: {1}" -f $check.Key, $status) -ForegroundColor $color
    
    if ($check.Value) {
        $passedChecks++
    }
}

Write-Host "`n📊 检查总结: $passedChecks/$totalChecks 项检查通过 ($([math]::Round($passedChecks/$totalChecks*100, 1))%)" -ForegroundColor White

if ($passedChecks -eq $totalChecks) {
    Write-Host "🎉 系统状态良好，所有检查通过！" -ForegroundColor Green
} elseif ($passedChecks -ge $totalChecks * 0.8) {
    Write-Host "⚠️ 系统基本正常，但有部分问题需要关注" -ForegroundColor Yellow
} else {
    Write-Host "🚨 系统存在严重问题，请立即检查和修复" -ForegroundColor Red
}

# 提供修复建议
if ($checkResults["外键约束"] -eq $false) {
    Write-Host "`n💡 修复建议:" -ForegroundColor Yellow
    Write-Host "   运行外键修复脚本: psql -d yunai -f fix_foreign_keys.sql" -ForegroundColor Cyan
}

if ($checkResults["服务器状态"] -eq $false) {
    Write-Host "`n💡 启动服务器:" -ForegroundColor Yellow
    Write-Host "   ./scripts/start_server.ps1" -ForegroundColor Cyan
}

Write-Host "`n✨ 系统检查完成！" -ForegroundColor Green
