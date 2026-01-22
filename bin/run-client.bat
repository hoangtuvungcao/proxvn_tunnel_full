@echo off
REM ProxVN Client Start Script for Windows

echo ========================================
echo   ProxVN Client - Quick Start
echo ========================================
echo.

REM Check if binary exists
if not exist "\client\proxvn-windows-amd64.exe" (
    echo [ERROR] Client binary not found!
    echo Please run build-all.sh first
    echo.
    pause
    exit /b 1
)

echo [OK] Client binary found
echo.

REM Menu
echo Select mode:
echo   1) HTTP Tunnel (Web development)
echo   2) TCP Tunnel (SSH, Database, RDP...)
echo   3) UDP Tunnel (Game server)
echo   4) File Sharing
echo   5) Custom command
echo.
set /p choice="Enter choice [1-5]: "

if "%choice%"=="1" goto http_tunnel
if "%choice%"=="2" goto tcp_tunnel
if "%choice%"=="3" goto udp_tunnel
if "%choice%"=="4" goto file_share
if "%choice%"=="5" goto custom
echo [ERROR] Invalid choice!
pause
exit /b 1

:http_tunnel
set /p port="Enter local port (default: 3000): "
if "%port%"=="" set port=3000
echo.
echo [*] Starting HTTP tunnel on port %port%...
echo.
client\proxvn-windows-amd64.exe --proto http %port%
pause
exit /b 0

:tcp_tunnel
set /p port="Enter local port (e.g., 22 for SSH): "
if "%port%"=="" (
    echo [ERROR] Port is required!
    pause
    exit /b 1
)
echo.
echo [*] Starting TCP tunnel on port %port%...
echo.
client\proxvn-windows-amd64.exe %port%
pause
exit /b 0

:udp_tunnel
set /p port="Enter local port (e.g., 19132 for Minecraft): "
if "%port%"=="" (
    echo [ERROR] Port is required!
    pause
    exit /b 1
)
echo.
echo [*] Starting UDP tunnel on port %port%...
echo.
client\proxvn-windows-amd64.exe --proto udp %port%
pause
exit /b 0

:file_share
set /p folder="Enter folder path to share (e.g., .\share): "
if "%folder%"=="" (
    echo [ERROR] Folder path is required!
    pause
    exit /b 1
)
set /p username="Enter username (default: proxvn): "
if "%username%"=="" set username=proxvn
set /p password="Enter password: "
if "%password%"=="" (
    echo [ERROR] Password is required!
    pause
    exit /b 1
)
set /p perms="Enter permissions [r/rw/rwx] (default: rw): "
if "%perms%"=="" set perms=rw
echo.
echo [*] Starting file sharing...
echo.
client\proxvn-windows-amd64.exe --file "%folder%" --user "%username%" --pass "%password%" --permissions "%perms%"
pause
exit /b 0

:custom
echo.
echo Available options:
echo   --proto [tcp^|udp^|http]   Protocol type
echo   --server ^<addr:port^>     Custom server address
echo   --file ^<path^>            File sharing mode
echo   --user ^<username^>        WebDAV username (default: proxvn)
echo   --pass ^<password^>        File sharing password
echo   --permissions ^<r^|rw^|rwx^> File sharing permissions
echo   --insecure               Skip TLS verification
echo   --cert-pin ^<fingerprint^> Certificate pinning
echo   --help                   Show all options
echo.
set /p custom="Enter custom command (without binary name): "
echo.
echo [*] Running: proxvn %custom%
echo.
client\proxvn-windows-amd64.exe %custom%
pause
exit /b 0
