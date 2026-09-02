package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/FareinheitsTemp/fire_station/internal/db"
	"github.com/FareinheitsTemp/fire_station/internal/report"
	"github.com/charmbracelet/huh"
)

// formWidth підбирає ширину форми за розміром термінала.
func (m *model) formWidth() int {
	w := m.width - 4
	if w < 50 {
		w = 70
	}
	return w
}

// openNewCallForm — форма реєстрації виклику.
func (m *model) openNewCallForm() {
	if m.store == nil {
		m.status = "БД недоступна"
		return
	}
	types, err := m.store.FireTypes()
	if err != nil || len(types) == 0 {
		m.status = "Довідник типів пожеж порожній"
		return
	}
	opts := make([]huh.Option[string], len(types))
	for i, t := range types {
		opts[i] = huh.NewOption(t.Name, fmt.Sprint(t.ID))
	}

	var address, district, caller, phone, descr, typeID string
	m.form = huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Адреса виклику").Value(&address),
		huh.NewInput().Title("Район").Value(&district),
		huh.NewSelect[string]().Title("Тип пожежі").Options(opts...).Value(&typeID),
		huh.NewInput().Title("Заявник (ПІБ)").Value(&caller),
		huh.NewInput().Title("Телефон заявника").Value(&phone),
		huh.NewInput().Title("Опис ситуації").Value(&descr),
	)).WithWidth(m.formWidth())

	m.onDone = func() {
		if address == "" {
			m.status = "Адреса порожня — виклик не збережено"
			return
		}
		var tid int64
		fmt.Sscan(typeID, &tid)
		id, err := m.store.CreateCall(db.CallInput{
			Address: address, District: district, CallerName: caller,
			CallerPhone: phone, FireTypeID: tid, Description: descr,
		})
		if err != nil {
			m.status = "Помилка запису: " + err.Error()
			return
		}
		m.status = fmt.Sprintf("Виклик №%d зареєстровано о %s", id, time.Now().Format("15:04 02.01.2006"))
		m.dash.Refresh(m.store)
	}
}

// openReportForm — форма генерації PDF-звіту за період.
func (m *model) openReportForm() {
	if m.store == nil {
		m.status = "БД недоступна"
		return
	}
	now := time.Now()
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	fromStr := first.Format("2006-01-02")
	toStr := now.Format("2006-01-02")

	m.form = huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("З дати (РРРР-ММ-ДД)").Value(&fromStr),
		huh.NewInput().Title("По дату (РРРР-ММ-ДД)").Value(&toStr),
	)).WithWidth(m.formWidth())

	m.onDone = func() {
		from, err1 := time.ParseInLocation("2006-01-02", fromStr, time.Local)
		to, err2 := time.ParseInLocation("2006-01-02", toStr, time.Local)
		if err1 != nil || err2 != nil {
			m.status = "Невірний формат дати (приклад: 2026-09-01)"
			return
		}
		to = to.Add(24*time.Hour - time.Second)

		rows, err := m.store.CallsByPeriod(context.Background(), from, to)
		if err != nil {
			m.status = "Запит: " + err.Error()
			return
		}
		if len(rows) == 0 {
			m.status = "За цей період викликів немає"
			return
		}
		out, err := report.CallsByPeriodPDF(exeRelative(m.cfg.FontPath), exeRelative("reports"), from, to, rows)
		if err != nil {
			m.status = "PDF: " + err.Error()
			return
		}
		m.status = "PDF збережено: " + out
	}
}

// openAIForm — форма запиту до AI-асистента.
func (m *model) openAIForm() {
	if m.store == nil {
		m.status = "БД недоступна"
		return
	}
	if m.cfg.AIKey == "" {
		m.status = "Немає AI-ключа — спочатку Налаштування (клавіша 6)"
		return
	}
	var question string
	m.form = huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Питання до даних українською").
			Placeholder("наприклад: скільки викликів по районах цього місяця").
			Value(&question),
	)).WithWidth(m.formWidth())

	m.onDone = func() {
		if question == "" {
			return
		}
		m.status = "[AI] Формую SQL і виконую запит..."
		m.pendingCmd = m.runAI(question)
	}
}

// openSettingsForm — форма налаштувань (зберігається у конфіг 0600).
func (m *model) openSettingsForm() {
	dbPath := m.cfg.DBPath
	fontPath := m.cfg.FontPath
	aiKey := ""
	aiModel := m.cfg.AIModel

	m.form = huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Шлях до БД (.accdb)").Value(&dbPath),
		huh.NewInput().Title("Шрифт для PDF (TTF)").Value(&fontPath),
		huh.NewInput().Title("API ключ ШІ aimlapi (порожньо — лишити)").Value(&aiKey).Password(true),
		huh.NewInput().Title("Модель ШІ").Value(&aiModel),
	)).WithWidth(m.formWidth())

	m.onDone = func() {
		m.cfg.DBPath = dbPath
		m.cfg.FontPath = fontPath
		if aiKey != "" {
			m.cfg.AIKey = aiKey
		}
		m.cfg.AIModel = aiModel
		if err := m.cfg.Save(); err != nil {
			m.status = "Збереження: " + err.Error()
			return
		}
		m.status = "Налаштування збережено (якщо змінено шлях БД — перезапусти програму)"
	}
}
