# YUNAI 数据库初始化脚本 (PowerShell)
# 作者: 小云
# 版本: v1.0
# 日期: 2025-08-26

param(
    [string]$DBHost = "localhost",
    [int]$DBPort = 5432,
    [string]$DBUser = "postgres", 
    [string]$DBPassword = "5201314hdz",
    [string]$DBName = "yunai_dev",
    [string]$AdminDB = "postgres"
)

# 颜色输出函数
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# 检查 PostgreSQL 是否运行
function Test-PostgreSQL {
    Write-Info "检查 PostgreSQL 服务状态..."
    
    try {
        $env:PGPASSWORD = $DBPassword
        $result = & pg_isready -h $DBHost -p $DBPort -U $DBUser 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Success "PostgreSQL 服务运行正常"
            return $true
        } else {
            Write-Error "PostgreSQL 服务未运行或连接失败: $result"
            return $false
        }
    } catch {
        Write-Error "无法检查 PostgreSQL 状态: $_"
        return $false
    }
}

# 创建数据库
function New-Database {
    Write-Info "创建数据库 $DBName..."
    
    try {
        $env:PGPASSWORD = $DBPassword
        
        # 检查数据库是否已存在
        $checkQuery = "SELECT 1 FROM pg_database WHERE datname='$DBName'"
        $exists = & psql -h $DBHost -p $DBPort -U $DBUser -d $AdminDB -tAc $checkQuery 2>&1
        
        if ($exists -eq "1") {
            Write-Warning "数据库 $DBName 已存在"
        } else {
            $createQuery = "CREATE DATABASE $DBName;"
            & psql -h $DBHost -p $DBPort -U $DBUser -d $AdminDB -c $createQuery 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Success "数据库 $DBName 创建成功"
            } else {
                throw "创建数据库失败"
            }
        }
    } catch {
        Write-Error "创建数据库时出错: $_"
        exit 1
    }
}

# 创建扩展
function New-Extensions {
    Write-Info "创建数据库扩展..."
    
    try {
        $env:PGPASSWORD = $DBPassword
        
        $extensions = @(
            "CREATE EXTENSION IF NOT EXISTS `"uuid-ossp`";",
            "CREATE EXTENSION IF NOT EXISTS `"pgcrypto`";",
            "CREATE EXTENSION IF NOT EXISTS `"pg_trgm`";",
            "CREATE EXTENSION IF NOT EXISTS `"btree_gin`";"
        )
        
        foreach ($ext in $extensions) {
            & psql -h $DBHost -p $DBPort -U $DBUser -d $DBName -c $ext 2>&1 | Out-Null
        }
        
        Write-Success "数据库扩展创建完成"
    } catch {
        Write-Error "创建扩展时出错: $_"
        exit 1
    }
}

# 运行迁移
function Invoke-Migrations {
    Write-Info "运行数据库迁移..."
    
    $migrationsPath = Join-Path $PSScriptRoot "..\migrations"
    
    if (-not (Test-Path $migrationsPath)) {
        Write-Error "迁移目录不存在: $migrationsPath"
        exit 1
    }
    
    try {
        $env:PGPASSWORD = $DBPassword
        
        # 获取所有 .up.sql 文件并排序
        $migrationFiles = Get-ChildItem -Path $migrationsPath -Filter "*.up.sql" | Sort-Object Name
        
        foreach ($file in $migrationFiles) {
            Write-Info "执行迁移: $($file.Name)"
            & psql -h $DBHost -p $DBPort -U $DBUser -d $DBName -f $file.FullName 2>&1
            if ($LASTEXITCODE -ne 0) {
                throw "迁移 $($file.Name) 执行失败"
            }
        }
        
        Write-Success "数据库迁移完成"
    } catch {
        Write-Error "运行迁移时出错: $_"
        exit 1
    }
}

