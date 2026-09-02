package db

// Метадані таблиць для веб-UI: типи полів, обов'язковість,
// довідники (FK), категорія-гілка і колір для графа структури.
// Увага: первинний ключ У ВСІХ таблицях називається "id" (COUNTER).

type ColumnMeta struct {
	Name     string   `json:"name"`
	Label    string   `json:"label"`
	Type     string   `json:"type"` // text | number | date | bool | select | ref
	Required bool     `json:"required"`
	Options  []string `json:"options,omitempty"`
	Ref      string   `json:"ref,omitempty"`      // таблиця-довідник
	RefLabel string   `json:"refLabel,omitempty"` // колонка-мітка в довіднику
}

type TableMeta struct {
	Name     string       `json:"name"`
	Label    string       `json:"label"`
	PK       string       `json:"pk"`
	LabelCol string       `json:"labelCol"`
	Category string       `json:"category"` // core | staff | equipment | refs
	Color    string       `json:"color"`
	Columns  []ColumnMeta `json:"columns"`
}

var tablesMeta = []TableMeta{
	{Name: "calls", Label: "Виклики", PK: "id", LabelCol: "address", Category: "core", Color: "#dc2626", Columns: []ColumnMeta{
		{Name: "address", Label: "Адреса", Type: "text", Required: true},
		{Name: "district", Label: "Район", Type: "text"},
		{Name: "caller_name", Label: "Заявник", Type: "text"},
		{Name: "caller_phone", Label: "Телефон заявника", Type: "text"},
		{Name: "fire_type_id", Label: "Тип пожежі", Type: "ref", Ref: "fire_types", RefLabel: "name"},
		{Name: "description", Label: "Опис", Type: "text"},
		{Name: "call_at", Label: "Час виклику", Type: "date"},
		{Name: "status", Label: "Статус", Type: "select", Options: []string{"новий", "в роботі", "завершений"}},
	}},
	{Name: "dispatches", Label: "Виїзди", PK: "id", LabelCol: "id", Category: "core", Color: "#ea580c", Columns: []ColumnMeta{
		{Name: "call_id", Label: "Виклик (адреса)", Type: "ref", Ref: "calls", RefLabel: "address", Required: true},
		{Name: "dispatched_at", Label: "Виїзд о", Type: "date"},
		{Name: "arrived_at", Label: "Прибуття о", Type: "date"},
		{Name: "returned_at", Label: "Повернення о", Type: "date"},
		{Name: "notes", Label: "Нотатки", Type: "text"},
	}},
	{Name: "dispatch_crew", Label: "Екіпажі виїздів", PK: "id", LabelCol: "id", Category: "core", Color: "#f59e0b", Columns: []ColumnMeta{
		{Name: "dispatch_id", Label: "Виїзд №", Type: "ref", Ref: "dispatches", RefLabel: "id", Required: true},
		{Name: "employee_id", Label: "Працівник", Type: "ref", Ref: "employees", RefLabel: "full_name", Required: true},
		{Name: "role", Label: "Роль", Type: "text"},
	}},
	{Name: "dispatch_equipment", Label: "Техніка на виїздах", PK: "id", LabelCol: "id", Category: "core", Color: "#eab308", Columns: []ColumnMeta{
		{Name: "dispatch_id", Label: "Виїзд №", Type: "ref", Ref: "dispatches", RefLabel: "id", Required: true},
		{Name: "equipment_id", Label: "Техніка", Type: "ref", Ref: "equipment", RefLabel: "name", Required: true},
	}},
	{Name: "damages", Label: "Збитки", PK: "id", LabelCol: "id", Category: "core", Color: "#db2777", Columns: []ColumnMeta{
		{Name: "call_id", Label: "Виклик (адреса)", Type: "ref", Ref: "calls", RefLabel: "address", Required: true},
		{Name: "description", Label: "Опис", Type: "text"},
		{Name: "amount", Label: "Сума (грн)", Type: "number"},
	}},
	{Name: "employees", Label: "Працівники", PK: "id", LabelCol: "full_name", Category: "staff", Color: "#16a34a", Columns: []ColumnMeta{
		{Name: "full_name", Label: "ПІБ", Type: "text", Required: true},
		{Name: "birth_date", Label: "Дата народження", Type: "date"},
		{Name: "phone", Label: "Телефон", Type: "text"},
		{Name: "hired_at", Label: "Дата прийняття", Type: "date"},
		{Name: "position_id", Label: "Посада", Type: "ref", Ref: "positions", RefLabel: "name"},
		{Name: "active", Label: "Активний", Type: "bool"},
	}},
	{Name: "positions", Label: "Посади", PK: "id", LabelCol: "name", Category: "staff", Color: "#0d9488", Columns: []ColumnMeta{
		{Name: "name", Label: "Назва посади", Type: "text", Required: true},
		{Name: "category", Label: "Категорія", Type: "text"},
	}},
	{Name: "shifts", Label: "Зміни", PK: "id", LabelCol: "label", Category: "staff", Color: "#0891b2", Columns: []ColumnMeta{
		{Name: "label", Label: "Назва зміни", Type: "text", Required: true},
		{Name: "start_time", Label: "Початок (HH:MM)", Type: "text"},
		{Name: "end_time", Label: "Кінець (HH:MM)", Type: "text"},
	}},
	{Name: "employee_shifts", Label: "Графік змін", PK: "id", LabelCol: "id", Category: "staff", Color: "#22c55e", Columns: []ColumnMeta{
		{Name: "employee_id", Label: "Працівник", Type: "ref", Ref: "employees", RefLabel: "full_name", Required: true},
		{Name: "shift_id", Label: "Зміна", Type: "ref", Ref: "shifts", RefLabel: "label", Required: true},
		{Name: "work_date", Label: "Дата", Type: "date"},
	}},
	{Name: "equipment", Label: "Техніка", PK: "id", LabelCol: "name", Category: "equipment", Color: "#2563eb", Columns: []ColumnMeta{
		{Name: "name", Label: "Назва", Type: "text", Required: true},
		{Name: "category", Label: "Категорія", Type: "select", Options: []string{"автоцистерна", "драбина", "спецавто", "інше"}},
		{Name: "serial_number", Label: "Серійний №", Type: "text"},
		{Name: "status", Label: "Статус", Type: "select", Options: []string{"в строю", "ремонт", "списано"}},
		{Name: "commissioned_at", Label: "Введено в експлуатацію", Type: "date"},
	}},
	{Name: "equipment_checks", Label: "Перевірки техніки", PK: "id", LabelCol: "id", Category: "equipment", Color: "#6366f1", Columns: []ColumnMeta{
		{Name: "equipment_id", Label: "Техніка", Type: "ref", Ref: "equipment", RefLabel: "name"},
		{Name: "checked_at", Label: "Дата перевірки", Type: "date"},
		{Name: "result", Label: "Результат", Type: "select", Options: []string{"справно", "несправно"}},
		{Name: "comment", Label: "Коментар", Type: "text"},
	}},
	{Name: "fire_types", Label: "Типи пожеж", PK: "id", LabelCol: "name", Category: "refs", Color: "#9333ea", Columns: []ColumnMeta{
		{Name: "name", Label: "Назва", Type: "text", Required: true},
		{Name: "risk_level", Label: "Рівень ризику", Type: "select", Options: []string{"низький", "середній", "високий"}},
	}},
}

// TablesMeta повертає метадані всіх таблиць.
func TablesMeta() []TableMeta { return tablesMeta }

// TableMetaByName шукає метадані таблиці за іменем.
func TableMetaByName(name string) (TableMeta, bool) {
	for _, t := range tablesMeta {
		if t.Name == name {
			return t, true
		}
	}
	return TableMeta{}, false
}
