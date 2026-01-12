@echo off
chcp 65001 >nul
setlocal

:: ===== THEME =====
color 0A
title ProxVN Tunnel Launcher

:: ===== BANNER =====
cls
echo.
echo    ██████╗ ██████╗  ██████╗ ██╗  ██╗██╗   ██╗███╗   ██╗
echo    ██╔══██╗██╔══██╗██╔═══██╗╚██╗██╔╝╚██╗ ██╔╝████╗  ██║
echo    ██████╔╝██████╔╝██║   ██║ ╚███╔╝  ╚████╔╝ ██╔██╗ ██║
echo    ██╔═══╝ ██╔══██╗██║   ██║ ██╔██╗   ╚██╔╝  ██║╚██╗██║
echo    ██║     ██║  ██║╚██████╔╝██╔╝ ██╗   ██║   ██║ ╚████║
echo    ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═══╝
echo.
echo            ProxVN Tunnel Launcher
echo        --------------------------------
echo.

:: ===== CONFIG CỐ ĐỊNH =====
set CERT_PIN=5d21642f9c2ac2aef414ecb27b54cdb5d53cb6d554bbf965de19d2c8652f47c6

:: ===== INPUT =====
set /p HOST=➤ Host   [127.0.0.1]: 
if "%HOST%"=="" set HOST=127.0.0.1

set /p PORT=➤ Port   [vd: 3389 / 80]: 
if "%PORT%"=="" (
    echo.
    echo [✗] Port khong duoc de trong
    timeout /t 2 >nul
    exit
)


set /p PROTO=➤ Proto  [tcp / udp /http]: 
if "%PROTO%"=="" set PROTO=tcp

:: ===== VALIDATE PROTO =====
if /I not "%PROTO%"=="tcp" if /I not "%PROTO%"=="udp" if /I not "%PROTO%"=="http" (
    echo.
    echo [✗] Proto khong hop le! Chi nhan tcp hoac udp
    timeout /t 2 >nul
    exit
)

:: ===== SUMMARY =====
cls
echo.
echo ========================================
echo   ✓ CẤU HÌNH PROXVN TUNNEL
echo ----------------------------------------
echo   Host     : %HOST%
echo   Port     : %PORT%
echo   Protocol : %PROTO%
echo ========================================
echo.

echo [→] Dang khoi chay ProxVN Tunnel...
timeout /t 1 >nul

:: ===== RUN =====
start "ProxVN Tunnel" cmd /k "proxvn.exe --host %HOST% --port %PORT% --proto %PROTO% --cert-pin %CERT_PIN%"

echo.
echo [✓] ProxVN da chay o cua so rieng
echo [i] Launcher se tu dong dong
timeout /t 2 >nul
exit
