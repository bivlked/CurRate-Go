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
    $ldflags = "-s -w -X 'github.com/bivlked/currate-go/internal/telegram.botToken=$BotToken' -X 'github.com/bivlked/currate-go/internal/telegram.chatID=$ChatID'"
    Write-Host "Telegram интеграция: ВКЛЮЧЕНА" -ForegroundColor Green
} else {
    Write-Host "Telegram интеграция: ОТКЛЮЧЕНА" -ForegroundColor Yellow
}

# Wails печатает переданные ldflags в свой вывод, а значит и токен.
# Пропускаем вывод через фильтр, который маскирует токен
function Invoke-Wails {
    & wails @args 2>&1 | ForEach-Object {
        $line = "$_"
        if ($BotToken) { $line = $line.Replace($BotToken, '***') }
        Write-Host $line
    }
}

# Сборка
if ($Dev) {
    Write-Host "Запуск в режиме разработки..."
    # Без ldflags токены не внедряются, и SendStar в dev-режиме недоступен
    Invoke-Wails dev -ldflags $ldflags
} else {
    Write-Host "Сборка production версии..."
    Invoke-Wails build -clean -ldflags $ldflags

    if ($LASTEXITCODE -eq 0) {
        $exePath = "build\bin\CurRate.exe"

        if ($UPX) {
            Write-Host "Применение UPX сжатия..."
            upx --best --lzma $exePath
            if ($LASTEXITCODE -ne 0) {
                Write-Warning "UPX сжатие не удалось, используется несжатый файл"
            }
        }

        $size = [math]::Round((Get-Item $exePath).Length / 1MB, 2)
        Write-Host "Сборка завершена: $exePath ($size MB)" -ForegroundColor Green
    } else {
        Write-Error "Ошибка сборки!"
        exit 1
    }
}
