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
	"sync"
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
	errs  *errLog
}

// --- журнал помилок (кільцевий буфер для авто-моду) ---

type errEntry struct {
	Time   time.Time `json:"time"`
	Method string    `json:"method"`
	Path   string    `json:"path"`
	Status int       `json:"status"`
}

type errLog struct {
	mu    sync.Mutex
	items []errEntry
}

func (l *errLog) add(e errEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, e)
	if len(l.items) > 30 {
		l.items = l.items[len(l.items)-30:]
	}
}

func (l *errLog) snapshot() []errEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]errEntry, len(l.items))
	copy(out, l.items)
	return out
}

// Run піднімає БД і HTTP-сервер на 127.0.0.1:8080.
func Run(cfg *config.Config) error {
	a := &api{cfg: cfg, errs: &errLog{}}
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
	mux.HandleFunc("POST /api/chat", a.chat)
	mux.HandleFunc("POST /api/agent/analyze", a.agentAnalyze)
	mux.HandleFunc("GET /api/config", a.getConfig)
	mux.HandleFunc("PUT /api/config", a.putConfig)

	addr := "127.0.0.1:8080"
	fmt.Println("АІС «Пожежна частина» — API-сервер")
	fmt.Println("Слухаю: http://" + addr)
	if a.dbErr != nil {
		fmt.Println("УВАГА: БД недоступна:", a.dbErr)
	}
	return http.ListenAndServe(addr, a.withLogging(withRecover(mux)))
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

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// withLogging логує кожен запит у консоль; статуси 4xx/5xx потрапляють у буфер для авто-моду.
func (a *api) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s → %d (%s)", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Millisecond))
		if rec.status >= 400 {
			a.errs.add(errEntry{Time: time.Now(), Method: r.Method, Path: r.URL.Path, Status: rec.status})
		}
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

func (a *api) aiOK(w http.ResponseWriter) bool {
	if a.cfg.AIKey == "" {
		writeErr(w, http.StatusBadRequest, "немає AI-ключа — додай його на сторінці «Налаштування»")
		return false
	}
	return true
}

func (a *api) aiClient() *ai.Client {
	return ai.NewClient(a.cfg.AIBaseURL, a.cfg.AIKey, a.cfg.AIModel)
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
			continue
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

// --- довідники, виклики, звіти ---

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

// --- ШІ: NL→SQL, чат, авто-мод ---

// systemContext: схема + правила БЗ (+ статистика для чату/агента).
func (a *api) systemContext(ctx context.Context, withStats bool) string {
	var sb strings.Builder
	sb.WriteString(db.SchemaDescription())
	if rules, err := a.store.KnowledgeRules(ctx); err == nil && len(rules) > 0 {
		sb.WriteString("\n\nБаза знань (чинні правила):\n")
		for _, r := range rules {
			fmt.Fprintf(&sb, "- [%s] %s: якщо %s → %s\n", r.Priority, r.Topic, r.Condition, r.Recommendation)
		}
	}
	if withStats {
		if st, err := a.store.Stats(ctx); err == nil {
			fmt.Fprintf(&sb, "\nПоточна статистика: викликів усього %d, сьогодні %d, працівників активно %d, техніки в строю %d.\n",
				st.TotalCalls, st.TodayCalls, st.ActiveEmployees, st.EquipmentOK)
		}
	}
	return sb.String()
}

type aiReq struct {
	Question string `json:"question"`
}

func (a *api) aiQuery(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) || !a.aiOK(w) {
		return
	}
	var req aiReq
	if !decode(w, r, &req) || strings.TrimSpace(req.Question) == "" {
		writeErr(w, http.StatusBadRequest, "питання порожнє")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	sqlText, err := a.aiClient().GenerateSQL(ctx, req.Question, a.systemContext(ctx, false))
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

type chatReq struct {
	Messages []ai.Message `json:"messages"`
}

// chat — розмова з агентом (контекст: схема БД + правила БЗ + статистика).
func (a *api) chat(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) || !a.aiOK(w) {
		return
	}
	var req chatReq
	if !decode(w, r, &req) || len(req.Messages) == 0 {
		writeErr(w, http.StatusBadRequest, "порожня розмова")
		return
	}
	if len(req.Messages) > 40 {
		req.Messages = req.Messages[len(req.Messages)-40:]
	}

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	sys := "Ти — агент АІС «Пожежна частина». Відповідай українською, коротко і по суті. " +
		"Ти знаєш структуру бази даних, чинні правила бази знань і поточну статистику частини. " +
		"Якщо питання про конкретні числа — спирайся на статистику; якщо треба глибший аналіз — порадь скористатись AI-асистентом або авто-модулем.\n\n" +
		a.systemContext(ctx, true)

	msgs := append([]ai.Message{{Role: "system", Content: sys}}, req.Messages...)
	reply, err := a.aiClient().Chat(ctx, msgs)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"reply": reply})
}

