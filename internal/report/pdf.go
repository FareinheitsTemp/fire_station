package report

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// CallRow — рядок звіту «Виклики за період».
type CallRow struct {
	CallAt   time.Time
	Address  string
	District string
	FireType string
	Status   string
}

// CallsByPeriodPDF генерує PDF-звіт про виклики за період.
// fontPath — шлях до TTF-шрифту з підтримкою кирилиці (DejaVuSans.ttf).
func CallsByPeriodPDF(fontPath, outDir string, from, to time.Time, rows []CallRow) (string, error) {
	if _, err := os.Stat(fontPath); err != nil {
		return "", fmt.Errorf("не знайдено шрифт %s — завантаж DejaVuSans.ttf (див. assets/fonts/README.txt)", fontPath)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8Font("dejavu", "", fontPath)
	pdf.AddPage()

	pdf.SetFont("dejavu", "", 16)
	pdf.CellFormat(0, 10, "Звіт: виклики пожежної частини", "", 1, "C", false, 0, "")
	pdf.SetFont("dejavu", "", 11)
	pdf.CellFormat(0, 8, fmt.Sprintf("Період: %s — %s", from.Format("02.01.2006"), to.Format("02.01.2006")), "", 1, "C", false, 0, "")
	pdf.Ln(4)

	headers := []string{"Дата/час", "Адреса", "Район", "Тип", "Статус"}
	widths := []float64{32, 62, 30, 34, 32}

	pdf.SetFont("dejavu", "", 10)
	pdf.SetFillColor(230, 230, 230)
	for i, h := range headers {
		pdf.CellFormat(widths[i], 7, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	for _, r := range rows {
		pdf.CellFormat(widths[0], 6, r.CallAt.Format("02.01 15:04"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[1], 6, trunc(r.Address, 30), "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[2], 6, trunc(r.District, 14), "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[3], 6, trunc(r.FireType, 16), "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[4], 6, trunc(r.Status, 14), "1", 1, "L", false, 0, "")
	}

	pdf.Ln(4)
	pdf.SetFont("dejavu", "", 11)
	pdf.CellFormat(0, 8, fmt.Sprintf("Усього викликів: %d", len(rows)), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 8, "Сформовано: "+time.Now().Format("02.01.2006 15:04"), "", 1, "L", false, 0, "")

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	out := filepath.Join(outDir, fmt.Sprintf("calls_%s_%s.pdf",
		from.Format("20060102"), to.Format("20060102")))
	if err := pdf.OutputFileAndClose(out); err != nil {
		return "", err
	}
	return out, nil
}

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "..."
}
