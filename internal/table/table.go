package table

import (
	"fmt"
	"os"
	"strings"
)

var noColor = os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb"

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
	gray   = "\033[90m"
)

func color(code, s string) string {
	if noColor {
		return s
	}
	return code + s + reset
}

func Bold(s string) string   { return color(bold, s) }
func Dim(s string) string    { return color(dim, s) }
func Red(s string) string    { return color(red, s) }
func Green(s string) string  { return color(green, s) }
func Yellow(s string) string { return color(yellow, s) }
func Blue(s string) string   { return color(blue, s) }
func Cyan(s string) string   { return color(cyan, s) }
func Gray(s string) string   { return color(gray, s) }

func StatusColor(status string) string {
	switch strings.ToLower(status) {
	case "up":
		return Green("● " + status)
	case "down":
		return Red("● " + status)
	case "degraded":
		return Yellow("● " + status)
	default:
		return Gray("○ " + status)
	}
}

func Bool(b bool) string {
	if b {
		return Green("✓")
	}
	return Gray("—")
}

func Print(headers []string, rows [][]string) {
	if len(rows) == 0 && len(headers) == 0 {
		return
	}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				// measure visible length (strip ANSI)
				vis := visibleLen(cell)
				if vis > widths[i] {
					widths[i] = vis
				}
			}
		}
	}

	// header
	parts := make([]string, len(headers))
	for i, h := range headers {
		parts[i] = Bold(pad(h, widths[i]))
	}
	fmt.Println(strings.Join(parts, "  "))

	// separator
	seps := make([]string, len(headers))
	for i, w := range widths {
		seps[i] = Gray(strings.Repeat("─", w))
	}
	fmt.Println(strings.Join(seps, "  "))

	// rows
	for _, row := range rows {
		cells := make([]string, len(headers))
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			cells[i] = padAnsi(cell, widths[i])
		}
		fmt.Println(strings.Join(cells, "  "))
	}
}

func pad(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

func padAnsi(s string, w int) string {
	vis := visibleLen(s)
	if vis >= w {
		return s
	}
	return s + strings.Repeat(" ", w-vis)
}

func visibleLen(s string) int {
	// strip ANSI escape sequences for width calculation
	inEsc := false
	n := 0
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		n++
	}
	return n
}
