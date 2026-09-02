package webapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FareinheitsTemp/fire_station/internal/ai"
	"github.com/FareinheitsTemp/fire_station/internal/config"
	"github.com/FareinheitsTemp/fire_station/internal/db"
	"github.com/FareinheitsTemp/fire_station/internal/report"
)

// API-сервер АІС: віддає JSON-ендпоінти для веб-фронтенду (Next.js).
// Бекенд — консольний процес; UI повністю у браузері.

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
	} else {
		a.store = s
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", a.health)
	mux.HandleFunc("GET /api/stats", a.stats)
	mux.HandleFunc("GET /api/recent", a.recent)
	mux.HandleFunc("GET /api/tables", a.tables)
	mux.HandleFunc("GET /api/tables/{name}", a.table)
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
	return http.ListenAndServe(addr, mux)
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

// --- handlers ---

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

func (a *api) tables(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, db.TableNames())
}

func (a *api) table(w http.ResponseWriter, r *http.Request) {
	if !a.dbOK(w) {
		return
	}
	name := r.PathValue("name")
	allowed := false
	for _, n := range db.TableNames() {
		if n == name {
			allowed = true
		}
	}
	if !allowed {
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

	client := ai.NewClient(a.cfg.AIKey, a.cfg.AIModel)
	sqlText, err := client.GenerateSQL(ctx, req.Question, db.SchemaDescription())
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
