package tableprint

import (
	"io"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Timezone string

const (
	Timezone_Local = "local"
	Timezone_UTC   = "utc"
)

// Format for CLI output
type Format string

const (
	Format_Table     = "table"
	Format_ValueOnly = "value-only"
)

type CommonTablePrintArgs struct {
	Format          Format
	Mask            bool
	Tz              Timezone
	W               io.Writer
	DesiredMaxWidth int
}

func mask(mask bool, val string) string {
	if mask {
		if len(val) < 2 {
			return "**"
		} else {
			return val[:2] + "****"
		}
	}
	return val
}

// truncate truncates a string to max-3 characters and appends "..." if the string is longer than max. If max < 3, it returns the original string.
func truncate(s string, maxWidth int) string {
	if maxWidth < 3 || len(s) <= maxWidth {
		return s
	}
	return s[:maxWidth-3] + "..."
}

func formatTime(t time.Time, timezone Timezone) string {
	timeFormat := "Mon 2006-01-02"
	switch timezone {
	case Timezone_Local:
		return t.Local().Format(timeFormat)
	case Timezone_UTC:
		return t.UTC().Format(timeFormat)
	default:
		panic("unknown timezone: " + timezone)
	}
}

type row struct {
	Key   string
	Value string
	Skip  bool
}

type rowOpt func(*row)

func skipRowIf(skip bool) rowOpt {
	return func(r *row) {
		r.Skip = skip
	}
}

func newRow(key string, value string, opts ...rowOpt) row {

	r := row{
		Key:   key,
		Value: value,
		Skip:  false,
	}
	for _, opt := range opts {
		opt(&r)
	}
	return r
}

type section []row

type keyValueTable struct {
	sections        []section
	maxKeyWidth     int
	w               io.Writer
	desiredMaxWidth int
}

// newKeyValueTable creates a new table and tries to fit it into desiredMaxWidth
// desiredMaxWidth is ignored if it == 0, or if it is less than the minimum width possible
//
//	width = len(key) + len(truncated_value) + len(padding)
//
// If desiredMaxWidth < len(key) + len(truncated_value) + len(padding) , it is ignored.
func newKeyValueTable(w io.Writer, desiredMaxWidth int) *keyValueTable {
	return &keyValueTable{
		sections:        nil,
		maxKeyWidth:     0,
		w:               w,
		desiredMaxWidth: desiredMaxWidth,
	}
}

func (k *keyValueTable) Section(rows ...row) {
	sec := make(section, 0, len(rows))
	for _, e := range rows {
		if !e.Skip {
			if len(e.Key) > k.maxKeyWidth {
				k.maxKeyWidth = len(e.Key)
			}
			sec = append(sec, e)
		}
	}
	if len(sec) > 0 {
		k.sections = append(k.sections, sec)
	}
}

func (k *keyValueTable) Render() {
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.SetOutputMirror(k.w)

	// ╭─────────┬───────────╮
	// 12--key--345--value--67
	// ╰─────────┴───────────╯
	const tablePadding = 7

	truncationWidth := k.desiredMaxWidth - k.maxKeyWidth - tablePadding
	if truncationWidth < 0 || k.desiredMaxWidth == 0 {
		truncationWidth = 0
	}

	for _, sec := range k.sections {
		for _, r := range sec {
			t.AppendRow(table.Row{
				r.Key,
				truncate(r.Value, truncationWidth),
			})
		}
		t.AppendSeparator()
	}
	t.Render()
}
