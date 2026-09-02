package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/FareinheitsTemp/fire_station/internal/report"
	_ "github.com/alexbrainman/odbc"
)

// accDriver — ODBC-драйвер ACE (йде з Office або з безкоштовним
// Access Database Engine Redistributable).
const accDriver = "Microsoft Access Driver (*.mdb, *.accdb)"

// Store — шар доступу до файлу Access (.accdb).
type Store struct {
	db *sql.DB
}

// EnsureDatabase створює порожній файл .accdb, якщо його ще немає.
// ODBC не вміє створювати файли, тому використовуємо ADOX через PowerShell COM.
func EnsureDatabase(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	script := fmt.Sprintf(
		`$c = New-Object -ComObject ADOX.Catalog; $c.Create("Provider=Microsoft.ACE.OLEDB.12.0;Data Source='%s'") | Out-Null`,
		abs,
	)
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("створення %s через ADOX: %w (%s)\n(потрібен Access Database Engine: безкоштовний редистрібутив Microsoft)", abs, err, out)
	}
	return nil
}

// Connect відкриває .accdb через ODBC.
func Connect(path string) (*Store, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	connStr := fmt.Sprintf("Driver={%s};DBQ=%s;", accDriver, abs)
	sqlDB, err := sql.Open("odbc", connStr)
	if err != nil {
		return nil, fmt.Errorf("odbc open: %w", err)
	}
	// Access ODBC погано переносить паралельні з'єднання — працюємо в одному.
	sqlDB.SetMaxOpenConns(1)
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("odbc ping: %w (перевір, чи встановлений драйвер «%s»)", err, accDriver)
	}
	return &Store{db: sqlDB}, nil
}

// Close закриває з'єднання.
func (s *Store) Close() error {
	return s.db.Close()
}

// Query виконує довільний SELECT (для довідників та AI-асистента).
func (s *Store) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

// FireType — довідник типів пожеж.
type FireType struct {
	ID   int64
	Name string
}

// FireTypes повертає довідник типів пожеж для форми виклику.
func (s *Store) FireTypes() ([]FireType, error) {
	rows, err := s.db.Query("SELECT id, name FROM fire_types ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []FireType
	for rows.Next() {
		var t FireType
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// CallInput — дані форми реєстрації виклику.
type CallInput struct {
	Address     string
	District    string
	CallerName  string
	CallerPhone string
	FireTypeID  int64
	Description string
}

// CreateCall реєструє новий виклик і повертає його номер (@@IDENTITY).
func (s *Store) CreateCall(c CallInput) (int64, error) {
	_, err := s.db.Exec(
		`INSERT INTO calls (call_at, address, district, caller_name, caller_phone, fire_type_id, status, description)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		time.Now(), c.Address, c.District, c.CallerName, c.CallerPhone, c.FireTypeID, "новий", c.Description,
	)
	if err != nil {
		return 0, err
	}
	var raw any
	if err := s.db.QueryRow("SELECT @@IDENTITY").Scan(&raw); err != nil {
		return 0, nil // номер некритичний, запис уже створено
	}
	switch v := raw.(type) {
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	}
	return 0, nil
}

// CallsByPeriod — вибірка викликів за період для PDF-звіту.
func (s *Store) CallsByPeriod(ctx context.Context, from, to time.Time) ([]report.CallRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.call_at, c.address, c.district, f.name, c.status
		 FROM calls c LEFT JOIN fire_types f ON c.fire_type_id = f.id
		 WHERE c.call_at >= ? AND c.call_at <= ?
		 ORDER BY c.call_at DESC`,
		from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []report.CallRow
	for rows.Next() {
		var r report.CallRow
		var district, ftype sql.NullString
		if err := rows.Scan(&r.CallAt, &r.Address, &district, &ftype, &r.Status); err != nil {
			return nil, err
		}
		r.District = district.String
		r.FireType = ftype.String
		out = append(out, r)
	}
	return out, rows.Err()
}

// Stats — зведена статистика для дашборда.
type Stats struct {
	TotalCalls      int
	TodayCalls      int
	ActiveEmployees int
	EquipmentOK     int
}

// Stats збирає лічильники для головної сторінки.
func (s *Store) Stats(ctx context.Context) (Stats, error) {
	var st Stats
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls").Scan(&st.TotalCalls); err != nil {
		return st, err
	}
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls WHERE call_at >= ?", today).Scan(&st.TodayCalls); err != nil {
		return st, err
	}
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM employees WHERE is_active = True").Scan(&st.ActiveEmployees); err != nil {
		return st, err
	}
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM equipment WHERE status = 'в строю'").Scan(&st.EquipmentOK); err != nil {
		return st, err
	}
	return st, nil
}

// RecentCall — останній виклик для дашборда.
type RecentCall struct {
	CallAt  time.Time
	Address string
	Status  string
}

// RecentCalls повертає останні n викликів.
func (s *Store) RecentCalls(ctx context.Context, n int) ([]RecentCall, error) {
	rows, err := s.db.QueryContext(ctx,
		fmt.Sprintf("SELECT TOP %d call_at, address, status FROM calls ORDER BY call_at DESC", n),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RecentCall
	for rows.Next() {
		var rc RecentCall
		if err := rows.Scan(&rc.CallAt, &rc.Address, &rc.Status); err != nil {
			return nil, err
		}
		out = append(out, rc)
	}
	return out, rows.Err()
}