// agentAnalyze — авто-мод: автономний аналіз стану системи й самостійне
// створення нових правил у базі знань, якщо щось не так або бракує покриття.
func (a *api) agentAnalyze(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) || !a.aiOK(w) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	stats, _ := a.store.Stats(ctx)
	recent, _ := a.store.RecentCalls(ctx, 10)
	rules, _ := a.store.KnowledgeRules(ctx)
	snapshot := map[string]any{
		"stats": stats, "recent_calls": recent,
		"existing_rules": rules, "recent_api_errors": a.errs.snapshot(),
	}
	rawSnap, _ := json.MarshalIndent(snapshot, "", "  ")

	sys := `Ти — автономний агент-аналітик АІС «Пожежна частина». Твоє завдання:
1. Проаналізуй знімок системи: статистику, останні виклики, чинні правила бази знань, останні помилки API.
2. Знайди проблеми, аномалії чи прогалини в покритті правил.
3. Якщо треба — запропонуй НОВІ правила (не дублюй чинні).
Відповідай СТРОГО валідним JSON без markdown і без зайвого тексту:
{"conclusion": "твій висновок українською (2-4 речення)",
 "new_rules": [{"topic": "...", "category": "реагування|техніка|персонал|безпека", "condition": "якщо...", "recommendation": "то...", "priority": "низький|середній|високий"}]}
Якщо нових правил не треба — поверни порожній масив new_rules.`

	out, err := a.aiClient().Chat(ctx, []ai.Message{
		{Role: "system", Content: sys},
		{Role: "user", Content: "Знімок системи:\n" + string(rawSnap)},
	})
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	type ruleIn struct {
		Topic          string `json:"topic"`
		Category       string `json:"category"`
		Condition      string `json:"condition"`
		Recommendation string `json:"recommendation"`
		Priority       string `json:"priority"`
	}
	var reply struct {
		Conclusion string   `json:"conclusion"`
		NewRules   []ruleIn `json:"new_rules"`
	}
	clean := ai.StripCodeFences(out)
	if i := strings.Index(clean, "{"); i > 0 {
		clean = clean[i:]
	}
	if err := json.Unmarshal([]byte(clean), &reply); err != nil {
		writeErr(w, http.StatusBadGateway, "агент повернув невалідний JSON: "+err.Error())
		return
	}

	added := 0
	for _, nr := range reply.NewRules {
		if strings.TrimSpace(nr.Topic) == "" {
			continue
		}
		cat, prio := nr.Category, nr.Priority
		if cat == "" {
			cat = "висновки"
		}
		if prio == "" {
			prio = "середній"
		}
		if err := a.store.InsertRule(nr.Topic, cat, nr.Condition, nr.Recommendation, prio); err == nil {
			added++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"conclusion":  reply.Conclusion,
		"new_rules":   reply.NewRules,
		"rules_added": added,
		"rules_total": len(rules) + added,
	})
}

// --- конфіг ---

func (a *api) getConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"db_path": a.cfg.DBPath, "font_path": a.cfg.FontPath,
		"ai_model": a.cfg.AIModel, "ai_base_url": a.cfg.AIBaseURL, "has_ai_key": a.cfg.AIKey != "",
	})
}

type configReq struct {
	DBPath    string `json:"db_path"`
	FontPath  string `json:"font_path"`
	AIKey     string `json:"ai_key"`
	AIModel   string `json:"ai_model"`
	AIBaseURL string `json:"ai_base_url"`
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
	if req.AIBaseURL != "" {
		a.cfg.AIBaseURL = req.AIBaseURL
	}
	if err := a.cfg.Save(); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
