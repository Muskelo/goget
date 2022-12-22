package printers

import (
	"fmt"
	"strings"
	"time"

	dl "ex.com/goget/goget/download"
	"ex.com/goget/goget/writers"
)

type Printer interface {
	Err(error)
    Msg(string)
	WatchDownloadManager(*dl.DownloadManager)
}

type StringPrinter struct {
	w writers.Writer
}
func NewStringPrinter(w writers.Writer) *StringPrinter {
    printer := &StringPrinter{w}
    return printer
}

func (printer *StringPrinter) Err(err error) {
	printer.w.Writef("Err: %v\n", err)
}
func (printer *StringPrinter) getDownloadManagerStatus(manager *dl.DownloadManager) string {
	status := dl.StatusAliases[manager.Status]
	lines := []string{status}
	for i, download := range manager.Downloads {
		lines = append(lines, fmt.Sprintf("#%v Download from %v to %v", i, download.URL, download.Path))
		switch download.Status {
		case dl.InProgresStatus:
			lines = append(lines, fmt.Sprintf("Downloading %v/%v bytes", download.Progress, download.Size))
		case dl.FinishedStatus:
			lines = append(lines, fmt.Sprintf("Downloaded %v bytes", download.Size))
		case dl.ErrStatus:
			lines = append(lines, fmt.Sprintf("Err: %v", download.Err))
		case dl.CreatedStatus:
			lines = append(lines, "Download created")
		}
	}
	return strings.Join(lines, "\n")
}

func (printer *StringPrinter) WatchDownloadManager(manager *dl.DownloadManager) {
	for {
		printer.w.Writeln(printer.getDownloadManagerStatus(manager))

		if manager.Status == dl.FinishedStatus || manager.Status == dl.ErrStatus {
			return
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func (printer *StringPrinter) Msg(help string) {
    printer.w.Write(help)
}
