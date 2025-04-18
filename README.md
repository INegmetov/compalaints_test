# 🛠 Complaint Service API

Сервис на Go для обработки клиентских жалоб:

- Принимает POST-запросы с текстом жалобы
- Определяет тональность через внешний API (apilayer)
- Сохраняет данные в базу Supabase
- Определяет категорию через OpenAI (n8n workflow)
- Выполняет действия в зависимости от категории (Telegram, Google Sheets)

---

## 🚀 Установка и запуск

### 1. Установить Go

Скачать и установить Go:  
👉 https://go.dev/dl/

Проверка установки:

```bash
go version
```

---

### 2. Клонировать проект и установить зависимости

```bash
git clone https://github.com/your-username/complaint-api.git
cd complaint-api
go mod tidy
```

---

### 3. Создать `.env` файл

Создайте `.env` в корне проекта:

```env
API_KEY=your_api_key
SENTIMENT_URL=https://api.apilayer.com/sentiment/analysis
DEFAULT_STATUS=open
DEFAULT_CATEGORY=другое

DB_USER=postgres.uurimqzsejrffplsfbyf
DB_PASSWORD=your_database_password_here
DB_HOST=aws-0-eu-central-1.pooler.supabase.com
DB_PORT=6543
DB_NAME=postgres

BOT_TOKEN=your_telegram_bot_token
```

---

### 4. Запуск сервиса

```bash
go run main.go
```

---

## 🔄 Пример запроса

**POST** `/complaints`

```json
{
  "text": "Не приходит SMS-код"
}
```

**Ответ:**

```json
{
  "id": 12,
  "status": "open",
  "sentiment": "negative",
  "category": "техническая"
}
```

---

## 🤖 Автоматизация через n8n

- Определение категории через OpenAI (`GPT-3.5`)
- Категория "техническая" — уведомление в Telegram
- Категория "оплата" — запись в Google Sheets
- Все обработанные жалобы переводятся в статус `closed`

> 💡 Файл n8n workflow доступен в папке `n8n/` или по запросу.

---

## 📚 Используемые технологии

- Go `net/http`
- PostgreSQL + Supabase
- API: [apilayer Sentiment Analysis](https://apilayer.com/marketplace/sentiment-analysis-api)
- n8n (интеграция с OpenAI, Telegram, Google Sheets)

---
