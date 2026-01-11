<#
.SYNOPSIS
    Сборка CurRate-Go с интеграцией Telegram
.DESCRIPTION
    Этот скрипт собирает приложение с внедрением токена Telegram бота.
    Токен и Chat ID должны быть установлены в переменных окружения или переданы как параметры.
.EXAMPLE
    .\build-with-telegram.ps1
    .\build-with-telegram.ps1 -BotToken "YOUR_TOKEN" -ChatID "YOUR_CHAT_ID"
#>

param(
    [string]$BotToken = $env:TELEGRAM_BOT_TOKEN,
    [string]$ChatID = $env:TELEGRAM_CHAT_ID,
    [switch]$Dev,
    [switch]$UPX
)

# Проверка наличия токенов
if ([string]::IsNullOrEmpty($BotToken)) {
    Write-Warning "TELEGRAM_BOT_TOKEN не установлен. Telegram интеграция будет отключена."
    Write-Host "Установите переменную окружения или передайте параметр -BotToken"
    $BotToken = ""
}

if ([string]::IsNullOrEmpty($ChatID)) {
    Write-Warning "TELEGRAM_CHAT_ID не установлен. Telegram интеграция будет отключена."
    Write-Host "Установите переменную окружения или передайте параметр -ChatID"
    $ChatID = ""
}

# Формируем ldflags
$ldflags = "-s -w"
if (-not [string]::IsNullOrEmpty($BotToken) -and -not [string]::IsNullOrEmpty($ChatID)) {
    $ldflags = "-s -w -X 'github.com/bivlked/CurRate-Go/internal/telegram.botToken=$BotToken' -X 'github.com/bivlked/CurRate-Go/internal/telegram.chatID=$ChatID'"
    Write-Host "Telegram интеграция: ВКЛЮЧЕНА" -ForegroundColor Green
} else {
    Write-Host "Telegram интеграция: ОТКЛЮЧЕНА" -ForegroundColor Yellow
}

# Сборка
if ($Dev) {
    Write-Host "Запуск в режиме разработки..."
    wails dev
} else {
    Write-Host "Сборка production версии..."
    wails build -clean -ldflags $ldflags

    if ($LASTEXITCODE -eq 0) {
        $exePath = "build\bin\CurRate.exe"

        if ($UPX) {
            Write-Host "Применение UPX сжатия..."
            upx --best --lzma $exePath
        }

        $size = [math]::Round((Get-Item $exePath).Length / 1MB, 2)
        Write-Host "Сборка завершена: $exePath ($size MB)" -ForegroundColor Green
    } else {
        Write-Error "Ошибка сборки!"
        exit 1
    }
}
