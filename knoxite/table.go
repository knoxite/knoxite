package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const timeFormat = "2006-01-02 15:04:05"

// Table is a helper for printing data in sheet form
type Table struct {
	Headers   []string
	Widths    []int64
	Rows      [][]interface{}
	Summary   []interface{}
	EmptyText string
}

// NewTable returns a new table
func NewTable(headers []string, widths []int64, emptyText string) Table {
	return Table{
		Headers:   headers,
		Widths:    widths,
		Rows:      [][]interface{}{},
		EmptyText: emptyText,
	}
}

// Print writes the entire table to stdout
func (t Table) Print() error {
	totalWidth := int64(0)
	format := ""
	for _, w := range t.Widths {
		format += "%" + strconv.FormatInt(w, 10) + "s  "
		totalWidth += int64(math.Abs(float64(w))) + 2
	}

	// print header
	fmt.Printf(format+"\n", ifaceify(t.Headers)...)
	fmt.Println(strings.Repeat("-", int(totalWidth)))

	// print rows
	for _, row := range t.Rows {
		fmt.Printf(format+"\n", row...)
	}
	if len(t.Rows) == 0 {
		fmt.Println(t.EmptyText)
	} else if len(t.Summary) > 0 {
		t.PrintSummary()
	}

	return nil
}

// PrintSummary writes the table summary to stdout
func (t Table) PrintSummary() error {
	totalWidth := int64(0)
	format := ""
	for _, w := range t.Widths {
		format += "%" + strconv.FormatInt(w, 10) + "s  "
		totalWidth += int64(math.Abs(float64(w))) + 2
	}

	// print divider
	fmt.Println(strings.Repeat("-", int(totalWidth)))

	// print summary
	fmt.Printf(format+"\n", t.Summary...)

	return nil
}

func ifaceify(list []string) []interface{} {
	vals := make([]interface{}, len(list))
	for i, v := range list {
		vals[i] = v
	}
	return vals
}
