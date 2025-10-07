@echo off
setlocal enabledelayedexpansion

REM Build script for Retrograde Application Server
REM Builds optimized production binaries

echo ======================================
echo Retrograde Application Server - Build
echo ======================================

REM Get current platform info
for /f "delims=" %%i in ('go env GOOS') do set CURRENT_OS=%%i
for /f "delims=" %%i in ('go env GOARCH') do set CURRENT_ARCH=%%i

echo Current Platform: %CURRENT_OS%/%CURRENT_ARCH%
echo.

REM Build flags for production optimization
set BUILD_FLAGS=-ldflags=-s -ldflags=-w -trimpath

REM 1. Build for current platform/OS in main directory
echo [1/2] Building for current platform (%CURRENT_OS%/%CURRENT_ARCH%)...

if "%CURRENT_OS%"=="windows" (
    set OUTPUT_SERVER=retrograde.exe
) else (
    set OUTPUT_SERVER=retrograde
)

go build %BUILD_FLAGS% -o %OUTPUT_SERVER% ./cmd/server
if errorlevel 1 (
    echo [X] Failed to build for current platform
    exit /b 1
)
echo [✓] Built: %CD%\%OUTPUT_SERVER%

if exist "%OUTPUT_SERVER%" (
    for %%A in ("%OUTPUT_SERVER%") do (
        echo   Size: %%~zA bytes
    )
)

REM Copy Windows executable to release directory
if "%CURRENT_OS%"=="windows" (
    if not exist "release" mkdir release
    copy "%OUTPUT_SERVER%" "release\%OUTPUT_SERVER%" >nul
    echo [✓] Copied to: %CD%\release\%OUTPUT_SERVER%
)

echo.

REM 2. Build Linux binary to release/ directory
echo [2/2] Building Linux binary (linux/amd64)...

REM Ensure release directory exists
if not exist "release" mkdir release

REM Build Linux binary with cross-compilation in isolated environment
setlocal
set GOOS=linux
set GOARCH=amd64
go build %BUILD_FLAGS% -o release/retrograde-linux ./cmd/server
if errorlevel 1 (
    echo [X] Failed to build Linux binary
    endlocal
    exit /b 1
)
endlocal

echo [✓] Built: %CD%\release\retrograde-linux

if exist "release\retrograde-linux" (
    for %%A in ("release\retrograde-linux") do (
        echo   Size: %%~zA bytes
    )
)

echo.
echo ======================================
echo Build Complete!
echo ======================================
echo.
echo Binaries created:
echo   + %CD%\%OUTPUT_SERVER% (%CURRENT_OS%/%CURRENT_ARCH%)
echo   + %CD%\release\%OUTPUT_SERVER% (%CURRENT_OS%/%CURRENT_ARCH%)
echo   + %CD%\release\retrograde-linux (linux/amd64)
echo.
echo Usage:
echo   + %OUTPUT_SERVER%        - Start BBS server
echo   + %OUTPUT_SERVER% config - Run configuration editor
echo   + %OUTPUT_SERVER% edit   - Run configuration editor (alias)
echo.
echo Production optimizations applied:
echo   + Strip debug symbols (-ldflags=-s)
echo   + Strip DWARF symbols (-ldflags=-w)
echo   + Remove file system paths (-trimpath)
echo.
echo Ready for deployment!

endlocal