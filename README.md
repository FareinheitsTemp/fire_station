# АІС «Пожежна частина»

Курсова робота з дисципліни «Бази даних»: автоматизована інформаційна система пожежної частини.

Go + MS Access (.accdb через ODBC) + TUI (Bubble Tea) + ШІ-асистент (aimlapi) + PDF-звіти.

## Можливості

- **Інтерактивне меню** — запуск exe подвійним кліком, жодних консольних команд
- **TUI-таблиці** — перегляд таблиць у інтерактивному вигляді зі скролом і підсвічуванням (Bubble Tea + Bubbles)
- **БД у файлі Access** — `data/fire_station.accdb` поруч з exe; відкривається в MS Access
- **CRUD** — реєстрація викликів, перегляд довідників
- **PDF-звіти** — «Виклики за період» з кирилицею (папка `reports/` поруч з exe)
- **AI-асистент** — запит українською → SQL → TUI-таблиця результатів (лише SELECT, політика безпеки)

## Вимоги (одноразово)

1. Go 1.22+
2. MinGW-w64 (gcc) — ODBC-драйвер потребує cgo. Наприклад через Scoop: `scoop install mingw`
3. Microsoft Access Database Engine 2016 (або встановлений Office з Access)
4. Шрифт `DejaVuSans.ttf` покласти в `assets/fonts/` (для кирилиці в PDF)

## Збірка та запуск

```powershell
go mod tidy          # підтягне залежності, включно з ODBC-драйвером
go build -o fire-station.exe .
./fire-station.exe   # або подвійний клік
```

Першим запуском програма сама створить `data/fire_station.accdb` **поруч з exe**, схему (12 таблиць) і демо-дані.

## AI-асистент

У меню «Налаштування» введи API-ключ aimlapi (маскований ввід, зберігається локально в
`~/.fire-station/config.yaml` з правами 0600, у репозиторій не потрапляє).
Далі пункт «AI-асистент» приймає питання на кшталт:

> «Скільки викликів було за останній тиждень по районах?»

ШІ генерує SELECT-запит під Access SQL, програма перевіряє його політикою безпеки
(заборонено INSERT/UPDATE/DELETE/DROP/INTO тощо), виконує і показує TUI-таблицю.

## Структура

```
cmd/            Cobra + меню (survey) + TUI-таблиці (Bubble Tea)
internal/
  db/           ODBC-підключення до .accdb, схема, демо-дані, запити
  report/       Генерація PDF (gofpdf)
  ai/           Клієнт aimlapi: NL→SQL з політикою безпеки
  config/       Конфіг користувача (0600)
assets/fonts/   Сюди покласти DejaVuSans.ttf
data/           Файл БД (створюється автоматично поруч з exe)
reports/        Згенеровані PDF (поруч з exe)
```

## Модель даних (12 таблиць)

positions, shifts, employees, employee_shifts, equipment, equipment_checks,
fire_types, calls, dispatches, dispatch_crew, dispatch_equipment, damages

Деталі — `internal/db/schema.go`.
