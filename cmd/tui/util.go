package tui

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// context_ — короткий помічник для фонового контексту запитів.
func context_() context.Context { return context.Background() }

// exeRelative перетворює відносний шлях у шлях відносно теки з exe-файлом.
func exeRelative(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	exe, err := os.Executable()
	if err != nil {
		return p
	}
	return filepath.Join(filepath.Dir(exe), p)
}

// rowsToTable конвертує *sql.Rows у заголовки + рядки рядків для TUI-таблиці.
func rowsToTable(rows *sql.Rows) ([]string, [][]string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	var data [][]string
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		line := make([]string, len(cols))
		for i, v := range vals {
			line[i] = cellText(v)
		}
		data = append(data, line)
	}
	return cols, data, rows.Err()
}

// colWidth підбирає ширину колонки за вмістом (мін. 8, макс. 40).
func colWidth(header string, data [][]string, idx int) int {
	w := len([]rune(header))
	for _, r := range data {
		if idx < len(r) {
			if l := len([]rune(r[idx])); l > w {
				w = l
			}
		}
	}
	if w > 40 {
		w = 40
	}
	if w < 8 {
		w = 8
	}
	return w
}

func cellText(v any) string {
	switch t := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(t)
	case time.Time:
		return t.Format("02.01.2006 15:04")
	case bool:
		if t {
			return "так"
		}
		return "ні"
	default:
		return fmt.Sprint(t)
	}
}
