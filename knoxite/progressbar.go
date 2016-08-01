package main

import (
	"fmt"
	"math"
	"strings"
	"syscall"
	"unsafe"

	"github.com/knoxite/knoxite"
)

const progressBarFormat = "[#>-]"

// ProgressBar is a helper for printing a progres bar
type ProgressBar struct {
	Text    string
	Total   int64
	Current int64
	Width   uint
}

type terminalInfo struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getTerminalInfo() terminalInfo {
	ws := &terminalInfo{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		panic(errno)
	}
	return *ws
}

// NewProgressBar returns a new progress bar
func NewProgressBar(text string, total, current int64, width uint) *ProgressBar {
	return &ProgressBar{
		Text:    text,
		Total:   total,
		Current: current,
		Width:   width,
	}
}

// Print writes the progress bar to stdout
func (p *ProgressBar) Print() {
	pct := float64(p.Current) / float64(p.Total)
	if p.Total == 0 {
		pct = 0
	}

	// Clear current line
	fmt.Print("\033[2K\r")

	sizes := fmt.Sprintf("%s / %s",
		knoxite.SizeToString(uint64(p.Current)),
		knoxite.SizeToString(uint64(p.Total)))

	pcts := fmt.Sprintf("%.2f%%", pct*100)
	for len(pcts) < 7 {
		pcts = " " + pcts
	}

	ti := getTerminalInfo()
	barWidth := uint(math.Min(float64(p.Width), float64(ti.Col)/3.0))

	size := int(barWidth) - len(pcts) - 4
	fill := int(math.Max(2, math.Floor((float64(size)*pct)+.5)))
	if size < 16 {
		barWidth = 0
	}

	text := p.Text
	maxTextWidth := int(ti.Col) - 3 - int(barWidth) - len(sizes)
	if len(p.Text) > maxTextWidth {
		if len(p.Text)-maxTextWidth+3 < len(p.Text) {
			text = "..." + p.Text[len(p.Text)-maxTextWidth+3:]
		} else {
			text = ""
		}
	}
	if maxTextWidth < 0 {
		maxTextWidth = 0
	}

	// Print text
	s := fmt.Sprintf("%s%s  %s ",
		text,
		strings.Repeat(" ", maxTextWidth-len(text)),
		sizes)
	fmt.Print(s)

	if barWidth > 0 {
		progChar := progressBarFormat[2]
		if p.Current == p.Total {
			progChar = progressBarFormat[1]
		}

		// Print progress bar
		fmt.Printf("%c%s%c%s%c %s",
			progressBarFormat[0],
			strings.Repeat(string(progressBarFormat[1]), fill-1),
			progChar,
			strings.Repeat(string(progressBarFormat[3]), size-fill),
			progressBarFormat[4],
			pcts)
	}
}
