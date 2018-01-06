package table

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	runewidth "github.com/mattn/go-runewidth"
)

var red = color.New(color.FgRed).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()

type Data struct {
	data []interface{}
}
type Table struct {
	Rows     []Row
	Header   Data
	Renderer map[int]Renderer
}

type Renderer func(string, interface{}) string

func Default(format string, data interface{}) string {
	return fmt.Sprintf(format, data)
}

func RedGreen(format string, data interface{}) string {
	switch t := data.(type) {
	case float64:
		if t < 0 {
			return red(fmt.Sprintf(format, data))
		} else {
			return green(fmt.Sprintf(format, data))
		}
	case int:
		if t < 0 {
			return red(fmt.Sprintf(format, data))
		} else {
			return green(fmt.Sprintf(format, data))
		}
	}
	return Default(format, data)
}
func NegativeRed(format string, data interface{}) string {
	switch t := data.(type) {
	case float64:
		if t < 0 {
			return red(fmt.Sprintf(format, data))
		}
	case int:
		if t < 0 {
			return red(fmt.Sprintf(format, data))
		}
	}
	return Default(format, data)
}

func PositiveRed(format string, data interface{}) string {
	switch t := data.(type) {
	case float64:
		if t > 0 {
			return red(fmt.Sprintf(format, data))
		}
	case int:
		if t > 0 {
			return red(fmt.Sprintf(format, data))
		}
	}
	return Default(format, data)

}

type Row interface {
	render(io.Writer, []int, map[int]Renderer)
}

func (d Data) render(writer io.Writer, constraints []int, renderer map[int]Renderer) {
	arr := make([]string, len(constraints))
	for k, c := range d.data {
		var format string
		width := constraints[k]
		switch c.(type) {
		case int:
			format = fmt.Sprintf(" %%%dd ", width)
		case float64:
			format = fmt.Sprintf(" %%%d.2f ", width)
		default:
			format = fmt.Sprintf(" %%%ds ", width)

		}
		r, ok := renderer[k]
		if !ok {
			arr[k] = Default(format, c)
		} else {
			arr[k] = r(format, c)
		}
	}
	render(writer, arr, "|")
}

type Separator struct {
}

func (s Separator) render(writer io.Writer, constraints []int, renderers map[int]Renderer) {
	arr := make([]string, len(constraints))
	for k, c := range constraints {
		s := "-"
		for i := 0; i < c; i++ {
			s += "-"
		}
		s += "-"
		arr[k] = s
	}
	render(writer, arr, "+")
}

func render(writer io.Writer, data []string, separator string) {
	s := separator + strings.Join(data, separator) + separator + "\n"
	writer.Write([]byte(s))
}

func NewTable() *Table {
	return &Table{
		Rows:     make([]Row, 0),
		Header:   Data{make([]interface{}, 0)},
		Renderer: map[int]Renderer{},
	}
}

func (t *Table) SetHeader(headers ...string) {
	h := make([]interface{}, len(headers))
	for k, v := range headers {
		h[k] = v
	}
	t.Header = Data{h}
}
func (t *Table) Append(data ...interface{}) {
	t.Rows = append(t.Rows, Data{data})
}
func (t *Table) AppendSeparator() {
	t.Rows = append(t.Rows, Separator{})
}

func (t *Table) Render(writer io.Writer) {
	s := Separator{}
	//Get the maximum length of each cell
	widths := t.computeCellWidth()
	s.render(writer, widths, t.Renderer)
	t.Header.render(writer, widths, t.Renderer)
	s.render(writer, widths, t.Renderer)
	t.AppendSeparator()
	previousIsSeparator := true
	for _, r := range t.Rows {
		_, isSeparator := r.(Separator)
		if !isSeparator {
			r.render(writer, widths, t.Renderer)
		} else if !previousIsSeparator {
			r.render(writer, widths, map[int]Renderer{})
		}
		previousIsSeparator = isSeparator
	}
	t.AppendSeparator()

}

func (t *Table) computeCellWidth() []int {
	max := len(t.Header.data)
	// compute the maximum number of column
	for _, v := range t.Rows {
		if r, ok := v.(Data); ok {
			if len(r.data) > max {
				max = len(r.data)
			}
		}
	}
	widths := make([]int, max)
	for k, v := range t.Header.data {
		widths[k] = runewidth.StringWidth(v.(string))

	}
	//get the max size for each cols
	for _, v := range t.Rows {
		if r, ok := v.(Data); ok {
			for j, s := range r.data {
				var l int
				switch t := s.(type) {
				case string:
					l = len(t)
				case int:
					l = len(fmt.Sprintf("%d", t))
				case float64:
					l = len(fmt.Sprintf("%.2f", t))
				default:
					l = 0
				}
				if l > widths[j] {
					widths[j] = l
				}
			}
		}
	}
	return widths
}
