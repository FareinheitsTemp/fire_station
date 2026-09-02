package webapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/FareinheitsTemp/fire_station/internal/ai"
	"github.com/FareinheitsTemp/fire_station/internal/config"
	"github.com/FareinheitsTemp/fire_station/internal/db"
	"github.com/FareinheitsTemp/fire_station/internal/report"
)

// API-сервер АІС: міст між веб-фронтендом (Next.js) і БД Access.
// Усі помилки ловляться й повертаються як JSON {error: ...}.

type api struct {
	cfg   *config.Config
	store *db.Store
	dbErr error
}

// Run піднімає БД і HTTP-сервер на 127.0.0.1:8080.
func Run(cfg *config.Config) error {
	a := &api{cfg: cfg}
	dbPath := exeRelative(cfg.DBPath)
	if err := db.EnsureDatabase(dbPath); err != nil {
		a.dbErr = err
	} else if s, err := db.Connect(dbPath); err != nil {
		a.dbErr = err
	} else if err = s.EnsureSchema(); err != nil {
		a.dbErr = err
	} else if err = s.SeedDemo(); err != nil {
		a.dbErr = err
	} else if err = s.EnsureExtras(); err != nil {
		a.dbErr = err
	} else {
		a.store = s
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", a.health)
	mux.HandleFunc("GET /api/stats", a.stats)
	mux.HandleFunc("GET /api/stats/calls-by-day", a.callsByDay)
	mux.HandleFunc("GET /api/recent", a.recent)
	mux.HandleFunc("GET /api/meta", a.metaList)
	mux.HandleFunc("GET /api/ref/{table}", a.refList)
	mux.HandleFunc("GET /api/layout", a.getLayout)
	mux.HandleFunc("PUT /api/layout", a.putLayout)
	mux.HandleFunc("GET /api/tables", a.tables)
	mux.HandleFunc("GET /api/tables/{name}", a.table)
	mux.HandleFunc("POST /api/tables/{name}/rows", a.insertRow)
	mux.HandleFunc("PUT /api/tables/{name}/rows/{id}", a.updateRow)
	mux.HandleFunc("DELETE /api/tables/{name}/rows/{id}", a.deleteRow)
	mux.HandleFunc("GET /api/fire-types", a.fireTypes)
	mux.HandleFunc("POST /api/calls", a.createCall)
	mux.HandleFunc("POST /api/reports/calls", a.reportCalls)
	mux.HandleFunc("GET /api/reports/file/{name}", a.reportFile)
	mux.HandleFunc("POST /api/ai", a.aiQuery)
	mux.HandleFunc("GET /api/config", a.getConfig)
	mux.HandleFunc("PUT /api/config", a.putConfig)

	addr := "127.0.0.1:8080"
	fmt.Println("АІС «Пожежна частина» — API-сервер")
	fmt.Println("Слухаю: http://" + addr)
	if a.dbErr != nil {
		fmt.Println("УВАГА: БД недоступна:", a.dbErr)
	}
	return http.ListenAndServe(addr, withLogging(withRecover(mux)))
}

// --- middleware ---

// withRecover ловить паніку в хендлері й відповідає JSON 500 замість падіння.
func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[panic] %s %s: %v", r.Method, r.URL.Path, rec)
				writeErr(w, http.StatusInternalServerError, "внутрішня помилка сервера")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// withLogging пише в консоль кожен запит: метод, шлях, тривалість.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

// --- helpers ---

func exeRelative(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), p)
	}
	return p
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func (a *api) dbOK(w http.ResponseWriter) bool {
	if a.store == nil {
		msg := "БД недоступна"
		if a.dbErr != nil {
			msg = a.dbErr.Error()
		}
		writeErr(w, http.StatusServiceUnavailable, msg)
		return false
	}
	return true
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeErr(w, http.StatusBadRequest, "невалідний JSON: "+err.Error())
		return false
	}
	return true
}

