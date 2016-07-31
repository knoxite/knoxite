package main

import (
	"fmt"
	"knoxite"
	"math"
	"strings"
	"syscall"
	"unsafe"
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

	// Print text
	s := fmt.Sprintf("%s%s %s ",
		p.Text,
		strings.Repeat(" ", int(ti.Col)-len(p.Text)-2-int(p.Width)-len(sizes)),
		sizes)
	fmt.Print(s)

	size := int(p.Width) - len(pcts) - 4
	fill := math.Max(2, math.Floor((float64(size)*pct)+.5))

	progChar := progressBarFormat[2]
	if p.Current == p.Total {
		progChar = progressBarFormat[1]
	}

	// Print progress bar
	fmt.Printf("%c%s%c%s%c %s",
		progressBarFormat[0],
		strings.Repeat(string(progressBarFormat[1]), int(fill)-1),
		progChar,
		strings.Repeat(string(progressBarFormat[3]), size-int(fill)),
		progressBarFormat[4],
		pcts)
}
