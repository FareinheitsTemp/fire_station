package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// tableModel — модель інтерактивної TUI-таблиці (Bubble Tea + Bubbles).
type tableModel struct {
	tbl   table.Model
	title string
}

func (m tableModel) Init() tea.Cmd { return nil }

func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	var c tea.Cmd
	m.tbl, c = m.tbl.Update(msg)
	return m, c
}

func (m tableModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(m.title)
	help := lipgloss.NewStyle().Faint(true).Render("↑/↓ — гортати • q/esc — назад до меню")
	return title + "\n" + m.tbl.View() + "\n" + help + "\n"
}

// showTable показує дані в інтерактивній TUI-таблиці зі скролом і підсвічуванням.
func showTable(title string, headers []string, data [][]string) error {
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
		table.WithHeight(15),
	)
	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Bold(true)
	styles.Selected = styles.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(styles)

	_, err := tea.NewProgram(tableModel{tbl: t, title: title}).Run()
	return err
}

// rowsToTable конвертує *sql.Rows у заголовки + рядки рядків для TUI-таблиці.
func rowsToTable(rows *sql.Rows) ([]string, [][]string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	var data [][]string
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		line := make([]string, len(cols))
		for i, v := range vals {
			line[i] = cellText(v)
		}
		data = append(data, line)
	}
	return cols, data, rows.Err()
}

// colWidth підбирає ширину колонки за вмістом (мін. 8, макс. 40).
func colWidth(header string, data [][]string, idx int) int {
	w := len([]rune(header))
	for _, r := range data {
		if idx < len(r) {
			if l := len([]rune(r[idx])); l > w {
				w = l
			}
		}
	}
	if w > 40 {
		w = 40
	}
	if w < 8 {
		w = 8
	}
	return w
}

func cellText(v any) string {
	switch t := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(t)
	case time.Time:
		return t.Format("02.01.2006 15:04")
	case bool:
		if t {
			return "так"
		}
		return "ні"
	default:
		return fmt.Sprint(t)
	}
}

// exeRelative перетворює відносний шлях у шлях відносно теки з exe-файлом.
// Інакше при запуску з іншої теки (наприклад, C:\Users\Zver>) файли
// створювалися б не поруч із програмою.
func exeRelative(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	exe, err := os.Executable()
	if err != nil {
		return p
	}
	return filepath.Join(filepath.Dir(exe), p)
}
