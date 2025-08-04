package tableprint

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTruncate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		s        string
		max      int
		expected string
	}{
		{
			name:     "hello 2",
			s:        "hello",
			max:      2,
			expected: "hello",
		},
		{
			name:     "hello 3",
			s:        "hello",
			max:      3,
			expected: "...",
		},
		{
			name:     "hello 4",
			s:        "hello",
			max:      4,
			expected: "h...",
		},
		{
			name:     "hello 5",
			s:        "hello",
			max:      5,
			expected: "hello",
		},
		{
			name:     "hello 5",
			s:        "hello",
			max:      6,
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := truncate(tt.s, tt.max)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestTableTruncation(t *testing.T) {
	t.Parallel()
	// ╭─────┬──────╮
	// │ key │ v... │
	// ╰─────┴──────╯
	// 12345678901234

	expected := `╭─────┬──────╮
│ key │ v... │
╰─────┴──────╯
`

	buf := new(bytes.Buffer)
	tw := newKeyValueTable(buf, 14)
	tw.Section(
		newRow("key", "value"),
	)
	tw.Render()
	actual := buf.String()

	require.Equal(t, expected, actual)
}

func TestTableTruncationNoTruncate(t *testing.T) {
	t.Parallel()
	// ╭─────┬───────╮
	// │ key │ value │
	// ╰─────┴───────╯
	// 123456789012345

	expected := `╭─────┬───────╮
│ key │ value │
╰─────┴───────╯
`

	buf := new(bytes.Buffer)
	tw := newKeyValueTable(buf, 15)
	tw.Section(
		newRow("key", "value"),
	)
	tw.Render()
	actual := buf.String()

	require.Equal(t, expected, actual)
}

func TestTableTruncationToNarrow(t *testing.T) {
	t.Parallel()
	// ╭─────┬───────╮
	// │ key │ value │
	// ╰─────┴───────╯
	// 123456789012345

	expected := `╭─────┬───────╮
│ key │ value │
╰─────┴───────╯
`

	buf := new(bytes.Buffer)
	tw := newKeyValueTable(buf, 5)
	tw.Section(
		newRow("key", "value"),
	)
	tw.Render()
	actual := buf.String()

	require.Equal(t, expected, actual)
}
