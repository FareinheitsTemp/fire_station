package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// tableView — повноекранна інтерактивна таблиця зі скролом.
// Монохром: вибраний рядок — інверсія кольорів (чорне на білому).
type tableView struct {
	tbl   table.Model
	title string
}

func newTableView(title string, headers []string, data [][]string, height int) *tableView {
	cols := make([]table.Column, len(headers))
	for i, h := range headers {
		cols[i] = table.Column{Title: h, Width: colWidth(h, data, i)}
	}
	rows := make([]table.Row, len(data))
	for i, r := range data {
		rows[i] = table.Row(r)
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)
	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(clrDim)).
		BorderBottom(true).
		Bold(true)
	styles.Selected = lipgloss.NewStyle().Reverse(true)
	t.SetStyles(styles)

	return &tableView{tbl: t, title: title}
}

func (v *tableView) Update(msg tea.Msg) tea.Cmd {
	var c tea.Cmd
	v.tbl, c = v.tbl.Update(msg)
	return c
}

func (v *tableView) View() string {
	title := lipgloss.NewStyle().Bold(true).Render(v.title)
	help := lipgloss.NewStyle().Foreground(lipgloss.Color(clrFaint)).Render("↑/↓ — гортати • esc — назад")
	return title + "\n" + v.tbl.View() + "\n" + help
}
