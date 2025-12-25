@echo off
setlocal enabledelayedexpansion

REM ========================================
REM CurRate-Go Production Build Script
REM ========================================
REM Usage:
REM   build.bat [mode] [version]
REM
REM Modes:
REM   dev       - Development build (no optimization)
REM   prod      - Production build (optimized, no UPX)
REM   prod-upx  - Production build with UPX compression
REM
REM Examples:
REM   build.bat dev
REM   build.bat prod
REM   build.bat prod-upx
REM   build.bat prod 1.0.1
REM ========================================

REM Параметры по умолчанию
set MODE=prod
set VERSION=1.0.0

REM Парсинг аргументов
if "%1"=="dev" set MODE=dev
if "%1"=="prod" set MODE=prod
if "%1"=="prod-upx" set MODE=prod-upx
if not "%2"=="" set VERSION=%2

echo.
echo ========================================
echo Building CurRate-Go
echo Mode: %MODE%
echo Version: %VERSION%
echo ========================================
echo.

REM Development build
if "%MODE%"=="dev" (
    echo [INFO] Building development version...
    echo [INFO] No optimization flags applied
    echo.
    wails build
    if %ERRORLEVEL% NEQ 0 (
        echo.
        echo [ERROR] Build failed!
        exit /b 1
    )
    goto :check
)

REM Production build
if "%MODE%"=="prod" (
    echo [INFO] Building production version...
    echo [INFO] Optimization flags: -s -w (strip symbols, no debug info)
    echo.
    wails build -clean -ldflags "-s -w"
    if %ERRORLEVEL% NEQ 0 (
        echo.
        echo [ERROR] Build failed!
        exit /b 1
    )
    goto :check
)

REM Production build with UPX
if "%MODE%"=="prod-upx" (
    echo [INFO] Building production version with UPX compression...
    echo [INFO] Optimization flags: -s -w (strip symbols, no debug info)
    echo.
    wails build -clean -ldflags "-s -w"
    if %ERRORLEVEL% NEQ 0 (
        echo.
        echo [ERROR] Build failed!
        exit /b 1
    )
    
    echo.
    echo [INFO] Checking for UPX...
    where upx >nul 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo [INFO] UPX found. Compressing binary...
        echo [INFO] UPX flags: --best --lzma (best compression)
        echo.
        upx --best --lzma build\bin\CurRate.exe
        if %ERRORLEVEL% EQU 0 (
            echo.
            echo [SUCCESS] UPX compression complete!
        ) else (
            echo.
            echo [WARNING] UPX compression failed, but build succeeded.
        )
    ) else (
        echo.
        echo [WARNING] UPX not found. Skipping compression.
        echo [INFO] To install UPX:
        echo   - Chocolatey: choco install upx
        echo   - Manual: https://upx.github.io/
        echo.
        echo [INFO] Build completed without compression.
    )
    goto :check
)

REM Invalid mode
echo [ERROR] Invalid mode: %MODE%
echo [INFO] Valid modes: dev, prod, prod-upx
exit /b 1

:check
REM Небольшая задержка для завершения записи файла
timeout /t 1 /nobreak >nul 2>&1

REM Проверка результата
set EXE_PATH=build\bin\CurRate.exe
if exist "%EXE_PATH%" (
    echo.
    echo ========================================
    echo [SUCCESS] Build completed successfully!
    echo ========================================
    echo.
    echo File: %EXE_PATH%
    echo.
    
    REM Получение размера файла
    for %%A in ("%EXE_PATH%") do (
        set SIZE=%%~zA
        set /a SIZE_MB=!SIZE!/1024/1024
        set /a SIZE_KB=!SIZE!/1024
    )
    
    echo Size: !SIZE_MB! MB (!SIZE_KB! KB)
    echo.
    
    REM Детальная информация о файле
    dir "%EXE_PATH%"
    echo.
    
    echo [INFO] Build output location: %EXE_PATH%
    echo [INFO] You can now distribute this file to users.
    echo.
    exit /b 0
) else (
    echo.
    echo ========================================
    echo [ERROR] Build failed - executable not found!
    echo ========================================
    echo.
    echo [INFO] Expected file: %EXE_PATH%
    echo [INFO] Check the build output above for errors.
    exit /b 1
)

endlocal