func scanRows(rows *sql.Rows) ([]string, [][]string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	var data [][]string
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			switch t := v.(type) {
			case nil:
				row[i] = ""
			case []byte:
				row[i] = string(t)
			case time.Time:
				row[i] = t.Format("02.01.2006 15:04")
			default:
				row[i] = fmt.Sprintf("%v", t)
			}
		}
		data = append(data, row)
	}
	return cols, data, rows.Err()
}

var dateLayouts = []string{"2006-01-02", "2006-01-02 15:04", "2006-01-02T15:04", "02.01.2006 15:04", time.RFC3339}

func parseTimeLoose(s string) time.Time {
	for _, l := range []string{"2006-01-02 15:04:05", "2006-01-02 15:04", time.RFC3339, "02.01.2006 15:04"} {
		if t, err := time.ParseInLocation(l, s, time.Local); err == nil {
			return t
		}
	}
	return time.Time{}
}

// rowValues валідує значення за метаданими й конвертує у типи Access.
func rowValues(tm db.TableMeta, vals map[string]string) ([]string, []any, error) {
	known := map[string]bool{}
	for _, c := range tm.Columns {
		known[c.Name] = true
	}
	for k := range vals {
		if !known[k] {
			return nil, nil, fmt.Errorf("невідома колонка: %s", k)
		}
	}
	var cols []string
	var args []any
	for _, c := range tm.Columns {
		raw, present := vals[c.Name]
		raw = strings.TrimSpace(raw)
		if c.Required && raw == "" {
			return nil, nil, fmt.Errorf("обов'язкове поле: %s", c.Label)
		}
		if !present || raw == "" {
			continue // лишаємо NULL/дефолт
		}
		var v any
		var err error
		switch c.Type {
		case "number":
			v, err = strconv.ParseFloat(strings.ReplaceAll(raw, ",", "."), 64)
		case "ref":
			v, err = strconv.ParseInt(raw, 10, 64)
		case "bool":
			l := strings.ToLower(raw)
			if l == "1" || l == "true" || l == "так" {
				v = 1
			} else {
				v = 0
			}
		case "date":
			for _, layout := range dateLayouts {
				if t, perr := time.ParseInLocation(layout, raw, time.Local); perr == nil {
					v = t
					break
				}
			}
			if v == nil {
				err = fmt.Errorf("невірна дата: %s", raw)
			}
		case "select":
			okOpt := false
			for _, o := range c.Options {
				if o == raw {
					okOpt = true
				}
			}
			if !okOpt {
				err = fmt.Errorf("недопустиме значення %q", raw)
			} else {
				v = raw
			}
		default:
			v = raw
		}
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %v", c.Label, err)
		}
		cols = append(cols, c.Name)
		args = append(args, v)
	}
	return cols, args, nil
}

// --- базові handlers ---

func (a *api) health(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{"ok": true, "db": a.store != nil}
	if a.dbErr != nil {
		resp["error"] = a.dbErr.Error()
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *api) stats(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	st, err := a.store.Stats(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// callsByDay — кількість викликів по днях за ?days= (1..90, дефолт 14).
func (a *api) callsByDay(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	days := 14
	if d := r.URL.Query().Get("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n >= 1 && n <= 90 {
			days = n
		}
	}
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, -(days - 1))

	rows, err := a.store.Query(r.Context(), "SELECT call_at FROM calls WHERE call_at >= ?", start)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var v any
		if err := rows.Scan(&v); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		var t time.Time
		switch x := v.(type) {
		case time.Time:
			t = x
		case []byte:
			t = parseTimeLoose(string(x))
		case string:
			t = parseTimeLoose(x)
		}
		if !t.IsZero() {
			counts[t.Format("2006-01-02")]++
		}
	}

	out := make([]map[string]any, 0, days)
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		out = append(out, map[string]any{"day": d.Format("02.01"), "count": counts[d.Format("2006-01-02")]})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *api) recent(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	list, err := a.store.RecentCalls(r.Context(), 10)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *api) metaList(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, db.TablesMeta())
}

