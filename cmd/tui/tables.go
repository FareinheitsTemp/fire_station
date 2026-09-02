package tui

import (
	"context"
	"fmt"

	"github.com/FareinheitsTemp/fire_station/internal/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// tablesModel — сторінка «Таблиці»: список таблиць + перегляд вмісту.
type tablesModel struct {
	names  []string
	idx    int
	viewer *tableView
	err    string
}

func newTablesModel() *tablesModel {
	return &tablesModel{names: db.TableNames()}
}

func (t *tablesModel) Update(msg tea.Msg, s *db.Store) tea.Cmd {
	// Всередині переглядача таблиці: esc — назад до списку
	if t.viewer != nil {
		if k, ok := msg.(tea.KeyMsg); ok && k.String() == "esc" {
			t.viewer = nil
			return nil
		}
		return t.viewer.Update(msg)
	}

	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	switch k.String() {
	case "up", "k":
		if t.idx > 0 {
			t.idx--
		}
	case "down", "j":
		if t.idx < len(t.names)-1 {
			t.idx++
		}
	case "enter":
		if s == nil {
			return nil
		}
		name := t.names[t.idx]
		rows, err := s.Query(context.Background(), fmt.Sprintf("SELECT TOP 200 * FROM [%s]", name))
		if err != nil {
			t.err = err.Error()
			return nil
		}
		defer rows.Close()
		headers, data, err := rowsToTable(rows)
		if err != nil {
			t.err = err.Error()
			return nil
		}
		t.err = ""
		t.viewer = newTableView("Таблиця: "+name, headers, data, 18)
	}
	return nil
}

func (t *tablesModel) View() string {
	if t.viewer != nil {
		return t.viewer.View()
	}

	out := lipgloss.NewStyle().Bold(true).Render("Таблиці бази даних:") + "\n\n"
	for i, name := range t.names {
		if i == t.idx {
			out += lipgloss.NewStyle().Bold(true).Reverse(true).Render(" "+name+" ") + "\n"
		} else {
			out += lipgloss.NewStyle().Foreground(lipgloss.Color(clrDim)).Render("  "+name) + "\n"
		}
	}
	if t.err != "" {
		out += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(clrError)).Render("Помилка: "+t.err) + "\n"
	}
	out += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(clrFaint)).Render("↑/↓ — вибір таблиці • enter — відкрити")
	return out
}
