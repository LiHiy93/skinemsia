# Deployment Guide — Telegram Mini App

Стек: React + Vite (frontend) · Go (backend) · PostgreSQL · Docker Compose (локально).

> **Данные актуальны на май 2026.** Бесплатные тарифы меняются часто — перед деплоем проверяй текущие лимиты на сайте сервиса.

---

## Сравнительная таблица вариантов

### Frontend

| Сервис | Бесплатно | Засыпает | Нужна карта | Ключевые лимиты | Для чего годится |
|--------|-----------|----------|-------------|-----------------|-----------------|
| **Netlify** (у тебя есть!) | Да (300 кредитов/мес) | Да, при 0 кредитов | Нет | 1 деплой = 15 кр., трафик/запросы — копейки; итого ~20 деплоев/мес для статики | ✅ Используй — аккаунт есть, кредитов хватит |
| **Cloudflare Pages** | Да, бессрочно | Нет | Нет | 500 сборок/мес, 20K файлов, безлимитный трафик | ✅ Резерв: лучше по лимитам |
| **Vercel (Hobby)** | Да, бессрочно | Нет | Нет | 100 GB трафик/мес, только личные проекты (не коммерция) | ✅ Отлично, если нет коммерции |

### Backend (Go)

| Сервис | Бесплатно | Засыпает | Нужна карта | Ключевые лимиты | Для чего годится |
|--------|-----------|----------|-------------|-----------------|-----------------|
| **Render Free** | Да, бессрочно | **Да, 15 мин** | Нет | 750 часов/мес, cold start ~1 мин | ✅ MVP, терпимый cold start |
| **Koyeb Free** | Да, бессрочно | **Да, scale-to-zero** | Нет | 1 сервис, 512 MB RAM, 0.1 vCPU | ✅ Альтернатива Render |
| **Railway** | Нет (trial $5, 30 дней; потом $1/мес) | Нет | Нет | $1 кредит/мес после trial — хватит лишь на минимальный трафик | ⚠️ Только для быстрого теста |
| **Fly.io** | **Нет** (только 2-часовой trial) | N/A | **Да** после trial | Trial 2 VM-часа или 7 дней | ❌ Не подходит для бесплатного MVP |
| **Oracle Cloud Always Free** | Да, бессрочно | Нет (если активен) | Для верификации (не списывается) | 4 ARM vCPU + 24 GB RAM суммарно, idle-reclaim после 7 дней CPU<20% | ✅ Лучший always-on бесплатный вариант, но нужна Linux-настройка |
| **Hetzner VPS (CAX11)** | **Нет** (~€3.99/мес) | Нет | Да | 2 ARM vCPU, 4 GB RAM, 40 GB SSD | ✅ Самый дешёвый надёжный VPS |

> **Fly.io** удалил free tier для новых аккаунтов в 2024. Только 2-часовой trial.  
> **Koyeb** был куплен Mistral AI (февраль 2026), но free tier сохранён.  
> **Railway** уменьшил free до $1/мес кредитов после разового $5 trial.

### PostgreSQL

| Сервис | Бесплатно | Засыпает/паузируется | Нужна карта | Ключевые лимиты | Для чего годится |
|--------|-----------|---------------------|-------------|-----------------|-----------------|
| **Neon Free** | Да, бессрочно | Да, при исчерпании 100 CU-часов/мес | Нет | 0.5 GB хранилища, 100 CU-часов/мес | ✅ Лучший выбор для MVP |
| **Supabase Free** | Да, бессрочно | **Да, после 7 дней неактивности** | Нет | 500 MB, 2 активных проекта, 5 GB egress | ✅ Хорошо, если есть регулярный трафик |
| **Render PostgreSQL** | Да, но **удаляется через 30 дней** | Нет (просто удаляется) | Нет | 1 GB, 256 MB RAM, нет бэкапов | ❌ Только для временного теста |
| **Railway PostgreSQL** | В рамках $1/мес кредитов | Нет | Нет | Расходует тот же кредитный бюджет | ⚠️ Очень мало ресурсов |
| **Supabase Pro** | Нет ($25/мес) | Нет | Да | Без пауз, 8 GB хранилища | ✅ Для продакшена |
| **Neon Launch** | Нет ($19/мес) | Нет | Да | Без пауз, 10 GB хранилища | ✅ Для продакшена |

