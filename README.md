# АІС «Пожежна частина»

Курсова робота з дисципліни «Бази даних і знань»: автоматизована інформаційна система пожежної частини.

Go (консольний API-сервер, міст до MS Access через ODBC) · React + Next.js + SCSS (BEM) — веб-інтерфейс · ШІ на Groq (OpenAI-сумісний API; за бажанням — локальний Ollama) · PDF-звіти · база знань · автономний агент.

## Архітектура

```
fire-station.exe   консольний API-сервер http://localhost:8080
                   (БД Access через ODBC, CRUD, статистика, PDF-звіти, ШІ + контекст БЗ;
                    логування запитів + recover від панік + журнал помилок для агента)
web/ (Next.js)     веб-інтерфейс http://localhost:3000
                   (proxy /api/* → localhost:8080)
```

## Сторінки

- **Огляд** — дашборд: картки статистики, графік викликів за 14 днів, останні виклики зі статусами
- **Структура** — інтерактивна ERD-мапа БД: ноди з колонками й типами, зум/панорама/перетягування, вигин FK-ліній, розкладка зберігається у БД (`graph_layouts`)
- **Таблиці** — повний CRUD: пошук, модальні форми, PK/FK-чипи, статуси-бейджі
- **База знань** — правила «якщо → то» (`kb_rules`), картки за категоріями, CRUD; правила — у контексті ШІ
- **Агент (авто-мод)** — автономний аналіз системи: знімок (статистика + виклики + правила + журнал помилок API) → ШІ → висновок + самостійно створені нові правила в БЗ
- **Чат** — плаваюче вікно на всіх сторінках: розмова з агентом (знає схему, правила, статистику)
- **Новий виклик · Звіти (PDF) · AI-асистент (NL→SQL) · Налаштування · 404**

UI у стилі Supabase; власні SVG-іконки (без емоджі і сторонніх бібліотек).

## Вимоги (одноразово)

1. Go 1.22+
2. MinGW-w64 (gcc) — ODBC потребує cgo
3. Node.js 20+ — для фронтенду
4. Microsoft Access Database Engine 2016 (або Office з Access)
5. `DejaVuSans.ttf` у `assets/fonts/` (кирилиця в PDF)
6. Ключ ШІ: безкоштовний [Groq](https://console.groq.com)

## Запуск

```powershell
# Термінал 1 — бекенд
go mod tidy; go build -o fire-station.exe .; ./fire-station.exe

# Термінал 2 — фронтенд
cd web; npm install; npm run dev
```

Відкрити [http://localhost:3000](http://localhost:3000).

## Налаштування ШІ

Сторінка «Налаштування» (зберігається локально у `~/.fire-station/config.yaml`, 0600, у репозиторій не потрапляє):

- **API ключ Groq** — `gsk_…`
- **Base URL** — порожньо = дефолт Groq `https://api.groq.com/openai/v1` (для локальної моделі через Ollama: `http://localhost:11434/v1`)
- **Модель** — `llama-3.3-70b-versatile` (дефолт); також production: `llama-3.1-8b-instant`, `openai/gpt-oss-120b`, `openai/gpt-oss-20b`

## Модель даних

12 основних таблиць (positions … damages) + службові: kb_rules (база знань), graph_layouts (розкладка ERD-мапи). PK усюди — `id` COUNTER.

## API (бекенд)

```
GET    /api/health · /api/stats · /api/stats/calls-by-day · /api/recent
GET    /api/meta · /api/ref/{table} · /api/layout · PUT /api/layout
GET    /api/tables · /api/tables/{name} · POST/PUT/DELETE rows (CRUD)
GET    /api/fire-types · POST /api/calls
POST   /api/reports/calls · GET /api/reports/file/{name}
POST   /api/ai            NL→SQL (з контекстом БЗ)
POST   /api/chat          чат з агентом
POST   /api/agent/analyze авто-мод: аналіз → нові правила в БЗ
GET/PUT /api/config
```
