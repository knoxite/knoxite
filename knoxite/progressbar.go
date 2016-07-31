package main

import (
	"fmt"
	"math"
	"strings"
	"syscall"
	"unsafe"
)

const progressBarFormat = "[=>-]"

// ProgressBar is a helper for printing a progres bar
type ProgressBar struct {
	Text    string
	Total   int64
	Current int64
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
func NewProgressBar(text string, total, current int64) *ProgressBar {
	return &ProgressBar{
		Text:    text,
		Total:   total,
		Current: current,
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

	// Print text
	s := fmt.Sprintf("%s %.2f%% ", p.Text, pct*100)
	fmt.Print(s)

	ti := getTerminalInfo()
	size := int(ti.Col) - len(s) - 3
	fill := math.Max(2, math.Floor((float64(size)*pct)+.5))

	// Print progress bar
	fmt.Printf("%c%s%c%s%c", progressBarFormat[0],
		strings.Repeat(string(progressBarFormat[1]), int(fill)-1),
		progressBarFormat[2],
		strings.Repeat(string(progressBarFormat[3]), size-int(fill)),
		progressBarFormat[4])
}
