package tui

import (
	"fmt"

	"github.com/FareinheitsTemp/fire_station/internal/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// dashModel — сторінка «Огляд»: статистика БД + останні виклики.
type dashModel struct {
	stats  db.Stats
	recent []db.RecentCall
}

func newDashModel() *dashModel { return &dashModel{} }

// Refresh перечитує статистику з БД (no-op без підключення).
func (d *dashModel) Refresh(s *db.Store) {
	if s == nil {
		return
	}
	if stats, err := s.Stats(context_()); err == nil {
		d.stats = stats
	}
	if recent, err := s.RecentCalls(context_(), 5); err == nil {
		d.recent = recent
	}
}

func (d *dashModel) Update(msg tea.Msg, s *db.Store) tea.Cmd {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "r" {
		d.Refresh(s)
	}
	return nil
}

func (d *dashModel) View() string {
	card := func(title string, value string) string {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(clrDim)).
			Padding(0, 2).Width(24).
			Render(lipgloss.NewStyle().Foreground(lipgloss.Color(clrDim)).Render(title) + "\n\n" +
				lipgloss.NewStyle().Bold(true).Render(value))
	}

	// Семантичний колір для картки техніки: зелений, якщо все в строю
	eqValue := fmt.Sprint(d.stats.EquipmentOK)
	eqColored := lipgloss.NewStyle().Foreground(lipgloss.Color(clrOK)).Render(eqValue)
	eqCard := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(clrDim)).
		Padding(0, 2).Width(24).
		Render(lipgloss.NewStyle().Foreground(lipgloss.Color(clrDim)).Render("Техніки в строю") + "\n\n" + eqColored)

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		card("Викликів усього", fmt.Sprint(d.stats.TotalCalls)), "  ",
		card("Викликів сьогодні", fmt.Sprint(d.stats.TodayCalls)), "  ",
		card("Працівників активно", fmt.Sprint(d.stats.ActiveEmployees)), "  ",
		eqCard,
	)

	out := row + "\n\n" + lipgloss.NewStyle().Bold(true).Render("Останні виклики:") + "\n"
	if len(d.recent) == 0 {
		out += lipgloss.NewStyle().Foreground(lipgloss.Color(clrFaint)).Render("  (поки порожньо)") + "\n"
	}
	for _, rc := range d.recent {
		out += fmt.Sprintf("  • %s — %s (%s)\n",
			rc.CallAt.Format("02.01 15:04"), rc.Address, statusColored(rc.Status))
	}
	out += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(clrFaint)).Render("r — оновити статистику")
	return out
}

// statusColored — семантичні кольори статусів:
// зелений — завершено, жовтий — в роботі, червоний — новий/критичний.
func statusColored(status string) string {
	switch status {
	case "завершений":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(clrOK)).Render(status)
	case "в роботі", "ремонт":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(clrWarn)).Render(status)
	case "новий":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(clrError)).Render(status)
	default:
		return status
	}
}
