package printers

import (
	"encoding/json"
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
		info := download.Info()
		lines = append(lines, fmt.Sprintf("#%v Download from %v to %v", i, info.URL, info.Path))
		switch info.Status {
		case dl.InProgresStatus:
			lines = append(lines, fmt.Sprintf("Downloading %v/%v bytes", info.Progress, info.Size))
		case dl.FinishedStatus:
			lines = append(lines, fmt.Sprintf("Downloaded %v bytes", info.Size))
		case dl.ErrStatus:
			lines = append(lines, fmt.Sprintf("Err: %v", info.Err))
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

// Schemes to json printer
type ErrorScheme struct {
	Err string `json:"err"`
}
type DownloadScheme struct {
	URL  string `json:"url"`
	Path string `json:"path"`

	Status      int    `json:"status"`
	StatusAlias string `json:"status_alias"`

	Size     int64 `json:"size"`
	Progress int64 `json:"progress"`
}
type DownloadManagerScheme struct {
	Status      int    `json:"status"`
	StatusAlias string `json:"status_alias"`

	Downloads []DownloadScheme `json:"downloads"`
}

type JsonPrinter struct {
	w writers.Writer
}

func NewJsonPrinter(w writers.Writer) *JsonPrinter {
	printer := &JsonPrinter{w}
	return printer
}

func (printer *JsonPrinter) convertToJsonString(scheme any) string {
	b, err := json.Marshal(scheme)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (printer *JsonPrinter) getDownloadManagerStatus(manager *dl.DownloadManager) string {
	managerScheme := DownloadManagerScheme{
		Status:      manager.Status,
		StatusAlias: dl.StatusAliases[manager.Status],
	}
	for _, download := range manager.Downloads {
		info := download.Info()
		downloadScheme := DownloadScheme{
			URL:         info.URL.String(),
			Path:        info.Path,
			Status:      info.Status,
			StatusAlias: dl.StatusAliases[info.Status],
			Size:        info.Size,
			Progress:    info.Progress,
		}
		managerScheme.Downloads = append(managerScheme.Downloads, downloadScheme)
	}
	return printer.convertToJsonString(managerScheme)
}
func (printer *JsonPrinter) WatchDownloadManager(manager *dl.DownloadManager) {
	for {
		printer.w.Writeln(printer.getDownloadManagerStatus(manager))

		if manager.Status == dl.FinishedStatus || manager.Status == dl.ErrStatus {
			return
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func (printer *JsonPrinter) Err(err error) {
	scheme := ErrorScheme{err.Error()}
	printer.w.Writeln(printer.convertToJsonString(scheme))
}
func (printer *JsonPrinter) Msg(help string) {
	printer.w.Write(help)
}
