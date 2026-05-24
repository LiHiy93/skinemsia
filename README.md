# 🤝 Скинемся

Telegram Mini App для группового разделения расходов.

**Стек:** React 18 + Vite (frontend) · Go 1.22 (backend) · PostgreSQL 16 · Docker Compose

---

## Быстрый старт (локально)

### 1. Создай бота

1. Открой [@BotFather](https://t.me/BotFather) → `/newbot`
2. Введи имя и username бота (например `SkinemsiaBot`)
3. Скопируй токен: `1234567890:AABBxxx...`

### 2. Настрой окружение

```bash
cp .env.example .env
```

Открой `.env` и вставь токен бота:
```
BOT_TOKEN=1234567890:AABBxxx...
WEB_APP_URL=http://localhost:5173
APP_SECRET=придумай_случайную_строку_32_символа
```

### 3. Запусти через Docker Compose

```bash
docker compose up --build
```

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432

### 4. Первый запуск без Telegram (тест в браузере)

Открой http://localhost:5173 в браузере. Создай файл `frontend/.env.local`:
```
VITE_DEV_USER_ID=123456789
```

Это позволит тестировать без открытия в Telegram (используется фиктивный пользователь).

---

## Структура проекта

```
tgbot/
├── docker-compose.yml
├── .env                    ← создай из .env.example
├── backend/
│   ├── cmd/server/main.go  ← точка входа
│   ├── internal/
│   │   ├── api/            ← HTTP handlers (Chi router)
│   │   ├── bot/            ← Telegram bot (long polling)
│   │   ├── config/         ← конфигурация из env
│   │   ├── db/             ← подключение + миграции
│   │   ├── models/         ← Go structs
│   │   └── store/          ← все запросы к PostgreSQL
│   ├── migrations/         ← SQL схема
│   ├── go.mod
│   └── Dockerfile
└── frontend/
    ├── src/
    │   ├── api/client.ts   ← API клиент
    │   ├── pages/          ← экраны приложения
    │   ├── components/     ← переиспользуемые компоненты
    │   ├── types/          ← TypeScript типы
    │   └── utils/          ← форматирование денег и др.
    ├── package.json
    └── Dockerfile
```

---

## API

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/api/events` | Список событий пользователя |
| `POST` | `/api/events` | Создать событие |
| `POST` | `/api/events/join` | Войти по коду |
| `GET` | `/api/events/:id/summary` | Сводка события |
| `GET` | `/api/events/:id/expenses` | Список расходов |
| `POST` | `/api/events/:id/expenses` | Добавить расход |
| `POST` | `/api/events/:id/payment/paid` | Отметить «оплачено» |

Авторизация: заголовок `Authorization: Bearer <initData>` (Telegram WebApp initData).  
В dev-режиме можно использовать `X-Dev-User-ID: 123456789`.

---

## Деплой бесплатно

> **Рекомендованная схема MVP:**  
> Frontend → **Netlify** (у тебя уже есть аккаунт) · Backend → **Render** (free, засыпает) · DB → **Neon**
>
> Подробный гайд с альтернативами — в [DEPLOYMENT.md](DEPLOYMENT.md).

### 1. Frontend на Netlify

У тебя уже есть аккаунт — используй его.

> **Твой тариф:** Free plan, 300 кредитов/мес (кредитная модель с мая 2026).  
> Для статического Vite-сайта расход минимальный: **1 деплой = 15 кредитов**, трафик и запросы — копейки.  
> 300 кредитов = ~20 деплоев в месяц. При 2-5 деплоях в неделю кредиты не закончатся.  
> Если всё же исчерпаются — сайт недоступен до 1-го числа; запасной вариант — Cloudflare Pages.

1. В дашборде нажми **Add new project → Import an existing project**.
2. Выбери **GitHub** → авторизуй → выбери репозиторий.
3. Настройки сборки:
   | Поле | Значение |
   |------|----------|
   | Build command | `npm run build` |
   | Publish directory | `dist` |
   | Base directory | `frontend` |
4. Environment variables (Site configuration → Environment variables):
   - `VITE_API_URL` = `https://your-backend.onrender.com`
5. Deploy site → дождись сборки.
6. Запомни URL: `https://your-app.netlify.app`

> 💡 Если кредиты Netlify закончатся — перенеси frontend на Cloudflare Pages за 10 минут,  
> просто поменяв URL в BotFather и в `ALLOWED_ORIGINS` на Render.

### 2. Получение HTTPS URL для Mini App

После деплоя получишь URL вида `https://your-app.netlify.app` — именно его вставлять в BotFather.

### 3. Вставить URL в BotFather

1. Открой [@BotFather](https://t.me/BotFather) → `/mybots` → выбери бота.
2. **Bot Settings → Menu Button → Configure menu button**.
3. Введи URL: `https://your-app.netlify.app`
4. Введи текст кнопки: `Скинемся`

### 4. Переменные окружения на хостинге

**Netlify (frontend):**  
Site configuration → Environment variables:
- `VITE_API_URL` = `https://your-backend.onrender.com`

**Render (backend):**  
Environment → Add Variable:
- `DATABASE_URL` = `<строка из Neon>`
- `BOT_TOKEN` = `<токен из BotFather>`
- `APP_SECRET` = `<случайная строка>`
- `WEB_APP_URL` = `https://your-app.netlify.app`
- `ALLOWED_ORIGINS` = `https://your-app.netlify.app`
- `PORT` = `8080`

### 5. Backend Go на Render

1. [render.com](https://render.com) → New → Web Service → GitHub.
2. Настройки:
   ```
   Root Directory:  backend
   Environment:     Go
   Build Command:   go build -o app ./cmd/server
   Start Command:   ./app
   Instance Type:   Free
   ```
3. Добавь переменные окружения (см. выше).

> ⚠️ Free-инстанс засыпает через **15 минут** простоя, cold start ~1 мин.  
> Добавь keep-alive пинг через [cron-job.org](https://cron-job.org) → URL `/health` каждые 10 мин.

### 6. База данных PostgreSQL (Neon)

1. [neon.tech](https://neon.tech) → Create project → регион US East.
2. Connection string: `postgresql://user:pass@ep-xxx.us-east-2.aws.neon.tech/neondb?sslmode=require`
3. Вставь в `DATABASE_URL` на Render.

### 7. Лимиты бесплатных вариантов

| Сервис | Ключевые лимиты |
|--------|-----------------|
| **Netlify** | 300 кредитов/мес, ~20 деплоев |
| **Render Free** | 750 часов/мес, sleep 15 мин, cold start ~1 мин |
| **Neon Free** | 100 CU-часов/мес, 0.5 GB хранилища |

### 8. Что делать если backend засыпает

- Добавь `/health` эндпоинт (уже реализован) + пинг через [cron-job.org](https://cron-job.org) каждые 10 мин
- Или апгрейд до **Render Starter ($7/мес)** — always-on, никакого cold start

### 9. Когда переходить на платный тариф

- Пользователи жалуются на медленный запуск (cold start)
- Трафик > 500 уникальных/день
- Neon 100 CU-часов/мес исчерпываются до конца месяца
- Появилась монетизация

---

## Команды бота

| Команда | Описание |
|---------|----------|
| `/start` | Приветствие и кнопка открытия Mini App |
| `/events` | Список твоих событий |
| `/new` | Создать новое событие |
| `/join КОД` | Войти в событие по коду |
| `/help` | Справка |