---

## Схема A — Максимально простая и бесплатная (MVP, тест с друзьями)

```
Telegram Mini App
       │
       ▼
Cloudflare Pages / Vercel   ← React + Vite (HTTPS, без sleep)
       │ API calls (HTTPS)
       ▼
   Render Free              ← Go backend (ЗАСЫПАЕТ после 15 мин простоя!)
       │ DATABASE_URL
       ▼
   Neon Free                ← PostgreSQL (пауза при исчерпании 100 CU-часов/мес)
```

**Плюсы:** Всё бесплатно, карта не нужна, деплой занимает 30-60 минут.  
**Минусы:** Backend засыпает → первый запрос после паузы ждёт ~1 минуту.

### Пошаговый деплой Схемы A

#### Шаг 1 — Создать репозиторий GitHub

```bash
git init
git add .
git commit -m "initial commit"
gh repo create my-tgbot --public --push
```

#### Шаг 2 — База данных (Neon)

1. Зайди на [neon.tech](https://neon.tech) → Sign up (GitHub).
2. Create project → выбери регион **us-east-2** (AWS, близко к Render US East).
3. Скопируй Connection string:
   ```
   postgresql://user:pass@ep-xxx.us-east-2.aws.neon.tech/neondb?sslmode=require
   ```
4. Сохрани — это твой `DATABASE_URL`.

#### Шаг 3 — Backend (Render)

1. Зайди на [render.com](https://render.com) → Sign up (GitHub).
2. New → Web Service → Connect GitHub → выбери репозиторий.
3. Настройки:
   ```
   Root Directory:  backend
   Environment:     Go
   Build Command:   go build -o app .
   Start Command:   ./app
   Instance Type:   Free
   ```
4. Environment Variables:
   ```
   DATABASE_URL     = <строка из Neon>
   BOT_TOKEN        = <токен из BotFather>
   APP_SECRET       = <случайная строка 32+ символа>
   PORT             = 8080
   ALLOWED_ORIGINS  = https://your-app.netlify.app
   ```
5. Create Web Service → дождись деплоя.
6. Запомни URL: `https://your-backend.onrender.com`

#### Шаг 4 — Frontend (Netlify — у тебя уже есть аккаунт)

1. Зайди в свой [app.netlify.com](https://app.netlify.com) дашборд.
2. **Add new project → Import an existing project → GitHub**.
3. Настройки:
   ```
   Build command:          npm run build
   Publish directory:      dist
   Base directory:         frontend
   ```
4. Environment variables (Site configuration → Environment variables):
   ```
   VITE_API_URL      = https://your-backend.onrender.com
   VITE_BOT_USERNAME = your_bot_username
   ```
5. Deploy site → дождись сборки.
6. Запомни URL: `https://your-app.netlify.app`

> 💡 Если кредиты Netlify закончатся — перенеси frontend на Cloudflare Pages за 10 минут,  
> просто поменяв URL в BotFather и в `ALLOWED_ORIGINS` на Render.

#### Шаг 5 — Обновить ALLOWED_ORIGINS на Render

В Render → Environment → обнови:
```
ALLOWED_ORIGINS = https://your-app.netlify.app
```
Manual Deploy → нажми Redeploy.

#### Шаг 6 — Вставить URL в BotFather

1. Открой [@BotFather](https://t.me/BotFather) в Telegram.
2. `/mybots` → выбери бота → **Bot Settings → Menu Button**.
3. Введи URL: `https://your-app.netlify.app`
4. Введи текст кнопки.

Готово. Открой Telegram, запусти бота — кнопка меню откроет Mini App.

---

## Схема B — Более надёжная (реальные пользователи)

```
Telegram Mini App
       │
       ▼
Cloudflare Pages / Vercel   ← React + Vite (HTTPS, без sleep)
       │ API calls (HTTPS)
       ▼
   Render Starter ($7/мес)  ← Go backend, always-on, 512 MB RAM
   ИЛИ Hetzner VPS (~€4/мес)← Docker + Nginx + Let's Encrypt
       │ DATABASE_URL
       ▼
   Neon Launch ($19/мес)    ← PostgreSQL без пауз, 10 GB
   ИЛИ Supabase Pro ($25/мес)
```

**Плюсы:** Нет cold start, нет пауз БД, надёжность для публичного запуска.  
**Стоимость:** $26–45/мес в зависимости от выбора.

---

## Keep-alive для Схемы A (продлить бесплатный период)

Если хочешь избежать sleep на Render без апгрейда:

### Вариант 1 — cron-job.org (бесплатно)

1. Добавь `/health` эндпоинт в Go-бэкенд (возвращает `200 OK`).
2. Зарегистрируйся на [cron-job.org](https://cron-job.org).
3. Create cronjob:
   - URL: `https://your-backend.onrender.com/health`
   - Schedule: каждые 10 минут
4. Render не заснёт, пока есть запросы.

⚠️ При keep-alive каждые 10 минут за месяц = ~4320 запросов.  
Render Free даёт 750 часов/мес — при постоянном keep-alive ~744 часа, всё вписывается.  
Но это единственный Free Web Service — если нужно несколько, часы делятся.

### Вариант 2 — Oracle Cloud Always Free (если хочешь избавиться от sleep совсем)

Даёт 4 ARM vCPU + 24 GB RAM суммарно — можно развернуть Go + PostgreSQL на одном VPS.

```bash
# На Oracle ARM VM (Ubuntu 22.04)
curl -fsSL https://get.docker.com | sh
git clone https://github.com/you/your-tgbot
cd your-tgbot
cp backend/.env.example backend/.env  # заполни переменные
docker compose -f docker-compose.prod.yml up -d
# Настрой Nginx + Certbot для HTTPS
```

Требует: знание Linux, настройка firewall, обновление ОС вручную.  
Зато бесплатно и навсегда, без cold start.

---

## Итоговая рекомендация

**Для MVP начни так:**
- Frontend → **Netlify** (у тебя уже есть аккаунт — просто подключи GitHub-репозиторий)
- Backend → **Render Free** (бесплатно, засыпает, но приемлемо для теста с друзьями)
- База → **Neon Free** (бесплатно, 100 CU-часов/мес, 0.5 GB хранилища)

Добавь **keep-alive через cron-job.org** — это уберёт большинство холодных стартов.

> **Резервный план по frontend:** если кредиты Netlify (300/мес) начнут заканчиваться —  
> за 10 минут мигрируй на **Cloudflare Pages** (безлимитный трафик, нет кредитной модели).

**Если backend засыпает или появились реальные пользователи:**
- Перенеси backend на **Render Starter ($7/мес)** — минимальные изменения конфигурации, нет cold start.
- Или разверни на **Hetzner VPS (€3.99/мес CAX11)** — больше контроля, дешевле, ARM-архитектура.
- Базу переведи на **Neon Launch ($19/мес)** или **Supabase Pro ($25/мес)** — без пауз.

---

## Часто задаваемые вопросы

**Q: Mini App не открывается — белый экран.**  
A: Проверь VITE_API_URL в Cloudflare Pages. Убедись, что backend задеплоен и ALLOWED_ORIGINS содержит URL frontend.

**Q: CORS ошибки в консоли браузера.**  
A: В ALLOWED_ORIGINS на Render укажи точный URL frontend без trailing slash.

**Q: Neon отвечает медленно / Connection timeout.**  
A: Убедись, что DATABASE_URL содержит `?sslmode=require`. Проверь регион — выбирай тот же, что у Render.

**Q: Суpabase проект "Paused".**  
A: Зайди в [supabase.com/dashboard](https://supabase.com/dashboard), выбери проект → нажми **Restore**. После этого сделай хотя бы один запрос к БД раз в неделю.

**Q: Render не видит Go-проект.**  
A: Убедись, что в корне `backend/` есть `go.mod`. Build command должен компилировать бинарник: `go build -o app .`

**Q: Telegram говорит "Web App unavailable".**  
A: URL в BotFather должен быть HTTPS. `https://your-app.pages.dev` работает, `http://` — нет.
