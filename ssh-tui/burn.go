package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

//go:embed burn.json
var burnRaw []byte

type burnDay struct {
	Date     string   `json:"date"`
	Tokens   int64    `json:"tokens"`
	Commits  int      `json:"commits"`
	TopRepos []string `json:"topRepos"`
	TopMsgs  []string `json:"topMsgs"`
}

type burnModel struct {
	Short  string `json:"short"`
	Model  string `json:"model"`
	Tokens int64  `json:"tokens"`
}

type burnFile struct {
	GeneratedAt string      `json:"generatedAt"`
	TotalTokens int64       `json:"totalTokens"`
	ByModelTop  []burnModel `json:"byModelTop"`
	Last30      []burnDay   `json:"last30"`
}

var burnData burnFile
var burnError string

func init() {
	if len(burnRaw) == 0 {
		burnError = "burn.json empty"
		return
	}
	if err := json.Unmarshal(burnRaw, &burnData); err != nil {
		burnError = err.Error()
	}
}

func fmtTokens(t int64) string {
	switch {
	case t >= 1_000_000_000:
		return fmt.Sprintf("%.2fB", float64(t)/1e9)
	case t >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(t)/1e6)
	case t >= 1_000:
		return fmt.Sprintf("%dk", t/1000)
	default:
		return fmt.Sprintf("%d", t)
	}
}

func burnSparkline(days []burnDay, width int) string {
	if len(days) == 0 || width < 4 {
		return ""
	}
	// Downsample or upsample to width
	vals := make([]int64, len(days))
	var maxTok int64 = 1
	for i, d := range days {
		vals[i] = d.Tokens
		if d.Tokens > maxTok {
			maxTok = d.Tokens
		}
	}
	// 8-level block chart: ▁▂▃▄▅▆▇█
	blocks := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var b strings.Builder
	for _, v := range vals {
		if maxTok == 0 {
			b.WriteRune(' ')
			continue
		}
		ratio := float64(v) / float64(maxTok)
		idx := int(math.Round(ratio * float64(len(blocks)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		b.WriteRune(blocks[idx])
	}
	// If days < width, pad; if more, we already truncated by using per-day
	s := b.String()
	if len([]rune(s)) > width {
		// simple downsample: pick evenly spaced
		runes := []rune(s)
		step := float64(len(runes)) / float64(width)
		var out strings.Builder
		for i := 0; i < width; i++ {
			out.WriteRune(runes[int(float64(i)*step)])
		}
		return out.String()
	}
	if len([]rune(s)) < width {
		s += strings.Repeat(" ", width-len([]rune(s)))
	}
	return s
}

func burnLineChart(days []burnDay, width, height int) string {
	if len(days) == 0 || width < 10 || height < 3 {
		return ""
	}
	vals := make([]int64, len(days))
	var maxV int64 = 1
	for i, d := range days {
		vals[i] = d.Tokens
		if d.Tokens > maxV {
			maxV = d.Tokens
		}
	}
	// Canvas of height rows, width cols
	grid := make([][]rune, height)
	for r := range grid {
		grid[r] = make([]rune, width)
		for c := range grid[r] {
			grid[r][c] = ' '
		}
	}
	// Map each day to x,y
	n := len(vals)
	for i, v := range vals {
		x := 0
		if n > 1 {
			x = int(math.Round(float64(i) / float64(n-1) * float64(width-1)))
		}
		ratio := float64(v) / float64(maxV)
		y := int(math.Round((1 - ratio) * float64(height-1)))
		if y < 0 {
			y = 0
		}
		if y >= height {
			y = height - 1
		}
		// Mark point
		if grid[y][x] == ' ' {
			grid[y][x] = '·'
		}
		// Connect to next with simple line interpolation
		if i+1 < n {
			nextX := int(math.Round(float64(i+1) / float64(n-1) * float64(width-1)))
			nextRatio := float64(vals[i+1]) / float64(maxV)
			nextY := int(math.Round((1 - nextRatio) * float64(height-1)))
			// Bresenham-like interpolation: use larger of dx, |dy|
			dx := nextX - x
			dy := nextY - y
			steps := dx
			if steps == 0 {
				steps = 1
			}
			ady := dy
			if ady < 0 {
				ady = -ady
			}
			if ady > dx {
				steps = ady
			}
			if steps < 1 {
				steps = 1
			}
			for s := 1; s < steps; s++ {
				ix := x + int(float64(dx)*float64(s)/float64(steps))
				iy := y + int(float64(dy)*float64(s)/float64(steps))
				if ix >= 0 && ix < width && iy >= 0 && iy < height {
					if grid[iy][ix] == ' ' {
						if dy == 0 {
							grid[iy][ix] = '─'
						} else if (dx > 0 && dy > 0) || (dx < 0 && dy < 0) {
							grid[iy][ix] = '╲'
						} else {
							grid[iy][ix] = '╱'
						}
					}
				}
			}
		}
	}
	// Mark release days with ●. On narrow terminals many days map to the
	// same x — don't overwrite an existing ●; nudge ±1/±2 if free.
	for i, d := range days {
		if d.Commits < 20 {
			continue
		}
		x := 0
		if n > 1 {
			x = int(math.Round(float64(i) / float64(n-1) * float64(width-1)))
		}
		ratio := float64(d.Tokens) / float64(maxV)
		y := int(math.Round((1 - ratio) * float64(height-1)))
		if y < 0 || y >= height {
			continue
		}
		for _, ox := range []int{0, 1, -1, 2, -2} {
			nx := x + ox
			if nx < 0 || nx >= width {
				continue
			}
			if grid[y][nx] == '●' {
				continue // try a neighboring column
			}
			grid[y][nx] = '●'
			break
		}
	}

	var out strings.Builder
	for r := 0; r < height; r++ {
		out.WriteString(string(grid[r]))
		if r+1 < height {
			out.WriteString("\n")
		}
	}
	return out.String()
}

func burnModelBars(models []burnModel, width int) string {
	if len(models) == 0 || width < 20 {
		return ""
	}
	var maxT int64 = 1
	for _, m := range models {
		if m.Tokens > maxT {
			maxT = m.Tokens
		}
	}
	var b strings.Builder
	for i, m := range models {
		barW := width - 28 // account for label + tokens
		if barW < 6 {
			barW = 6
		}
		filled := int(math.Round(float64(m.Tokens) / float64(maxT) * float64(barW)))
		if filled < 1 && m.Tokens > 0 {
			filled = 1
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barW-filled)
		b.WriteString(fmt.Sprintf("%02d %-*s %s %s", i+1, 14, m.Short, bar, fmtTokens(m.Tokens)))
		if i+1 < len(models) {
			b.WriteString("\n")
		}
	}
	return b.String()
}
