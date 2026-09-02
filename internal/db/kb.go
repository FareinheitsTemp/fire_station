package db

import (
	"context"
	"database/sql"
	"fmt"
)

// EnsureExtras створює службові таблиці (база знань, розкладка графа), якщо їх ще немає.
func (s *Store) EnsureExtras() error {
	if err := s.ensureTable("kb_rules", `CREATE TABLE [kb_rules] (
		[id] COUNTER PRIMARY KEY,
		[topic] TEXT NOT NULL,
		[category] TEXT,
		[condition_text] TEXT,
		[recommendation] TEXT,
		[priority] TEXT
	)`); err != nil {
		return fmt.Errorf("kb_rules: %w", err)
	}
	if err := s.ensureTable("graph_layouts", `CREATE TABLE [graph_layouts] (
		[id] COUNTER PRIMARY KEY,
		[node_name] TEXT NOT NULL,
		[dx] DOUBLE,
		[dy] DOUBLE
	)`); err != nil {
		return fmt.Errorf("graph_layouts: %w", err)
	}
	return s.seedKnowledge()
}

// ensureTable створює таблицю, якщо її не існує (перевірка простим запитом).
func (s *Store) ensureTable(name, ddl string) error {
	ctx := context.Background()
	rows, err := s.Query(ctx, fmt.Sprintf("SELECT COUNT(*) FROM [%s]", name))
	if err == nil {
		rows.Close()
		return nil
	}
	_, err = s.db.ExecContext(ctx, ddl)
	return err
}

// seedKnowledge наповнює базу знань демонстраційними правилами реагування.
func (s *Store) seedKnowledge() error {
	ctx := context.Background()
	var n int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM kb_rules").Scan(&n); err != nil || n > 0 {
		return err
	}
	rules := []struct{ topic, cat, cond, rec, prio string }{
		{"Пожежа в житловому будинку", "реагування", "тип пожежі — житлова, ризик високий", "висувати мінімум 2 відділення та автодрабину; евакуація мешканців — пріоритет", "високий"},
		{"Пожежа на виробничому об'єкті", "реагування", "адреса належить промисловій зоні", "уточнити наявність небезпечних речовин; узгодити дії з газорятувальною службою", "високий"},
		{"Пожежа транспорту", "реагування", "пожежа транспортного засобу на дорозі", "від'єднати живлення/паливо; виставити огородження проїжджої частини", "середній"},
		{"Відкрита територія (ліс/поле)", "реагування", "пожежа на відкритій території за вітру", "залучати автоцистерни великої ємності; враховувати напрямок вітру", "середній"},
		{"Нічний виклик", "персонал", "виклик у проміжку 22:00–06:00", "посилений караул; другий ешелон у підвищеній готовності", "низький"},
		{"Несправна техніка", "техніка", "результат перевірки — «несправно»", "не допускати техніку до виїздів до усунення; перевести статус у «ремонт»", "високий"},
	}
	for _, r := range rules {
		if err := s.InsertRule(r.topic, r.cat, r.cond, r.rec, r.prio); err != nil {
			return err
		}
	}
	return nil
}

// InsertRule додає правило до бази знань (використовується і агентом авто-моду).
func (s *Store) InsertRule(topic, category, condition, recommendation, priority string) error {
	_, err := s.db.ExecContext(context.Background(),
		"INSERT INTO kb_rules (topic, category, condition_text, recommendation, priority) VALUES (?,?,?,?,?)",
		topic, category, condition, recommendation, priority)
	return err
}

// KnowledgeRule — правило бази знань.
type KnowledgeRule struct {
	Topic, Category, Condition, Recommendation, Priority string
}

// KnowledgeRules повертає всі правила бази знань.
func (s *Store) KnowledgeRules(ctx context.Context) ([]KnowledgeRule, error) {
	rows, err := s.Query(ctx, "SELECT topic, category, condition_text, recommendation, priority FROM kb_rules")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []KnowledgeRule
	for rows.Next() {
		var r KnowledgeRule
		var cat, cond, rec, pr sql.NullString
		if err := rows.Scan(&r.Topic, &cat, &cond, &rec, &pr); err != nil {
			return nil, err
		}
		r.Category, r.Condition, r.Recommendation, r.Priority = cat.String, cond.String, rec.String, pr.String
		out = append(out, r)
	}
	return out, rows.Err()
}

// LayoutPos — зміщення елемента графа відносно позиції за замовчуванням.
type LayoutPos struct {
	DX float64 `json:"dx"`
	DY float64 `json:"dy"`
}

// Layouts повертає збережені позиції елементів графа структури.
func (s *Store) Layouts() map[string]LayoutPos {
	out := map[string]LayoutPos{}
	rows, err := s.Query(context.Background(), "SELECT node_name, dx, dy FROM graph_layouts")
	if err != nil {
		return out // таблиці може не бути — не фатально
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var dx, dy float64
		if err := rows.Scan(&name, &dx, &dy); err == nil {
			out[name] = LayoutPos{DX: dx, DY: dy}
		}
	}
	return out
}

// SaveLayout зберігає (upsert) позицію елемента графа: ноди, ядра чи вигину лінії.
func (s *Store) SaveLayout(node string, dx, dy float64) error {
	ctx := context.Background()
	res, err := s.db.ExecContext(ctx, "UPDATE graph_layouts SET dx = ?, dy = ? WHERE node_name = ?", dx, dy, node)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		_, err = s.db.ExecContext(ctx, "INSERT INTO graph_layouts (node_name, dx, dy) VALUES (?, ?, ?)", node, dx, dy)
	}
	return err
}
