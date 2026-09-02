# АІС «Пожежна частина»

Курсова робота з дисципліни «Бази даних і знань»: автоматизована інформаційна система пожежної частини.

Go (консольний API-сервер, міст до MS Access через ODBC) · React + Next.js + SCSS (BEM) — веб-інтерфейс · ШІ-асистент (aimlapi) · PDF-звіти · база знань.

## Архітектура

```
fire-station.exe   консольний API-сервер http://localhost:8080
                   (БД Access через ODBC, CRUD, статистика, PDF-звіти, ШІ + контекст БЗ;
                    логування запитів + recover від панік)
web/ (Next.js)     веб-інтерфейс http://localhost:3000
                   (proxy /api/* → localhost:8080)
```

## Сторінки

- **Огляд** — дашборд: картки статистики, графік викликів за 14 днів, останні виклики зі статусами
- **Структура** — інтерактивна ERD-мапа БД: ноди з колонками й типами, зум (коліщатко/кнопки), панорама, перетягування нод і ядра, вигин FK-ліній за ручки, фільтри гілок, мінімапа; розкладка зберігається у БД (`graph_layouts`); клік по ноді (без руху) відкриває таблицю
- **Таблиці** — картки всіх таблиць; усередині — повний CRUD: пошук, додавання, редагування, видалення; PK-чип з копіюванням, FK-чипи кольору гілки, статуси-бейджі; форми у модальних вікнах
- **База знань** — правила реагування «якщо → то» (таблиця `kb_rules`), картки за категоріями, CRUD; правила підмішуються в контекст AI-асистента
- **Новий виклик** — швидка форма реєстрації
- **Звіти** — PDF «Виклики за період» зі скачуванням
- **AI-асистент** — запит українською → SQL (Access) → таблиця (лише SELECT)
- **Налаштування** — шлях БД, шрифт, ключ ШІ
- **404** — власна сторінка для неіснуючих маршрутів

UI у стилі Supabase: сайдбар з групами таблиць, топбар зі станом БД, власні SVG-іконки (без емоджі і сторонніх бібліотек).

## Вимоги (одноразово)

1. Go 1.22+
2. MinGW-w64 (gcc) — ODBC потребує cgo. Наприклад через Scoop: `scoop install mingw`
3. Node.js 20+ — для фронтенду
4. Microsoft Access Database Engine 2016 (або встановлений Office з Access)
5. Шрифт `DejaVuSans.ttf` у `assets/fonts/` (для кирилиці в PDF)

## Запуск

Термінал 1 — бекенд (консоль):

```powershell
go mod tidy
go build -o fire-station.exe .
./fire-station.exe
```

Термінал 2 — фронтенд:

```powershell
cd web
npm install        # лише перший раз
npm run dev
```

Відкрити [http://localhost:3000](http://localhost:3000).

Першим запуском бекенд сам створить `data/fire_station.accdb`, схему (12 таблиць + `kb_rules` + `graph_layouts`), демо-дані й демо-правила бази знань.

## Модель даних

Основні 12 таблиць: positions, shifts, employees, employee_shifts, equipment,
equipment_checks, fire_types, calls, dispatches, dispatch_crew, dispatch_equipment, damages.

Службові: kb_rules (база знань), graph_layouts (розкладка ERD-мапи).

Деталі — `internal/db/schema.go`, метадані для UI — `internal/db/meta.go`, БЗ і розкладка — `internal/db/kb.go`.

## API (бекенд)

```
GET    /api/health · /api/stats · /api/stats/calls-by-day · /api/recent
GET    /api/meta · /api/ref/{table}
GET    /api/layout · PUT /api/layout        розкладка ERD-мапи
GET    /api/tables · /api/tables/{name}
POST   /api/tables/{name}/rows · PUT/DELETE /api/tables/{name}/rows/{id}
GET    /api/fire-types · POST /api/calls
POST   /api/reports/calls · GET /api/reports/file/{name}
POST   /api/ai                          запит до ШІ (з контекстом бази знань)
GET/PUT /api/config
```
