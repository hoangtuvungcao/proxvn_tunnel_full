@echo off
setlocal
echo ========================================================
echo       ProxVN Multi-Platform Build Tool
echo ========================================================

REM Switch to project root directory regardless of where script is run
pushd "%~dp0.."

REM Create bin directory
if not exist "bin" mkdir bin

REM Ensure Icon exists in bin for Linux/Mac usage
echo [0/5] Preparing resources...
if exist "scripts\icon.png" copy "scripts\icon.png" "bin\icon.png" >nul
if exist "scripts\icon.png" copy "scripts\icon.png" "bin\proxvn.png" >nul

REM --------------------------------------------------------
REM 1. Windows Build (AMD64)
REM --------------------------------------------------------
echo [1/5] Building for Windows (amd64)...
%USERPROFILE%\go\bin\rsrc.exe -ico backend\cmd\client\proxvn.ico -o backend\cmd\client\rsrc_windows.syso
%USERPROFILE%\go\bin\rsrc.exe -ico backend\cmd\server\proxvn.ico -o backend\cmd\server\rsrc_windows.syso
cd backend
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o ..\bin\svproxvn.exe .\cmd\server
go build -ldflags="-s -w" -o ..\bin\proxvn.exe .\cmd\client
cd ..

REM --------------------------------------------------------
REM 2. Linux Build (AMD64) + Desktop Entry
REM --------------------------------------------------------
echo [2/5] Building for Linux (amd64)...
cd backend
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o ..\bin\proxvn-linux-server .\cmd\server
go build -ldflags="-s -w" -o ..\bin\proxvn-linux-client .\cmd\client
cd ..

REM Generate .desktop file for Linux
echo    - Generating Linux Desktop Entry...
(
echo [Desktop Entry]
echo Version=1.0
echo Type=Application
echo Name=ProxVN Client
echo Comment=Secure Tunnel Client
echo Exec=./proxvn-linux-client
echo Icon=./icon.png
echo Terminal=true
echo Categories=Network;Utility;
) > bin\proxvn-linux.desktop

REM --------------------------------------------------------
REM 3. macOS Build
REM --------------------------------------------------------
echo [3/5] Building for macOS...
cd backend
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o ..\bin\proxvn-mac-intel .\cmd\client
set GOARCH=arm64
go build -ldflags="-s -w" -o ..\bin\proxvn-mac-m1 .\cmd\client
cd ..

REM --------------------------------------------------------
REM 4. Android (Termux)
REM --------------------------------------------------------
echo [4/5] Building for Android...
cd backend
set GOOS=android
set GOARCH=arm64
go build -ldflags="-s -w" -o ..\bin\proxvn-android .\cmd\client
cd ..

REM --------------------------------------------------------
REM 5. Packaging Server
REM --------------------------------------------------------
echo [5/5] Packaging Server (server.tar.gz)...
REM Create a temporary directory for packaging to keep root clean
if not exist "bin\server_package" mkdir "bin\server_package"

REM Copy binaries
copy "bin\svproxvn.exe" "bin\server_package\" >nul
copy "bin\proxvn-linux-server" "bin\server_package\" >nul

REM Copy frontend folder for Dashboard
echo    - Copying frontend assets...
xcopy "frontend" "bin\server_package\frontend\" /E /I /Q /Y >nul

REM Compress using tar (available on Windows 10+)
echo    - Compressing...
cd bin
tar -czf server.tar.gz -C server_package .
cd ..

REM Cleanup temp folder
rd /s /q "bin\server_package"

REM Restore original directory
popd

REM --------------------------------------------------------
echo.
echo ========================================================
echo ✅ Build Complete!
echo ========================================================
echo Windows:  bin\proxvn.exe
echo Linux:    bin\proxvn-linux-client
echo macOS:    bin\proxvn-mac-m1 / intel
echo Android:  bin\proxvn-android
echo.
echo 📦 SERVER PACKAGE: bin\server.tar.gz
echo    (Contains: svproxvn.exe, proxvn-linux-server, frontend/)
echo.
