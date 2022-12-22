package writers

import (
	"fmt"
	"time"

	"github.com/gosuri/uilive"
)

type Writer interface {
	Start()
	Stop()
	Write(a ...any)
	Writef(format string, a ...any)
	Writeln(a ...any)
	WriteErr(error)
}

// Write to console using uilive writer
type ConsoleWriter struct {
	uiliveWriter *uilive.Writer
}

func NewConsoleWriter() *ConsoleWriter {
	w := &ConsoleWriter{}
	w.uiliveWriter = uilive.New()
	w.uiliveWriter.RefreshInterval = time.Minute
	return w
}
func (w *ConsoleWriter) Start() {
	w.uiliveWriter.Start()
}
func (w *ConsoleWriter) Stop() {
	w.uiliveWriter.Stop()
}

func (w *ConsoleWriter) Write(a ...any) {
	_, err := fmt.Fprint(w.uiliveWriter, a...)
	if err != nil {
		panic(err)
	}
	w.uiliveWriter.Flush()
}
func (w *ConsoleWriter) Writef(format string, a ...any) {
	_, err := fmt.Fprintf(w.uiliveWriter, format, a...)
	if err != nil {
		panic(err)
	}
	w.uiliveWriter.Flush()
}
func (w *ConsoleWriter) Writeln(a ...any) {
	_, err := fmt.Fprintln(w.uiliveWriter, a...)
	if err != nil {
		panic(err)
	}
	w.uiliveWriter.Flush()
}
func (w *ConsoleWriter) WriteErr(err error) {
	_, printErr := fmt.Fprintf(w.uiliveWriter, "Err: %v\n", err)
	if printErr != nil {
		panic(err)
	}
	w.uiliveWriter.Flush()
}

// Don't write anything
type QuiteWriter struct {
}

func NewQuiteWriter() *QuiteWriter {
	w := &QuiteWriter{}
	return w
}
func (w *QuiteWriter) Start() {
}
func (w *QuiteWriter) Stop() {
}

func (w *QuiteWriter) Write(a ...any) {
}
func (w *QuiteWriter) Writef(format string, a ...any) {
}
func (w *QuiteWriter) Writeln(a ...any) {
}
func (w *QuiteWriter) WriteErr(err error) {
}