# 验证数据库
function Test-Database {
    Write-Info "验证数据库结构..."
    
    $tables = @("users", "wallets", "gift_cards", "feature_flags", "ai_models")
    
    try {
        $env:PGPASSWORD = $DBPassword
        
        foreach ($table in $tables) {
            $checkQuery = "SELECT 1 FROM information_schema.tables WHERE table_name='$table'"
            $exists = & psql -h $DBHost -p $DBPort -U $DBUser -d $DBName -tAc $checkQuery 2>&1
            
            if ($exists -eq "1") {
                Write-Success "表 $table 存在"
            } else {
                Write-Error "表 $table 不存在"
                exit 1
            }
        }
        
        # 检查默认管理员用户
        $adminQuery = "SELECT 1 FROM users WHERE username='admin'"
        $adminExists = & psql -h $DBHost -p $DBPort -U $DBUser -d $DBName -tAc $adminQuery 2>&1
        
        if ($adminExists -eq "1") {
            Write-Success "默认管理员用户存在"
        } else {
            Write-Error "默认管理员用户不存在"
            exit 1
        }
        
        Write-Success "数据库验证通过"
    } catch {
        Write-Error "验证数据库时出错: $_"
        exit 1
    }
}

# 显示数据库信息
function Show-DatabaseInfo {
    Write-Info "数据库信息:"
    Write-Host "  主机: $DBHost`:$DBPort" -ForegroundColor Cyan
    Write-Host "  数据库: $DBName" -ForegroundColor Cyan
    Write-Host "  用户: $DBUser" -ForegroundColor Cyan
    Write-Host ""
    
    Write-Info "默认管理员账号:"
    Write-Host "  用户名: admin" -ForegroundColor Cyan
    Write-Host "  邮箱: admin@yunai.com" -ForegroundColor Cyan
    Write-Host "  密码: admin123" -ForegroundColor Cyan
    Write-Host ""
    
    Write-Info "测试卡密:"
    try {
        $env:PGPASSWORD = $DBPassword
        & psql -h $DBHost -p $DBPort -U $DBUser -d $DBName -c "SELECT card_code, card_type, description FROM gift_cards WHERE card_code LIKE '%TEST%' OR card_code LIKE '668899%';" 2>&1
    } catch {
        Write-Warning "无法显示测试卡密"
    }
}

# 主函数
function Main {
    Write-Info "开始初始化 YUNAI 数据库..."
    Write-Host "========================================" -ForegroundColor Magenta
    
    # 检查 PostgreSQL
    if (-not (Test-PostgreSQL)) {
        Write-Error "请确保 PostgreSQL 已安装并正在运行"
        exit 1
    }
    
    # 创建数据库
    New-Database
    
    # 创建扩展
    New-Extensions
    
    # 运行迁移
    Invoke-Migrations
    
    # 验证数据库
    Test-Database
    
    # 显示信息
    Show-DatabaseInfo
    
    Write-Host "========================================" -ForegroundColor Magenta
    Write-Success "YUNAI 数据库初始化完成！"
}

# 帮助信息
function Show-Help {
    Write-Host "YUNAI 数据库初始化脚本 (PowerShell)" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "用法: .\init_database.ps1 [参数]" -ForegroundColor White
    Write-Host ""
    Write-Host "参数:" -ForegroundColor White
    Write-Host "  -DBHost        数据库主机 (默认: localhost)" -ForegroundColor Gray
    Write-Host "  -DBPort        数据库端口 (默认: 5432)" -ForegroundColor Gray
    Write-Host "  -DBUser        数据库用户 (默认: postgres)" -ForegroundColor Gray
    Write-Host "  -DBPassword    数据库密码 (默认: 5201314hdz)" -ForegroundColor Gray
    Write-Host "  -DBName        数据库名称 (默认: yunai_dev)" -ForegroundColor Gray
    Write-Host ""
    Write-Host "示例:" -ForegroundColor White
    Write-Host "  .\init_database.ps1" -ForegroundColor Gray
    Write-Host "  .\init_database.ps1 -DBName yunai_prod" -ForegroundColor Gray
}

# 检查是否请求帮助
if ($args -contains "-h" -or $args -contains "--help" -or $args -contains "-Help") {
    Show-Help
    exit 0
}

# 运行主函数
Main
