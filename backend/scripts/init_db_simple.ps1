# YUNAI Database Initialization Script
# Author: XiaoYun
# Version: v1.0

param(
    [string]$DBHost = "localhost",
    [int]$DBPort = 5432,
    [string]$DBUser = "postgres", 
    [string]$DBPassword = "5201314hdz",
    [string]$DBName = "yunai_dev",
    [string]$AdminDB = "postgres"
)

Write-Host "Starting YUNAI database initialization..." -ForegroundColor Green
Write-Host "Database: $DBName on ${DBHost}:${DBPort}" -ForegroundColor Cyan

# Set password environment variable
$env:PGPASSWORD = $DBPassword

try {
    # Test PostgreSQL connection
    Write-Host "Testing PostgreSQL connection..." -ForegroundColor Yellow
    $testResult = & pg_isready -h $DBHost -p $DBPort -U $DBUser 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "PostgreSQL connection failed: $testResult" -ForegroundColor Red
        Write-Host "Please make sure PostgreSQL is running and accessible." -ForegroundColor Red
        exit 1
    }
    Write-Host "PostgreSQL connection OK" -ForegroundColor Green

    # Check if database exists
    Write-Host "Checking if database exists..." -ForegroundColor Yellow
    $checkDB = "SELECT 1 FROM pg_database WHERE datname='$DBName'"
    $dbExists = & psql -h $DBHost -p $DBPort -U $DBUser -d $AdminDB -tAc $checkDB 2>&1
    
    if ($dbExists -eq "1") {
        Write-Host "Database $DBName already exists" -ForegroundColor Yellow
    } else {
        # Create database
        Write-Host "Creating database $DBName..." -ForegroundColor Yellow
        $createDB = "CREATE DATABASE $DBName;"
        & psql -h $DBHost -p $DBPort -U $DBUser -d $AdminDB -c $createDB 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Database $DBName created successfully" -ForegroundColor Green
        } else {
            Write-Host "Failed to create database" -ForegroundColor Red
            exit 1
        }
    }

    # Create extensions
    Write-Host "Creating database extensions..." -ForegroundColor Yellow
    $extensions = @(
        'CREATE EXTENSION IF NOT EXISTS "uuid-ossp";',
        'CREATE EXTENSION IF NOT EXISTS "pgcrypto";',
        'CREATE EXTENSION IF NOT EXISTS "pg_trgm";',
        'CREATE EXTENSION IF NOT EXISTS "btree_gin";'
    )
    
    foreach ($ext in $extensions) {
        & psql -h $DBHost -p $DBPort -U $DBUser -d $DBName -c $ext 2>&1 | Out-Null
    }
    Write-Host "Extensions created successfully" -ForegroundColor Green

    # Run migrations
    Write-Host "Running database migrations..." -ForegroundColor Yellow
    $migrationsPath = Join-Path $PSScriptRoot "..\migrations"
    
    if (Test-Path $migrationsPath) {
        $migrationFiles = Get-ChildItem -Path $migrationsPath -Filter "*.up.sql" | Sort-Object Name
        
        foreach ($file in $migrationFiles) {
            Write-Host "  Executing: $($file.Name)" -ForegroundColor Cyan
            & psql -h $DBHost -p $DBPort -U $DBUser -d $DBName -f $file.FullName 2>&1
            if ($LASTEXITCODE -ne 0) {
                Write-Host "Migration $($file.Name) failed" -ForegroundColor Red
                exit 1
            }
        }
        Write-Host "All migrations completed successfully" -ForegroundColor Green
    } else {
        Write-Host "Migrations directory not found: $migrationsPath" -ForegroundColor Red
        exit 1
    }

    # Verify database
    Write-Host "Verifying database structure..." -ForegroundColor Yellow
    $tables = @("users", "wallets", "gift_cards", "feature_flags", "ai_models")
    
    foreach ($table in $tables) {
        $checkTable = "SELECT 1 FROM information_schema.tables WHERE table_name='$table'"
        $tableExists = & psql -h $DBHost -p $DBPort -U $DBUser -d $DBName -tAc $checkTable 2>&1
        
        if ($tableExists -eq "1") {
            Write-Host "  Table ${table}: OK" -ForegroundColor Green
        } else {
            Write-Host "  Table ${table}: MISSING" -ForegroundColor Red
            exit 1
        }
    }

    # Check admin user
    $adminCheck = "SELECT 1 FROM users WHERE username='admin'"
    $adminExists = & psql -h $DBHost -p $DBPort -U $DBUser -d $DBName -tAc $adminCheck 2>&1
    
    if ($adminExists -eq "1") {
        Write-Host "  Admin user: OK" -ForegroundColor Green
    } else {
        Write-Host "  Admin user: MISSING" -ForegroundColor Red
        exit 1
    }

    Write-Host "Database verification completed successfully" -ForegroundColor Green

    # Show database info
    Write-Host "`n=== Database Information ===" -ForegroundColor Magenta
    Write-Host "Host: ${DBHost}:${DBPort}" -ForegroundColor Cyan
    Write-Host "Database: $DBName" -ForegroundColor Cyan
    Write-Host "User: $DBUser" -ForegroundColor Cyan
    Write-Host "`nDefault Admin Account:" -ForegroundColor Cyan
    Write-Host "  Username: admin" -ForegroundColor White
    Write-Host "  Email: admin@yunai.com" -ForegroundColor White
    Write-Host "  Password: admin123" -ForegroundColor White

    Write-Host "`nTest Gift Cards:" -ForegroundColor Cyan
    & psql -h $DBHost -p $DBPort -U $DBUser -d $DBName -c "SELECT card_code, card_type, description FROM gift_cards ORDER BY card_code;" 2>&1

    Write-Host "`n=== YUNAI Database Initialization Completed Successfully! ===" -ForegroundColor Green

} catch {
    Write-Host "Error occurred: $_" -ForegroundColor Red
    exit 1
} finally {
    # Clean up environment variable
    Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
}
