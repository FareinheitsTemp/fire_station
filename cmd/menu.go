package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/FareinheitsTemp/fire_station/internal/ai"
	"github.com/FareinheitsTemp/fire_station/internal/db"
	"github.com/FareinheitsTemp/fire_station/internal/report"
	"github.com/charmbracelet/lipgloss"
)

var store *db.Store

// runInteractive — головне меню програми (запуск exe без аргументів).
func runInteractive() error {
	banner := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("202")).Render("=== АІС «Пожежна частина» ===")
	fmt.Println(banner)
	fmt.Printf("БД: %s | ШІ: %s\n\n", exeRelative(cfg.DBPath), statusOf(cfg.AIKey))

	for {
		var action string
		err := survey.AskOne(&survey.Select{
			Message: "Головне меню:",
			Options: []string{
				"Довідники (перегляд таблиць)",
				"Зареєструвати виклик",
				"Звіти (PDF)",
				"AI-асистент",
				"Налаштування",
				"Вихід",
			},
		}, &action)
		if err != nil {
			return nil // Ctrl+C — тихий вихід
		}

		switch {
		case strings.HasPrefix(action, "Довідники"):
			menuDictionaries()
		case strings.HasPrefix(action, "Зареєструвати"):
			menuNewCall()
		case strings.HasPrefix(action, "Звіти"):
			menuReports()
		case strings.HasPrefix(action, "AI"):
			menuAI()
		case strings.HasPrefix(action, "Налаштування"):
			menuSettings()
		default:
			return nil
		}
		fmt.Println()
	}
}

// getStore ліниво відкриває БД: створює файл .accdb, схему та демо-дані за потреби.
// Шлях до файлу рахується від теки, де лежить exe.
func getStore() (*db.Store, error) {
	if store != nil {
		return store, nil
	}
	dbPath := exeRelative(cfg.DBPath)
	if err := db.EnsureDatabase(dbPath); err != nil {
		return nil, err
	}
	s, err := db.Connect(dbPath)
	if err != nil {
		return nil, err
	}
	if err := s.EnsureSchema(); err != nil {
		s.Close()
		return nil, err
	}
	if err := s.SeedDemo(); err != nil {
		s.Close()
		return nil, err
	}
	store = s
	return store, nil
}

func menuDictionaries() {
	s, err := getStore()
	if err != nil {
		fmt.Println("БД:", err)
		return
	}
	var tableName string
	if err := survey.AskOne(&survey.Select{
		Message: "Яка таблиця?",
		Options: db.TableNames(),
	}, &tableName); err != nil {
		return
	}
	rows, err := s.Query(context.Background(), fmt.Sprintf("SELECT TOP 200 * FROM [%s]", tableName))
	if err != nil {
		fmt.Println("Запит:", err)
		return
	}
	defer rows.Close()

	headers, data, err := rowsToTable(rows)
	if err != nil {
		fmt.Println("Вивід:", err)
		return
	}
	if err := showTable("Таблиця: "+tableName, headers, data); err != nil {
		fmt.Println("TUI:", err)
	}
}

func menuNewCall() {
	s, err := getStore()
	if err != nil {
		fmt.Println("БД:", err)
		return
	}

	types, err := s.FireTypes()
	if err != nil {
		fmt.Println("Довідник типів:", err)
		return
	}
	typeNames := make([]string, len(types))
	for i, t := range types {
		typeNames[i] = t.Name
	}

	ans := struct {
		Address  string
		District string
		Caller   string
		Phone    string
		Descr    string
	}{}
	qs := []*survey.Question{
		{Name: "address", Prompt: &survey.Input{Message: "Адреса виклику:"}, Validate: survey.Required},
		{Name: "district", Prompt: &survey.Input{Message: "Район:"}},
		{Name: "caller", Prompt: &survey.Input{Message: "Заявник (ПІБ):"}},
		{Name: "phone", Prompt: &survey.Input{Message: "Телефон заявника:"}},
		{Name: "descr", Prompt: &survey.Input{Message: "Опис ситуації:"}},
	}
	if err := survey.Ask(qs, &ans); err != nil {
		return
	}

	var typeName string
	if err := survey.AskOne(&survey.Select{Message: "Тип пожежі:", Options: typeNames}, &typeName); err != nil {
		return
	}
	var typeID int64
	for _, t := range types {
		if t.Name == typeName {
			typeID = t.ID
		}
	}

	id, err := s.CreateCall(db.CallInput{
		Address:     ans.Address,
		District:    ans.District,
		CallerName:  ans.Caller,
		CallerPhone: ans.Phone,
		FireTypeID:  typeID,
		Description: ans.Descr,
	})
	if err != nil {
		fmt.Println("Запис:", err)
		return
	}
	fmt.Printf("Виклик №%d зареєстровано о %s\n", id, time.Now().Format("15:04 02.01.2006"))
}