func (a *api) refList(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	tm, ok := db.TableMetaByName(r.PathValue("table"))
	if !ok {
		writeErr(w, http.StatusNotFound, "невідома таблиця")
		return
	}
	rows, err := a.store.Query(r.Context(), fmt.Sprintf("SELECT TOP 500 [%s], [%s] FROM [%s]", tm.PK, tm.LabelCol, tm.Name))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	_, data, err := scanRows(rows)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]map[string]string, len(data))
	for i, row := range data {
		out[i] = map[string]string{"id": row[0], "label": row[1]}
	}
	writeJSON(w, http.StatusOK, out)
}

// --- розкладка графа структури ---

func (a *api) getLayout(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	writeJSON(w, http.StatusOK, a.store.Layouts())
}

type layoutReq struct {
	Node string  `json:"node"`
	DX   float64 `json:"dx"`
	DY   float64 `json:"dy"`
}

func (a *api) putLayout(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	var req layoutReq
	if !decode(w, r, &req) || strings.TrimSpace(req.Node) == "" {
		writeErr(w, http.StatusBadRequest, "порожнє ім'я елемента")
		return
	}
	if err := a.store.SaveLayout(req.Node, req.DX, req.DY); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *api) tables(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, db.TableNames())
}

func (a *api) table(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	name := r.PathValue("name")
	if _, ok := db.TableMetaByName(name); !ok {
		writeErr(w, http.StatusNotFound, "невідома таблиця")
		return
	}
	rows, err := a.store.Query(r.Context(), fmt.Sprintf("SELECT TOP 500 * FROM [%s]", name))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	cols, data, err := scanRows(rows)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"columns": cols, "rows": data})
}

// --- CRUD ---

type rowReq struct {
	Values map[string]string `json:"values"`
}

func (a *api) insertRow(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	tm, ok := db.TableMetaByName(r.PathValue("name"))
	if !ok {
		writeErr(w, http.StatusNotFound, "невідома таблиця")
		return
	}
	var req rowReq
	if !decode(w, r, &req) {
		return
	}
	cols, args, err := rowValues(tm, req.Values)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(cols) == 0 {
		writeErr(w, http.StatusBadRequest, "немає даних для вставки")
		return
	}
	q := fmt.Sprintf("INSERT INTO [%s] ([%s]) VALUES (%s)",
		tm.Name, strings.Join(cols, "],["), strings.TrimSuffix(strings.Repeat("?,", len(cols)), ","))
	if _, err := a.store.ExecContext(r.Context(), q, args...); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *api) updateRow(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	tm, ok := db.TableMetaByName(r.PathValue("name"))
	if !ok {
		writeErr(w, http.StatusNotFound, "невідома таблиця")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "невірний id")
		return
	}
	var req rowReq
	if !decode(w, r, &req) {
		return
	}
	cols, args, err := rowValues(tm, req.Values)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(cols) == 0 {
		writeErr(w, http.StatusBadRequest, "немає даних для оновлення")
		return
	}
	set := make([]string, len(cols))
	for i, c := range cols {
		set[i] = "[" + c + "] = ?"
	}
	args = append(args, id)
	q := fmt.Sprintf("UPDATE [%s] SET %s WHERE [%s] = ?", tm.Name, strings.Join(set, ", "), tm.PK)
	if _, err := a.store.ExecContext(r.Context(), q, args...); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *api) deleteRow(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	tm, ok := db.TableMetaByName(r.PathValue("name"))
	if !ok {
		writeErr(w, http.StatusNotFound, "невідома таблиця")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "невірний id")
		return
	}
	q := fmt.Sprintf("DELETE FROM [%s] WHERE [%s] = ?", tm.Name, tm.PK)
	if _, err := a.store.ExecContext(r.Context(), q, id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// --- довідники, виклики, звіти, ШІ, конфіг ---

func (a *api) fireTypes(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	list, err := a.store.FireTypes()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

type callReq struct {
	Address     string `json:"address"`
	District    string `json:"district"`
	CallerName  string `json:"caller_name"`
	CallerPhone string `json:"caller_phone"`
	FireTypeID  int64  `json:"fire_type_id"`
	Description string `json:"description"`
}

func (a *api) createCall(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	var req callReq
	if !decode(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Address) == "" {
		writeErr(w, http.StatusBadRequest, "адреса порожня")
		return
	}
	id, err := a.store.CreateCall(db.CallInput{
		Address: req.Address, District: req.District, CallerName: req.CallerName,
		CallerPhone: req.CallerPhone, FireTypeID: req.FireTypeID, Description: req.Description,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id})
}

type reportReq struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (a *api) reportCalls(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	var req reportReq
	if !decode(w, r, &req) {
		return
	}
	from, err1 := time.ParseInLocation("2006-01-02", req.From, time.Local)
	to, err2 := time.ParseInLocation("2006-01-02", req.To, time.Local)
	if err1 != nil || err2 != nil {
		writeErr(w, http.StatusBadRequest, "невірний формат дати (РРРР-ММ-ДД)")
		return
	}
	to = to.Add(24*time.Hour - time.Second)

	rows, err := a.store.CallsByPeriod(r.Context(), from, to)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(rows) == 0 {
		writeErr(w, http.StatusNotFound, "за цей період викликів немає")
		return
	}
	out, err := report.CallsByPeriodPDF(exeRelative(a.cfg.FontPath), exeRelative("reports"), from, to, rows)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"file": filepath.Base(out)})
}

