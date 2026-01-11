# 🔏 Настройка подписи кода через SignPath Foundation

> **Руководство по получению и настройке бесплатного сертификата подписи кода для Open Source проектов**

---

## 📋 Что такое SignPath Foundation?

[SignPath Foundation](https://signpath.org/) предоставляет **бесплатные сертификаты подписи кода** для Open Source проектов. Подписанные приложения:

- Не вызывают предупреждений Windows SmartScreen
- Показывают информацию о издателе при установке
- Вызывают больше доверия у пользователей

---

## ✅ Требования для получения сертификата

| Требование | Описание |
|------------|----------|
| **OSI License** | Проект должен использовать OSI-одобренную лицензию (MIT, Apache, GPL и др.) |
| **Active Project** | Проект должен активно развиваться |
| **Released** | Должен быть хотя бы один релиз |
| **Public Repo** | Репозиторий должен быть публичным |
| **Code of Conduct** | Код должен соответствовать этическим требованиям |

**CurRate-Go соответствует всем требованиям:**
- ✅ MIT License
- ✅ Активная разработка (v1.0.0, v1.1.0)
- ✅ Публичный репозиторий на GitHub
- ✅ Code Signing Policy добавлена в README

---

## 📝 Процесс получения сертификата

### Шаг 1: Скачивание формы заявки

Скачайте форму:
```
https://signpath.org/assets/OSSRequestForm-v4.xlsx
```

### Шаг 2: Заполнение формы

| Поле | Значение |
|------|----------|
| Name | CurRate-Go |
| Handle | currate-go |
| Synopsis | Currency converter based on official CBR exchange rates |
| Description | High-performance currency converter written in Go that fetches exchange rates from the official XML API of the Central Bank of Russia. Desktop application built with Wails v2. |
| Homepage | https://github.com/bivlked/CurRate-Go |
| Repository | https://github.com/bivlked/CurRate-Go |
| License | MIT |
| Your Name | Ivan Bondarev |
| Your Email | biv@lesnet.ru |
| Your GitHub | bivlked |

### Шаг 3: Отправка заявки

Отправьте заполненную форму на:
```
oss-support@signpath.org
```

**Тема:** `SignPath Foundation OSS Application - CurRate-Go`

**Пример письма:**
```
Hello SignPath Foundation Team,

I am applying for the free code signing program for my open-source project CurRate-Go.

Project details:
- Name: CurRate-Go
- Repository: https://github.com/bivlked/CurRate-Go
- License: MIT
- Description: High-performance currency converter desktop application
  using Go and Wails framework

The project meets all Code of Conduct requirements:
- OSI-approved license (MIT)
- Active maintenance with recent releases
- Open source with full source code available
- Code Signing Policy has been added to README

Please find the completed OSSRequestForm-v4.xlsx attached.

Best regards,
Ivan Bondarev
```

### Шаг 4: Ожидание одобрения

SignPath обычно отвечает в течение **3-7 рабочих дней**.

---

## ⚙️ Настройка после одобрения

### 1. Установка SignPath GitHub App

После одобрения:
1. Войдите в [SignPath Dashboard](https://app.signpath.io/)
2. Перейдите в настройки проекта
3. Установите GitHub App для вашего репозитория

### 2. Добавление секретов в GitHub

Перейдите в **Settings → Secrets and variables → Actions** и добавьте:

| Секрет | Описание |
|--------|----------|
| `SIGNPATH_API_TOKEN` | API токен из SignPath Dashboard |
| `SIGNPATH_ORGANIZATION_ID` | ID организации (UUID) |

### 3. Активация workflow с подписью

1. Удалите текущий `.github/workflows/release.yml`
2. Переименуйте `.github/workflows/release-with-signing.yml.template` в `release.yml`
3. Закоммитьте изменения

```bash
git mv .github/workflows/release.yml .github/workflows/release-old.yml
git mv .github/workflows/release-with-signing.yml.template .github/workflows/release.yml
git add .
git commit -m "feat: enable SignPath code signing for releases"
git push
```

---

## 🔄 Как работает подпись

```
┌─────────────────────────────────────────────────────────────────┐
│                    GitHub Actions Workflow                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Build Job (Windows)                                          │
│     ├── Checkout code                                            │
│     ├── Build CurRate.exe                                        │
│     └── Upload unsigned artifact                                 │
│                                                                  │
│  2. Sign Job (Ubuntu)                                            │
│     ├── Download unsigned artifact                               │
│     ├── Submit to SignPath API                                   │
│     │   └── SignPath verifies:                                   │
│     │       - Build origin (GitHub Actions)                      │
│     │       - Source code integrity                              │
│     │       - Artifact hash                                      │
│     ├── Wait for signing completion                              │
│     └── Download signed artifact                                 │
│                                                                  │
│  3. Release Job (Ubuntu)                                         │
│     ├── Download signed artifact                                 │
│     └── Create GitHub Release with signed .exe                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📊 Origin Verification

SignPath использует **Origin Verification** для гарантии того, что:

1. ✅ Артефакт собран из официального репозитория
2. ✅ Сборка выполнена в GitHub Actions (не локально)
3. ✅ Исходный код соответствует тегу релиза
4. ✅ Никакие изменения не внесены между сборкой и подписью

Это обеспечивает **максимальный уровень доверия** к подписанным файлам.

---

## ⚠️ Важные замечания

### UPX компрессия

**UPX-сжатие несовместимо с подписью кода.** SignPath требует оригинальный PE файл. Если нужна компрессия, она должна применяться **после** подписи, но это может сломать подпись.

**Рекомендация:** Отказаться от UPX в пользу подписи кода. Размер файла увеличится (~20-25 MB вместо ~10 MB), но приложение будет доверенным.

### Время подписи

Процесс подписи занимает **1-5 минут**. Workflow учитывает это с помощью `wait-for-completion: true`.

---

## 🔗 Полезные ссылки

- [SignPath Foundation](https://signpath.org/)
- [SignPath Documentation](https://docs.signpath.io/)
- [GitHub Action](https://github.com/SignPath/github-action-submit-signing-request)
- [Code of Conduct](https://signpath.org/terms)

---

## 📅 Статус

| Этап | Статус |
|------|--------|
| Code Signing Policy в README | ✅ Добавлено |
| Форма заявки | ⏳ Готова к заполнению |
| Отправка заявки | ⏳ Ожидает |
| Одобрение SignPath | ⏳ Ожидает |
| Настройка GitHub | ⏳ Ожидает |
| Первый подписанный релиз | ⏳ Ожидает |

---

*Документ создан: 2026-01-11*