func menuReports() {
	s, err := getStore()
	if err != nil {
		fmt.Println("БД:", err)
		return
	}

	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

	ans := struct{ From, To string }{}
	qs := []*survey.Question{
		{Name: "from", Prompt: &survey.Input{Message: "З дати (РРРР-ММ-ДД):", Default: firstOfMonth.Format("2006-01-02")}, Validate: survey.Required},
		{Name: "to", Prompt: &survey.Input{Message: "По дату (РРРР-ММ-ДД):", Default: now.Format("2006-01-02")}, Validate: survey.Required},
	}
	if err := survey.Ask(qs, &ans); err != nil {
		return
	}

	from, err1 := time.ParseInLocation("2006-01-02", ans.From, time.Local)
	to, err2 := time.ParseInLocation("2006-01-02", ans.To, time.Local)
	if err1 != nil || err2 != nil {
		fmt.Println("Невірний формат дати. Приклад: 2026-09-01")
		return
	}
	to = to.Add(24*time.Hour - time.Second)

	rows, err := s.CallsByPeriod(context.Background(), from, to)
	if err != nil {
		fmt.Println("Запит:", err)
		return
	}
	if len(rows) == 0 {
		fmt.Println("За цей період викликів немає")
		return
	}

	out, err := report.CallsByPeriodPDF(exeRelative(cfg.FontPath), exeRelative("reports"), from, to, rows)
	if err != nil {
		fmt.Println("PDF:", err)
		return
	}
	fmt.Println("Звіт збережено:", out)
}

func menuAI() {
	if cfg.AIKey == "" {
		fmt.Println("API-ключ ШІ не налаштовано: Налаштування -> API ключ ШІ")
		return
	}
	s, err := getStore()
	if err != nil {
		fmt.Println("БД:", err)
		return
	}

	var question string
	if err := survey.AskOne(&survey.Input{
		Message: "Питання до даних (наприклад: «скільки викликів по районах цього місяця»):",
	}, &question, survey.WithValidator(survey.Required)); err != nil {
		return
	}

	client := ai.NewClient(cfg.AIKey, cfg.AIModel)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	fmt.Println("[AI] Формую SQL-запит...")
	sqlText, err := client.GenerateSQL(ctx, question, db.SchemaDescription())
	if err != nil {
		fmt.Println("[AI]", err)
		return
	}
	fmt.Println("[AI] SQL:", sqlText)

	if !ai.IsSafeSelect(sqlText) {
		fmt.Println("[AI] Запит відхилено політикою безпеки (дозволено лише SELECT)")
		return
	}

	rows, err := s.Query(ctx, sqlText)
	if err != nil {
		fmt.Println("Помилка виконання:", err)
		return
	}
	defer rows.Close()

	headers, data, err := rowsToTable(rows)
	if err != nil {
		fmt.Println("Вивід:", err)
		return
	}
	if err := showTable("[AI] Результат запиту", headers, data); err != nil {
		fmt.Println("TUI:", err)
	}
}

func menuSettings() {
	ans := struct {
		DBPath   string
		FontPath string
		AIKey    string
		AIModel  string
	}{}
	qs := []*survey.Question{
		{Name: "dbpath", Prompt: &survey.Input{Message: "Шлях до файлу БД (.accdb):", Default: cfg.DBPath}},
		{Name: "font", Prompt: &survey.Input{Message: "Шрифт для PDF (TTF):", Default: cfg.FontPath}},
		{Name: "aikey", Prompt: &survey.Password{Message: "API ключ ШІ aimlapi (Enter — лишити):"}},
		{Name: "aimodel", Prompt: &survey.Input{Message: "Модель ШІ:", Default: cfg.AIModel}},
	}
	if err := survey.Ask(qs, &ans); err != nil {
		return
	}

	cfg.DBPath = ans.DBPath
	cfg.FontPath = ans.FontPath
	if ans.AIKey != "" {
		cfg.AIKey = ans.AIKey
	}
	cfg.AIModel = ans.AIModel

	if err := cfg.Save(); err != nil {
		fmt.Println("Збереження:", err)
		return
	}
	fmt.Println("Налаштування збережено. Наступні запуски підхоплять їх автоматично.")
}

func statusOf(v string) string {
	if v == "" {
		return "вимкнено"
	}
	return "OK"
}
