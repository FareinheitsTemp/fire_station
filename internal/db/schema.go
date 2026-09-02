package db

// Схема БД у діалекті Access SQL (ACE):
// COUNTER = автоінкремент, TEXT(n) = рядок, LONG = ціле, DOUBLE = дробове,
// DATETIME = дата/час, YESNO = булеве, CURRENCY = гроші, LONGTEXT = memo.
// Зовнішні ключі оголошуються інлайн через CONSTRAINT ... FOREIGN KEY.

type ddlStatement struct {
	table string
	sql   string
}

var schemaStatements = []ddlStatement{
	{"positions", `CREATE TABLE positions (
		id COUNTER CONSTRAINT pk_positions PRIMARY KEY,
		name TEXT(100) NOT NULL,
		category TEXT(50)
	)`},
	{"fire_types", `CREATE TABLE fire_types (
		id COUNTER CONSTRAINT pk_fire_types PRIMARY KEY,
		name TEXT(100) NOT NULL
	)`},
	{"shifts", `CREATE TABLE shifts (
		id COUNTER CONSTRAINT pk_shifts PRIMARY KEY,
		shift_no LONG NOT NULL,
		start_at DATETIME,
		end_at DATETIME
	)`},
	{"employees", `CREATE TABLE employees (
		id COUNTER CONSTRAINT pk_employees PRIMARY KEY,
		full_name TEXT(150) NOT NULL,
		position_id LONG,
		birth_date DATETIME,
		phone TEXT(30),
		hire_date DATETIME,
		is_active YESNO,
		CONSTRAINT fk_employees_position FOREIGN KEY (position_id) REFERENCES positions(id)
	)`},
	{"employee_shifts", `CREATE TABLE employee_shifts (
		employee_id LONG NOT NULL,
		shift_id LONG NOT NULL,
		duty_date DATETIME NOT NULL,
		CONSTRAINT pk_employee_shifts PRIMARY KEY (employee_id, shift_id, duty_date),
		CONSTRAINT fk_es_employee FOREIGN KEY (employee_id) REFERENCES employees(id),
		CONSTRAINT fk_es_shift FOREIGN KEY (shift_id) REFERENCES shifts(id)
	)`},
	{"equipment", `CREATE TABLE equipment (
		id COUNTER CONSTRAINT pk_equipment PRIMARY KEY,
		name TEXT(100) NOT NULL,
		eq_type TEXT(50),
		reg_number TEXT(30),
		status TEXT(30),
		commissioned_at DATETIME
	)`},
	{"equipment_checks", `CREATE TABLE equipment_checks (
		id COUNTER CONSTRAINT pk_equipment_checks PRIMARY KEY,
		equipment_id LONG NOT NULL,
		check_date DATETIME,
		result TEXT(30),
		notes LONGTEXT,
		CONSTRAINT fk_checks_equipment FOREIGN KEY (equipment_id) REFERENCES equipment(id)
	)`},
	{"calls", `CREATE TABLE calls (
		id COUNTER CONSTRAINT pk_calls PRIMARY KEY,
		call_at DATETIME NOT NULL,
		address TEXT(200),
		district TEXT(100),
		caller_name TEXT(150),
		caller_phone TEXT(30),
		fire_type_id LONG,
		status TEXT(30),
		description LONGTEXT,
		CONSTRAINT fk_calls_fire_type FOREIGN KEY (fire_type_id) REFERENCES fire_types(id)
	)`},
	{"dispatches", `CREATE TABLE dispatches (
		id COUNTER CONSTRAINT pk_dispatches PRIMARY KEY,
		call_id LONG NOT NULL,
		depart_at DATETIME,
		arrive_at DATETIME,
		return_at DATETIME,
		water_used_liters LONG,
		result TEXT(200),
		CONSTRAINT fk_dispatches_call FOREIGN KEY (call_id) REFERENCES calls(id)
	)`},
	{"dispatch_crew", `CREATE TABLE dispatch_crew (
		dispatch_id LONG NOT NULL,
		employee_id LONG NOT NULL,
		role TEXT(50),
		CONSTRAINT pk_dispatch_crew PRIMARY KEY (dispatch_id, employee_id),
		CONSTRAINT fk_crew_dispatch FOREIGN KEY (dispatch_id) REFERENCES dispatches(id),
		CONSTRAINT fk_crew_employee FOREIGN KEY (employee_id) REFERENCES employees(id)
	)`},
	{"dispatch_equipment", `CREATE TABLE dispatch_equipment (
		dispatch_id LONG NOT NULL,
		equipment_id LONG NOT NULL,
		CONSTRAINT pk_dispatch_equipment PRIMARY KEY (dispatch_id, equipment_id),
		CONSTRAINT fk_de_dispatch FOREIGN KEY (dispatch_id) REFERENCES dispatches(id),
		CONSTRAINT fk_de_equipment FOREIGN KEY (equipment_id) REFERENCES equipment(id)
	)`},
	{"damages", `CREATE TABLE damages (
		id COUNTER CONSTRAINT pk_damages PRIMARY KEY,
		dispatch_id LONG NOT NULL,
		area_m2 DOUBLE,
		victims LONG,
		injured LONG,
		damage_uah CURRENCY,
		CONSTRAINT fk_damages_dispatch FOREIGN KEY (dispatch_id) REFERENCES dispatches(id)
	)`},
}

// EnsureSchema створює таблиці, яких ще немає у файлі БД.
func (s *Store) EnsureSchema() error {
	for _, st := range schemaStatements {
		var one int
		if err := s.db.QueryRow("SELECT COUNT(*) FROM [" + st.table + "]").Scan(&one); err == nil {
			continue // таблиця вже існує
		}
		if _, err := s.db.Exec(st.sql); err != nil {
			return fmt.Errorf("create %s: %w", st.table, err)
		}
	}
	return nil
}

// TableNames — список таблиць для меню «Довідники».
func TableNames() []string {
	return []string{
		"positions", "employees", "shifts", "employee_shifts",
		"equipment", "equipment_checks", "fire_types", "calls",
		"dispatches", "dispatch_crew", "dispatch_equipment", "damages",
	}
}

// SchemaDescription — текстовий опис схеми для системного промпта ШІ.
func SchemaDescription() string {
	return `Схема БД пожежної частини (MS Access, ACE SQL):
positions(id, name, category) — посади.
employees(id, full_name, position_id->positions, birth_date, phone, hire_date, is_active) — особовий склад.
shifts(id, shift_no, start_at, end_at) — караули/зміни.
employee_shifts(employee_id->employees, shift_id->shifts, duty_date) — графік чергувань.
equipment(id, name, eq_type, reg_number, status, commissioned_at) — техніка.
equipment_checks(id, equipment_id->equipment, check_date, result, notes) — техобслуговування.
fire_types(id, name) — типи пожеж.
calls(id, call_at, address, district, caller_name, caller_phone, fire_type_id->fire_types, status, description) — виклики.
dispatches(id, call_id->calls, depart_at, arrive_at, return_at, water_used_liters, result) — виїзди.
dispatch_crew(dispatch_id->dispatches, employee_id->employees, role) — бригада виїзду.
dispatch_equipment(dispatch_id->dispatches, equipment_id->equipment) — техніка на виїзді.
damages(id, dispatch_id->dispatches, area_m2, victims, injured, damage_uah) — наслідки.`
}
