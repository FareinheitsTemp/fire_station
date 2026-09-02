package db

import (
	"fmt"
	"time"
)

// SeedDemo наповнює БД демонстраційними даними (один раз, якщо positions порожня).
func (s *Store) SeedDemo() error {
	var n int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM positions").Scan(&n); err != nil || n > 0 {
		return err
	}

	exec := func(q string, args ...any) error {
		_, err := s.db.Exec(q, args...)
		return err
	}

	positions := []string{"Начальник караулу", "Командир відділення", "Пожежник", "Водій", "Диспетчер"}
	for _, p := range positions {
		if err := exec("INSERT INTO positions (name, category) VALUES (?, ?)", p, "особовий склад"); err != nil {
			return fmt.Errorf("seed positions: %w", err)
		}
	}

	fireTypes := []string{"Житлова", "Лісова", "Автотранспорт", "Електромережа", "Промислова"}
	for _, t := range fireTypes {
		if err := exec("INSERT INTO fire_types (name) VALUES (?)", t); err != nil {
			return fmt.Errorf("seed fire_types: %w", err)
		}
	}

	for i := 1; i <= 3; i++ {
		if err := exec("INSERT INTO shifts (shift_no, start_at, end_at) VALUES (?, ?, ?)",
			i, time.Date(2026, 9, i, 8, 0, 0, 0, time.Local), time.Date(2026, 9, i+1, 8, 0, 0, 0, time.Local)); err != nil {
			return fmt.Errorf("seed shifts: %w", err)
		}
	}

	equipment := []struct{ name, typ, reg string }{
		{"АЦ-40 (автоцистерна)", "автоцистерна", "АІ1234ВК"},
		{"АЛ-30 (автодрабина)", "драбина", "АІ5678ВК"},
		{"АНР (аварійно-рятувальний)", "спецтехніка", "АІ9012ВК"},
	}
	for _, e := range equipment {
		if err := exec("INSERT INTO equipment (name, eq_type, reg_number, status, commissioned_at) VALUES (?, ?, ?, ?, ?)",
			e.name, e.typ, e.reg, "в строю", time.Date(2022, 3, 15, 0, 0, 0, 0, time.Local)); err != nil {
			return fmt.Errorf("seed equipment: %w", err)
		}
	}

	employees := []struct {
		name  string
		posID int64
		phone string
	}{
		{"Коваленко Андрій Петрович", 1, "0671112233"},
		{"Шевченко Олег Іванович", 2, "0672223344"},
		{"Бойко Максим Олегович", 3, "0673334455"},
		{"Ткаченко Дмитро Сергійович", 3, "0674445566"},
		{"Мельник Василь Юрійович", 4, "0675556677"},
	}
	for _, e := range employees {
		if err := exec(
			"INSERT INTO employees (full_name, position_id, birth_date, phone, hire_date, is_active) VALUES (?, ?, ?, ?, ?, ?)",
			e.name, e.posID, time.Date(1990, 5, 20, 0, 0, 0, 0, time.Local), e.phone,
			time.Date(2018, 6, 1, 0, 0, 0, 0, time.Local), true,
		); err != nil {
			return fmt.Errorf("seed employees: %w", err)
		}
	}

	// Демонстраційний виклик + виїзд + бригада + техніка + наслідки
	callAt := time.Now().Add(-48 * time.Hour)
	if err := exec(
		"INSERT INTO calls (call_at, address, district, caller_name, caller_phone, fire_type_id, status, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		callAt, "вул. Соборна, 12", "Замостянський", "Громадянка", "0679998877", int64(1), "завершений", "Займання в квартирі на 3 поверсі",
	); err != nil {
		return fmt.Errorf("seed calls: %w", err)
	}
	if err := exec(
		"INSERT INTO dispatches (call_id, depart_at, arrive_at, return_at, water_used_liters, result) VALUES (?, ?, ?, ?, ?, ?)",
		int64(1), callAt.Add(2*time.Minute), callAt.Add(9*time.Minute), callAt.Add(95*time.Minute), 2400, "Локалізовано та ліквідовано",
	); err != nil {
		return fmt.Errorf("seed dispatches: %w", err)
	}
	for _, empID := range []int64{1, 2, 3, 5} {
		if err := exec("INSERT INTO dispatch_crew (dispatch_id, employee_id, role) VALUES (?, ?, ?)", int64(1), empID, "член бригади"); err != nil {
			return fmt.Errorf("seed crew: %w", err)
		}
	}
	for _, eqID := range []int64{1, 2} {
		if err := exec("INSERT INTO dispatch_equipment (dispatch_id, equipment_id) VALUES (?, ?)", int64(1), eqID); err != nil {
			return fmt.Errorf("seed dispatch_equipment: %w", err)
		}
	}
	if err := exec(
		"INSERT INTO damages (dispatch_id, area_m2, victims, injured, damage_uah) VALUES (?, ?, ?, ?, ?)",
		int64(1), 42.5, 0, 1, 185000.0,
	); err != nil {
		return fmt.Errorf("seed damages: %w", err)
	}

	return nil
}
