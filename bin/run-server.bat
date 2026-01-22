@echo off
REM ProxVN Server Start Script for Windows

echo ========================================
echo   ProxVN Server - Windows
echo ========================================
echo.

REM Check if binary exists
if not exist "bin\server\proxvn-server-windows-amd64.exe" (
    echo [ERROR] Server binary not found!
    echo Please run build-all.sh first
    echo.
    pause
    exit /b 1
)

REM Check .env file
if not exist ".env" (
    echo [WARNING] .env file not found, using defaults
    echo [TIP] Copy .env.server.example to .env and customize
    echo.
)

REM Run server
echo [OK] Server binary found
echo.
echo Starting server...
echo.
bin\server\proxvn-server-windows-amd64.exe

pause
