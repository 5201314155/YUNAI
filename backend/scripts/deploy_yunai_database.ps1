# YUNAI Database Deployment Script
# Version: 2.0

Write-Host "===============================================================================" -ForegroundColor Blue
Write-Host "🚀 YUNAI Database Deployment Script" -ForegroundColor Blue
Write-Host "===============================================================================" -ForegroundColor Blue
Write-Host ""

# Database configuration
$DB_HOST = "localhost"
$DB_PORT = "5432"
$DB_USER = "postgres"
$DB_PASSWORD = "5201314hdz"
$DB_NAME = "yunai"

# Set environment variable to avoid password prompt
$env:PGPASSWORD = $DB_PASSWORD

Write-Host "📋 Deployment Configuration:" -ForegroundColor Blue
Write-Host "   Database Host: $DB_HOST`:$DB_PORT"
Write-Host "   Database User: $DB_USER"
Write-Host "   Database Name: $DB_NAME"
Write-Host ""

# Check PostgreSQL service
Write-Host "🔍 Checking PostgreSQL service status..." -ForegroundColor Blue
try {
    $result = & pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "PostgreSQL connection failed"
    }
    Write-Host "✅ PostgreSQL service is running" -ForegroundColor Green
} catch {
    Write-Host "❌ PostgreSQL service is not running or connection failed!" -ForegroundColor Red
    Write-Host "   Please ensure PostgreSQL service is started and connection parameters are correct." -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

# Check if database exists
Write-Host ""
Write-Host "🔍 Checking if database exists..." -ForegroundColor Blue
try {
    $dbExists = & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -lqt 2>&1 | Select-String $DB_NAME
    if ($dbExists) {
        Write-Host "⚠️  Database '$DB_NAME' already exists!" -ForegroundColor Yellow
        $choice = Read-Host "Do you want to drop the existing database and recreate it? (y/N)"
        if ($choice -eq "y" -or $choice -eq "Y") {
            Write-Host "🗑️  Dropping existing database..." -ForegroundColor Yellow
            
            # Terminate all connections to the database
            & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();" 2>&1 | Out-Null
            
            # Drop database
            & dropdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME 2>&1
            if ($LASTEXITCODE -ne 0) {
                Write-Host "❌ Failed to drop database!" -ForegroundColor Red
                Read-Host "Press Enter to exit"
                exit 1
            }
            Write-Host "✅ Database dropped successfully" -ForegroundColor Green
        } else {
            Write-Host "🚫 Deployment cancelled" -ForegroundColor Red
            Read-Host "Press Enter to exit"
            exit 0
        }
    }
} catch {
    Write-Host "Warning: Could not check database existence" -ForegroundColor Yellow
}

# Create new database
Write-Host ""
Write-Host "🏗️  Creating database '$DB_NAME'..." -ForegroundColor Blue
try {
    & createdb -h $DB_HOST -p $DB_PORT -U $DB_USER -E UTF8 $DB_NAME 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to create database"
    }
    Write-Host "✅ Database created successfully" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed to create database!" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

# Execute table creation script
Write-Host ""
Write-Host "📊 Executing table creation script..." -ForegroundColor Blue
try {
    & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "yunai_complete_database_setup.sql" 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to execute table creation script"
    }
} catch {
    Write-Host "❌ Failed to execute table creation script!" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

# Verify table creation
Write-Host ""
Write-Host "🔍 Verifying table creation..." -ForegroundColor Blue
try {
    $tableCount = & psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>&1
    Write-Host "✅ Successfully created $($tableCount.Trim()) tables" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Could not verify table creation" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "===============================================================================" -ForegroundColor Blue
Write-Host "🎉 YUNAI Database Deployment Complete!" -ForegroundColor Green
Write-Host "===============================================================================" -ForegroundColor Blue
Write-Host ""
Write-Host "📊 Database Information:" -ForegroundColor Blue
Write-Host "   Database Name: $DB_NAME"
Write-Host "   Connection: $DB_HOST`:$DB_PORT"
Write-Host "   Username: $DB_USER"
Write-Host ""
Write-Host "🚀 Next Steps:" -ForegroundColor Blue
Write-Host "   1. Start YUNAI backend server: go run yunai_core_server.go"
Write-Host "   2. Access API documentation: http://localhost:8081/swagger/index.html"
Write-Host "   3. Start functional testing"
Write-Host ""

# Clean up environment variable
Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue

Read-Host "Press Enter to exit"
