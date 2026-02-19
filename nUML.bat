@echo off
REM Check if go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Go is not installed or not in PATH.
    pause
    exit /b
)

REM Run all go files in the directory
go run . %*
