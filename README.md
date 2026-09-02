# АІС «Пожежна частина»

Курсова робота з дисципліни «Бази даних»: автоматизована інформаційна система пожежної частини.

Go (консольний API-сервер, міст до MS Access через ODBC) · React + Next.js + SCSS (BEM) — веб-інтерфейс · ШІ-асистент (aimlapi) · PDF-звіти.

## Архітектура

```
fire-station.exe   консольний API-сервер http://localhost:8080
                   (БД Access через ODBC, CRUD, статистика, PDF-звіти, ШІ;
                    логування запитів + recover від панік)
web/ (Next.js)     веб-інтерфейс http://localhost:3000
                   (proxy /api/* → localhost:8080)
```

## Сторінки

- **Огляд** — дашборд: картки статистики, графік викликів за 14 днів, останні виклики зі статусами
- **Структура** — інтерактивна мапа БД: зум (коліщатко/кнопки), панорама, перетягування нод, фільтри гілок (ядро викликів / персонал / техніка / довідники), підсвітка зв'язків, мінімапа; клік по ноді відкриває таблицю
- **Таблиці** — картки всіх 12 таблиць; усередині — повний CRUD: пошук, додавання, редагування, видалення; PK-чип з копіюванням, FK-чипи кольору гілки, статуси-бейджі; форми у модальних вікнах
- **Новий виклик** — швидка форма реєстрації
- **Звіти** — PDF «Виклики за період» зі скачуванням
- **AI-асистент** — запит українською → SQL (Access) → таблиця (лише SELECT)
- **Налаштування** — шлях БД, шрифт, ключ ШІ
- **404** — власна сторінка для неіснуючих маршрутів

UI у стилі Supabase: сайдбар з групами таблиць, топбар зі станом БД, тонкі бордери, липкі заголовки таблиць.
Дизайн-мова: монохромний білий; колір — семантика станів і гілок графа.

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
Продакшн-режим фронтенду: `npm run build` і далі `npm start`.

Першим запуском бекенд сам створить `data/fire_station.accdb` поруч з exe, схему (12 таблиць, PK = `id` COUNTER) і демо-дані.

## Структура проєкту

```
cmd/
  webapi/       JSON API-сервер (net/http, Go 1.22 routing, middleware)
internal/
  db/           ODBC-підключення, схема, демо-дані, запити, статистика, метадані таблиць
  report/       Генерація PDF (gofpdf)
  ai/           Клієнт aimlapi: NL→SQL з політикою безпеки
  config/       Конфіг користувача (0600)
web/            Next.js фронтенд (React, pages router)
  pages/        Огляд, Структура, Таблиці (CRUD), Новий виклик, Звіти, AI, Налаштування, 404
  components/   Sidebar, Topbar, Layout, NodeGraph, TableGrid, RecordForm, LineChart, DataTable, StatusBadge
  styles/       SCSS за BEM: variables + blocks
assets/fonts/   Сюди покласти DejaVuSans.ttf
data/           Файл БД (створюється автоматично поруч з exe)
reports/        Згенеровані PDF (поруч з exe)
```

## API (бекенд)

```
GET    /api/health                 стан сервера і БД
GET    /api/stats                  статистика дашборду
GET    /api/stats/calls-by-day     виклики по днях (?days=1..90)
GET    /api/recent                 останні виклики
GET    /api/meta                   метадані таблиць (поля, типи, довідники, категорії, кольори)
GET    /api/ref/{table}            довідник [{id, label}] для селектів
GET    /api/tables                 список таблиць
GET    /api/tables/{name}          вміст таблиці (TOP 500)
POST   /api/tables/{name}/rows     створити запис {values}
PUT    /api/tables/{name}/rows/{id}   оновити запис {values}
DELETE /api/tables/{name}/rows/{id}   видалити запис
GET    /api/fire-types             довідник типів пожеж
POST   /api/calls                  реєстрація виклику
POST   /api/reports/calls          PDF-звіт {from, to} → {file}
GET    /api/reports/file/{name}    скачування PDF
POST   /api/ai                     запит до ШІ {question} → {sql, columns, rows}
GET/PUT /api/config                налаштування
```

## Модель даних (12 таблиць)

positions, shifts, employees, employee_shifts, equipment, equipment_checks,
fire_types, calls, dispatches, dispatch_crew, dispatch_equipment, damages

Деталі — `internal/db/schema.go`, метадані для UI — `internal/db/meta.go`.
