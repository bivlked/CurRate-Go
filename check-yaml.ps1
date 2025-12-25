# PowerShell скрипт для проверки всех YAML файлов с помощью yamllint
# Использование: .\check-yaml.ps1

Write-Host "Проверка YAML файлов с помощью yamllint..." -ForegroundColor Cyan

$yamlFiles = @(
    ".github/workflows/build.yml",
    ".github/workflows/test.yml",
    ".github/workflows/lint.yml",
    ".github/workflows/release.yml",
    ".github/dependabot.yml"
)

$errors = 0

foreach ($file in $yamlFiles) {
    if (Test-Path $file) {
        Write-Host "`nПроверка: $file" -ForegroundColor Yellow
        yamllint $file
        if ($LASTEXITCODE -ne 0) {
            $errors++
        }
    } else {
        Write-Host "Файл не найден: $file" -ForegroundColor Red
        $errors++
    }
}

Write-Host "`n" -NoNewline
if ($errors -eq 0) {
    Write-Host "Все YAML файлы проверены успешно!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "Обнаружено ошибок: $errors" -ForegroundColor Red
    exit 1
}

