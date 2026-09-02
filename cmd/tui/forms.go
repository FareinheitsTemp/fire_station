package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/FareinheitsTemp/fire_station/internal/db"
	"github.com/FareinheitsTemp/fire_station/internal/report"
	tea "github.com/charmbracelet/bubbletea"
)

// openNewCallForm — форма реєстрації виклику на власному компоненті.
func (m *model) openNewCallForm() tea.Cmd {
	if m.store == nil {
		m.status = "БД недоступна"
		return nil
	}
	types, err := m.store.FireTypes()
	if err != nil || len(types) == 0 {
		m.status = "Довідник типів пожеж порожній"
		return nil
	}
	typeNames := make([]string, len(types))
	typeIDs := make([]int64, len(types))
	for i, t := range types {
		typeNames[i] = t.Name
		typeIDs[i] = t.ID
	}

	f, cmd := newSimpleForm("Новий виклик", []formField{
		newTextField("Адреса виклику", "вул. Соборна, 12", false),
		newTextField("Район", "Замостянський", false),
		newSelectField("Тип пожежі", typeNames),
		newTextField("Заявник (ПІБ)", "", false),
		newTextField("Телефон заявника", "067...", false),
		newTextField("Опис ситуації", "", false),
	})
	m.form = f

	m.onDone = func() {
		address := f.value(0)
		if address == "" {
			m.status = "Адреса порожня — виклик не збережено"
			return
		}
		var tid int64
		if sel := f.value(2); sel != "" {
			for i, name := range typeNames {
				if name == sel {
					tid = typeIDs[i]
				}
			}
		}
		id, err := m.store.CreateCall(db.CallInput{
			Address: address, District: f.value(1), CallerName: f.value(3),
			CallerPhone: f.value(4), FireTypeID: tid, Description: f.value(5),
		})
		if err != nil {
			m.status = "Помилка запису: " + err.Error()
			return
		}
		m.status = fmt.Sprintf("Виклик №%d зареєстровано о %s", id, time.Now().Format("15:04 02.01.2006"))
		m.dash.Refresh(m.store)
	}
	return cmd
}

// openReportForm — форма генерації PDF-звіту за період.
func (m *model) openReportForm() tea.Cmd {
	if m.store == nil {
		m.status = "БД недоступна"
		return nil
	}
	now := time.Now()
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

	f, cmd := newSimpleForm("Звіт «Виклики за період» (PDF)", []formField{
		newTextField("З дати (РРРР-ММ-ДД)", "2026-09-01", false),
		newTextField("По дату (РРРР-ММ-ДД)", "2026-09-30", false),
	})
	f.setValue(0, first.Format("2006-01-02"))
	f.setValue(1, now.Format("2006-01-02"))
	m.form = f

	m.onDone = func() {
		from, err1 := time.ParseInLocation("2006-01-02", f.value(0), time.Local)
		to, err2 := time.ParseInLocation("2006-01-02", f.value(1), time.Local)
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
	return cmd
}

// openAIForm — форма запиту до AI-асистента.
func (m *model) openAIForm() tea.Cmd {
	if m.store == nil {
		m.status = "БД недоступна"
		return nil
	}
	if m.cfg.AIKey == "" {
		m.status = "Немає AI-ключа — спочатку Налаштування (клавіша 6)"
		return nil
	}
	f, cmd := newSimpleForm("AI-асистент", []formField{
		newTextField("Питання до даних українською", "скільки викликів по районах цього місяця", false),
	})
	m.form = f

	m.onDone = func() {
		question := f.value(0)
		if question == "" {
			return
		}
		m.status = "[AI] Формую SQL і виконую запит..."
		m.pendingCmd = m.runAI(question)
	}
	return cmd
}

// openSettingsForm — форма налаштувань (зберігається у конфіг 0600).
func (m *model) openSettingsForm() tea.Cmd {
	f, cmd := newSimpleForm("Налаштування", []formField{
		newTextField("Шлях до БД (.accdb)", "data/fire_station.accdb", false),
		newTextField("Шрифт для PDF (TTF)", "assets/fonts/DejaVuSans.ttf", false),
		newTextField("API ключ ШІ aimlapi (порожньо — лишити)", "введи ключ", true),
		newTextField("Модель ШІ", "openai/gpt-5-5", false),
	})
	f.setValue(0, m.cfg.DBPath)
	f.setValue(1, m.cfg.FontPath)
	f.setValue(3, m.cfg.AIModel)
	m.form = f

	m.onDone = func() {
		m.cfg.DBPath = f.value(0)
		m.cfg.FontPath = f.value(1)
		if key := f.value(2); key != "" {
			m.cfg.AIKey = key
		}
		m.cfg.AIModel = f.value(3)
		if err := m.cfg.Save(); err != nil {
			m.status = "Збереження: " + err.Error()
			return
		}
		m.status = "Налаштування збережено (якщо змінено шлях БД — перезапусти програму)"
	}
	return cmd
}
