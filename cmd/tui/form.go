package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// simpleForm — власний легкий компонент форми на bubbles/textinput.
// Жодних зовнішніх форм-бібліотек: повний контроль над клавішами й фокусом,
// тож немає зависань, характерних для вбудовування важких форм.

type fieldKind int

const (
	fieldText fieldKind = iota
	fieldPassword
	fieldSelect
)

type formField struct {
	label   string
	kind    fieldKind
	input   textinput.Model
	options []string // для fieldSelect
	sel     int      // для fieldSelect
}

type simpleForm struct {
	title     string
	fields    []formField
	focus     int
	submitted bool
	cancelled bool
}

// newTextField створює текстове поле (password=true — маскований ввід).
func newTextField(label, placeholder string, password bool) formField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 500
	kind := fieldText
	if password {
		ti.EchoMode = textinput.EchoPassword
		kind = fieldPassword
	}
	return formField{label: label, kind: kind, input: ti}
}

// newSelectField створює поле-перемикач опцій (←/→).
func newSelectField(label string, options []string) formField {
	return formField{label: label, kind: fieldSelect, options: options}
}

// newSimpleForm збирає форму й фокусує перше поле.
func newSimpleForm(title string, fields []formField) (*simpleForm, tea.Cmd) {
	f := &simpleForm{title: title, fields: fields}
	return f, f.focusField(0)
}

func (f *simpleForm) focusField(i int) tea.Cmd {
	for j := range f.fields {
		f.fields[j].input.Blur()
	}
	f.focus = i
	if f.fields[i].kind != fieldSelect {
		return f.fields[i].input.Focus()
	}
	return nil
}

// value повертає значення поля: для select — обрану опцію, для тексту — введене.
func (f *simpleForm) value(i int) string {
	fld := f.fields[i]
	if fld.kind == fieldSelect {
		if len(fld.options) == 0 {
			return ""
		}
		return fld.options[fld.sel]
	}
	return strings.TrimSpace(fld.input.Value())
}

// setValue задає початкове значення текстового поля.
func (f *simpleForm) setValue(i int, v string) {
	f.fields[i].input.SetValue(v)
}

func (f *simpleForm) Update(msg tea.Msg) tea.Cmd {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	cur := &f.fields[f.focus]

	switch k.String() {
	case "esc":
		f.cancelled = true
		return nil
	case "enter":
		if f.focus == len(f.fields)-1 {
			f.submitted = true
			return nil
		}
		return f.focusField(f.focus + 1)
	case "tab", "down":
		if f.focus < len(f.fields)-1 {
			return f.focusField(f.focus + 1)
		}
		return nil
	case "shift+tab", "up":
		if f.focus > 0 {
			return f.focusField(f.focus - 1)
		}
		return nil
	case "left":
		if cur.kind == fieldSelect && cur.sel > 0 {
			cur.sel--
		}
		return nil
	case "right":
		if cur.kind == fieldSelect && cur.sel < len(cur.options)-1 {
			cur.sel++
		}
		return nil
	}

	if cur.kind != fieldSelect {
		var c tea.Cmd
		cur.input, c = cur.input.Update(msg)
		return c
	}
	return nil
}

func (f *simpleForm) View() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(f.title) + "\n\n")
	for i, fld := range f.fields {
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(clrDim))
		if i == f.focus {
			labelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(clrText))
		}
		b.WriteString(labelStyle.Render(fld.label) + "\n")
		if fld.kind == fieldSelect {
			val := fld.options[fld.sel]
			if i == f.focus {
				b.WriteString("  [ " + val + " ]  ←/→ — змінити\n")
			} else {
				b.WriteString("    " + val + "\n")
			}
		} else {
			b.WriteString(fld.input.View() + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(clrFaint)).Render("tab/↑/↓ — між полями • enter — далі/готово • esc — скасувати"))
	return b.String()
}
