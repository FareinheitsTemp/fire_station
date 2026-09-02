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
	card := func(title string, value int) string {
		v := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(clrText)).Render(fmt.Sprint(value))
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(clrCardBorder)).
			Padding(0, 2).Width(24).
			Render(lipgloss.NewStyle().Foreground(lipgloss.Color(clrTextDim)).Render(title) + "\n\n" + v)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		card("Викликів усього", d.stats.TotalCalls), "  ",
		card("Викликів сьогодні", d.stats.TodayCalls), "  ",
		card("Працівників активно", d.stats.ActiveEmployees), "  ",
		card("Техніки в строю", d.stats.EquipmentOK),
	)

	out := row + "\n\n" + lipgloss.NewStyle().Bold(true).Render("Останні виклики:") + "\n"
	if len(d.recent) == 0 {
		out += lipgloss.NewStyle().Foreground(lipgloss.Color(clrFaint)).Render("  (поки порожньо)") + "\n"
	}
	for _, rc := range d.recent {
		out += fmt.Sprintf("  • %s — %s (%s)\n",
			rc.CallAt.Format("02.01 15:04"), rc.Address, rc.Status)
	}
	out += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color(clrFaint)).Render("r — оновити статистику")
	return out
}
