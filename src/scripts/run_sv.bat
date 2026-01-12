@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion

:: 1. Chuyển thư mục làm việc về nơi chứa file .bat này
cd /d "%~dp0"

:: ===== COLOR THEME =====
color 0A
title ProxVN Server Launcher

:: ===== BANNER =====
cls
echo.
echo    ██████╗ ██████╗  ██████╗ ██╗  ██╗██╗   ██╗███╗   ██╗
echo    ██╔══██╗██╔══██╗██╔═══██╗╚██╗██╔╝╚██╗ ██╔╝████╗  ██║
echo    ██████╔╝██████╔╝██║   ██║ ╚███╔╝  ╚████╔╝ ██╔██╗ ██║
echo    ██╔═══╝ ██╔══██╗██║   ██║ ██╔██╗   ╚██╔╝  ██║╚██╗██║
echo    ██║     ██║  ██║╚██████╔╝██╔╝ ██╗   ██║   ██║ ╚████║
echo    ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝ ╚═══╝
echo.
echo             ProxVN Server Launcher (Auto-Path)
echo         ----------------------------------------
echo.

:: 2. Thiết lập biến môi trường (Sửa log báo thiếu HTTP_DOMAIN)
:: Bạn có thể sửa domain dưới đây theo ý muốn
set "HTTP_DOMAIN=vutrungocrong.fun"


echo [i] Thư mục hiện tại: %CD%
echo [i] Đang cấu hình HTTP_DOMAIN: !HTTP_DOMAIN!
echo.

:: 3. Kiểm tra và Chạy Server
:: Kiểm tra xem file exe có nằm trong thư mục bin/ không
if exist "bin\svproxvn.exe" (
    echo [→] Đang khởi chạy: bin\svproxvn.exe
    echo.
    :: Dùng start để mở cửa sổ mới, truyền biến môi trường trực tiếp
    start "ProxVN Server" cmd /k "set HTTP_DOMAIN=!HTTP_DOMAIN!&& bin\svproxvn.exe"
) else (
    echo [!] LỖI: Không tìm thấy file "bin\svproxvn.exe"
    echo [i] Đảm bảo file .bat nằm cùng cấp với thư mục "bin"
    pause
    exit
)

echo [✓] Đã yêu cầu khởi chạy Server.
timeout /t 3 >nul
exit