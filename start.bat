@echo off
title Service Manage Platform
cd /d "%~dp0"
set ROOT_DIR=%cd%

set LOG_DIR=%~dp0logs
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"

echo ========================================
echo   Service Manage Platform
echo ========================================
echo.

echo [1/4] Checking Go...
go version >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [FAIL] Go not found. Please install Go 1.21+
    pause
    exit /b 1
)
for /f "tokens=3" %%i in ('go version') do echo   Version: %%i

echo [2/4] Downloading Go dependencies...
cd server
call go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo [FAIL] go mod tidy failed.
    echo   Try: go env -w GOPROXY=https://goproxy.cn,direct
    pause
    exit /b 1
)
echo   OK

cd ..

echo [3/4] Building backend...
cd server
call go build -o server.exe .
if %ERRORLEVEL% NEQ 0 (
    echo [FAIL] Backend build failed.
    pause
    exit /b 1
)
echo   Build OK
echo.
echo [4/5] Starting backend (port 8080)...
start "backend" /D "%ROOT_DIR%\server" cmd /k "chcp 65001 >nul && set NO_COLOR=1 && server.exe >> "%LOG_DIR%\server.log" 2>&1"
echo   Backend log: %LOG_DIR%\server.log
timeout /t 3 /nobreak >nul

echo [5/5] Starting frontend (port 5173)...
start "frontend" /D "%ROOT_DIR%\web" cmd /k "chcp 65001 >nul && set NO_COLOR=1 && npm run dev >> "%LOG_DIR%\web.log" 2>&1"
echo   Frontend log: %LOG_DIR%\web.log

echo.
echo ========================================
echo   Done!
echo   Frontend: http://localhost:5173
echo   Backend:  http://localhost:8080
echo.
echo   Server log:   %LOG_DIR%\server.log
echo   Web log:      %LOG_DIR%\web.log
echo ========================================
echo.
echo Press any key to close (services keep running)...
pause >nul