func (a *api) reportFile(w http.ResponseWriter, r *http.Request) {
	base := filepath.Base(r.PathValue("name"))
	if !strings.HasSuffix(base, ".pdf") {
		writeErr(w, http.StatusBadRequest, "лише .pdf")
		return
	}
	http.ServeFile(w, r, filepath.Join(exeRelative("reports"), base))
}

type aiReq struct {
	Question string `json:"question"`
}

func (a *api) aiQuery(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	if a.cfg.AIKey == "" {
		writeErr(w, http.StatusBadRequest, "немає AI-ключа — додай його на сторінці «Налаштування»")
		return
	}
	var req aiReq
	if !decode(w, r, &req) || strings.TrimSpace(req.Question) == "" {
		writeErr(w, http.StatusBadRequest, "питання порожнє")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	// Схема + база знань як контекст для ШІ
	schemaDesc := db.SchemaDescription()
	if rules, err := a.store.KnowledgeRules(ctx); err == nil && len(rules) > 0 {
		var sb strings.Builder
		sb.WriteString(schemaDesc)
		sb.WriteString("\n\nБаза знань (правила реагування, для контексту):\n")
		for _, r := range rules {
			fmt.Fprintf(&sb, "- [%s] %s: якщо %s → %s\n", r.Priority, r.Topic, r.Condition, r.Recommendation)
		}
		schemaDesc = sb.String()
	}

	client := ai.NewClient(a.cfg.AIKey, a.cfg.AIModel)
	sqlText, err := client.GenerateSQL(ctx, req.Question, schemaDesc)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	if !ai.IsSafeSelect(sqlText) {
		writeErr(w, http.StatusForbidden, "запит відхилено політикою безпеки (лише SELECT)")
		return
	}
	rows, err := a.store.Query(ctx, sqlText)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, fmt.Sprintf("виконання: %v (SQL: %s)", err, sqlText))
		return
	}
	defer rows.Close()
	cols, data, err := scanRows(rows)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sql": sqlText, "columns": cols, "rows": data})
}

func (a *api) getConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"db_path": a.cfg.DBPath, "font_path": a.cfg.FontPath,
		"ai_model": a.cfg.AIModel, "has_ai_key": a.cfg.AIKey != "",
	})
}

type configReq struct {
	DBPath   string `json:"db_path"`
	FontPath string `json:"font_path"`
	AIKey    string `json:"ai_key"`
	AIModel  string `json:"ai_model"`
}

func (a *api) putConfig(w http.ResponseWriter, r *http.Request) {
	var req configReq
	if !decode(w, r, &req) {
		return
	}
	if req.DBPath != "" {
		a.cfg.DBPath = req.DBPath
	}
	if req.FontPath != "" {
		a.cfg.FontPath = req.FontPath
	}
	if req.AIKey != "" {
		a.cfg.AIKey = req.AIKey
	}
	if req.AIModel != "" {
		a.cfg.AIModel = req.AIModel
	}
	if err := a.cfg.Save(); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
