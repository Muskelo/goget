package writers

import (
	"fmt"
	"os"
	"time"

	"github.com/gosuri/uilive"
)

// Writer interface
// write in a certain output
type Writer interface {
	Start()
	Stop()
	Write(a ...any)
	Writef(format string, a ...any)
	Writeln(a ...any)
}

// ConsoleWriter
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

// QuiteWriter
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

// FileWriter
// Write to file
type FileWriter struct {
	FN   string
	file *os.File
}

func NewFileWriter(fn string) *FileWriter {
	return &FileWriter{FN: fn}
}
func (w *FileWriter) Start() {
	file, err := os.Create(w.FN)
	if err != nil {
		panic(err)
	}
	w.file = file
}
func (w *FileWriter) Stop() {
	w.file.Close()
}

func (w *FileWriter) cleanFile() {
	if err := w.file.Truncate(0); err != nil {
		panic(err)
	}
	if _, err := w.file.Seek(0, 0); err != nil {
		panic(err)
	}
}
func (w *FileWriter) Write(a ...any) {
	w.cleanFile()
	_, err := w.file.WriteString(fmt.Sprint(a...))
	if err != nil {
		panic(err)
	}
}
func (w *FileWriter) Writef(format string, a ...any) {
	w.cleanFile()
	_, err := w.file.WriteString(fmt.Sprintf(format, a...))
	if err != nil {
		panic(err)
	}
}
func (w *FileWriter) Writeln(a ...any) {
	w.cleanFile()
	_, err := w.file.WriteString(fmt.Sprintln(a...))
	if err != nil {
		panic(err)
	}
}
