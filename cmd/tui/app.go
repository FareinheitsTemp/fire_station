package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/FareinheitsTemp/fire_station/internal/ai"
	"github.com/FareinheitsTemp/fire_station/internal/config"
	"github.com/FareinheitsTemp/fire_station/internal/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Стримана «службова» палітра:
// 166 — приглушений помаранч (акцент, тема пожежної служби)
// 238 — темно-сірий фон активних елементів
// 255/250 — основний текст, 244/241 — другорядний текст
// 24  — темно-синє тло вибраного рядка
// 160 — приглушений червоний (помилки), 178 — бурштиновий (статуси)
const (
	clrAccent     = "166"
	clrActiveBg   = "238"
	clrText       = "255"
	clrTextDim    = "244"
	clrFaint      = "241"
	clrSelectedBg = "24"
	clrError      = "160"
	clrStatus     = "178"
	clrCardBorder = "240"
)

type page int

const (
	pageDash page = iota
	pageTables
	pageNewCall
	pageReports
	pageAI
	pageSettings
)

var pageNames = []string{"Огляд", "Таблиці", "Новий виклик", "Звіти", "AI-асистент", "Налаштування"}

// aiResultMsg — результат асинхронного AI-запиту.
type aiResultMsg struct {
	title   string
	headers []string
	data    [][]string
	err     error
}

type model struct {
	cfg   *config.Config
	store *db.Store
	dbErr error

	page   page
	width  int
	status string

	dash   *dashModel
	tables *tablesModel

	form       *huh.Form // активна huh-форма
	onDone     func()    // дія після завершення форми
	pendingCmd tea.Cmd   // асинхронна дія (AI-запит)
	result     *tableView
}

// Run запускає повноекранний TUI-застосунок.
func Run(cfg *config.Config) error {
	m := &model{cfg: cfg, page: pageDash}
	m.dash = newDashModel()
	m.tables = newTablesModel()

	dbPath := exeRelative(cfg.DBPath)
	if err := db.EnsureDatabase(dbPath); err != nil {
		m.dbErr = err
	} else if s, err := db.Connect(dbPath); err != nil {
		m.dbErr = err
	} else {
		if err := s.EnsureSchema(); err != nil {
			m.dbErr = err
		} else if err := s.SeedDemo(); err != nil {
			m.dbErr = err
		} else {
			m.store = s
			m.dash.Refresh(s)
		}
	}

	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		if m.form != nil {
			var c tea.Cmd
			m.form, c = m.updateForm(msg)
			return m, c
		}
		return m, nil

	case aiResultMsg:
		if msg.err != nil {
			m.status = "[AI] " + msg.err.Error()
		} else {
			m.status = ""
			m.result = newTableView("[AI] SQL: "+msg.title, msg.headers, msg.data, 16)
		}
		return m, nil

	case tea.KeyMsg:
		// Глобальний вихід — завжди, навіть поверх форм
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// 1. Активна форма: esc — скасувати, решта клавіш — формі
		if m.form != nil {
			if msg.String() == "esc" {
				m.closeForm()
				return m, nil
			}
			var c tea.Cmd
			m.form, c = m.updateForm(msg)
			switch m.form.State {
			case huh.StateCompleted:
				done := m.onDone
				m.closeForm()
				if done != nil {
					done()
				}
				if m.pendingCmd != nil {
					pc := m.pendingCmd
					m.pendingCmd = nil
					return m, pc
				}
			case huh.StateAborted:
				m.closeForm()
			}
			return m, c
		}

		// 2. Таблиця результатів: esc — закрити, решта — таблиці
		if m.result != nil {
			if msg.String() == "esc" {
				m.result = nil
				return m, nil
			}
			return m, m.result.Update(msg)
		}

		// 3. Глобальна навігація
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "1":
			return m, m.enterPage(pageDash)
		case "2":
			return m, m.enterPage(pageTables)
		case "3":
			return m, m.enterPage(pageNewCall)
		case "4":
			return m, m.enterPage(pageReports)
		case "5":
			return m, m.enterPage(pageAI)
		case "6":
			return m, m.enterPage(pageSettings)
		case "right", "l":
			return m, m.enterPage((m.page + 1) % page(len(pageNames)))
		case "left", "h":
			return m, m.enterPage((m.page + page(len(pageNames)) - 1) % page(len(pageNames)))
		default:
			switch m.page {
			case pageDash:
				return m, m.dash.Update(msg, m.store)
			case pageTables:
				return m, m.tables.Update(msg, m.store)
			}
		}
	}
	return m, nil
}

// updateForm делегує повідомлення формі (huh Update міняє receiver-стан).
func (m *model) updateForm(msg tea.Msg) (*huh.Form, tea.Cmd) {
	fm, c := m.form.Update(msg)
	if f, ok := fm.(*huh.Form); ok {
		return f, c
	}
	return m.form, c
}

// closeForm скидає стан активної форми.
func (m *model) closeForm() {
	m.form = nil
	m.onDone = nil
	m.pendingCmd = nil
}

// enterPage переходить на сторінку; для сторінок-форм створює форму й
// повертає її Init-команду (фокус першого поля, блінк курсора).
func (m *model) enterPage(p page) tea.Cmd {
	m.page = p
	m.status = ""
	switch p {
	case pageDash:
		m.dash.Refresh(m.store)
	case pageNewCall:
		return m.openNewCallForm()
	case pageReports:
		return m.openReportForm()
	case pageAI:
		return m.openAIForm()
	case pageSettings:
		return m.openSettingsForm()
	}
	return nil
}

func (m *model) View() string {
	var b strings.Builder
	b.WriteString(m.headerView())
	b.WriteString("\n\n")

	switch {
	case m.dbErr != nil:
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(clrError)).Bold(true).Render("БД недоступна"))
		b.WriteString("\n\n" + m.dbErr.Error() + "\n\n")
		b.WriteString("Підказка: встанови Access Database Engine 2016\n")
		b.WriteString("(https://www.microsoft.com/en-us/download/details.aspx?id=54920)\n")
	case m.form != nil:
		b.WriteString(m.form.View())
	case m.result != nil:
		b.WriteString(m.result.View())
	case m.page == pageDash:
		b.WriteString(m.dash.View())
	case m.page == pageTables:
		b.WriteString(m.tables.View())
	}

	b.WriteString("\n\n" + m.footerView())
	return b.String()
}

func (m *model) headerView() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(clrAccent)).Render("АІС «Пожежна частина»")
	var tabs []string
	for i, name := range pageNames {
		style := lipgloss.NewStyle().Padding(0, 1)
		if page(i) == m.page {
			style = style.Bold(true).Foreground(lipgloss.Color(clrText)).Background(lipgloss.Color(clrActiveBg))
		} else {
			style = style.Foreground(lipgloss.Color(clrTextDim))
		}
		tabs = append(tabs, style.Render(fmt.Sprintf("%d %s", i+1, name)))
	}
	return title + "\n" + lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m *model) footerView() string {
	help := lipgloss.NewStyle().Foreground(lipgloss.Color(clrFaint)).Render("1–6 — сторінки • ←/→ — перемикання • esc — назад • q — вихід")
	if m.status != "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(clrStatus)).Render(m.status) + "\n" + help
	}
	return help
}

// runAI — асинхронний запит до ШІ: NL → SQL (Access) → виконання.
func (m *model) runAI(question string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		client := ai.NewClient(m.cfg.AIKey, m.cfg.AIModel)
		sqlText, err := client.GenerateSQL(ctx, question, db.SchemaDescription())
		if err != nil {
			return aiResultMsg{err: err}
		}
		if !ai.IsSafeSelect(sqlText) {
			return aiResultMsg{err: fmt.Errorf("запит відхилено політикою безпеки (дозволено лише SELECT)")}
		}
		rows, err := m.store.Query(ctx, sqlText)
		if err != nil {
			return aiResultMsg{err: fmt.Errorf("виконання: %w (SQL: %s)", err, sqlText)}
		}
		defer rows.Close()

		headers, data, err := rowsToTable(rows)
		if err != nil {
			return aiResultMsg{err: err}
		}
		return aiResultMsg{title: sqlText, headers: headers, data: data}
	}
}
