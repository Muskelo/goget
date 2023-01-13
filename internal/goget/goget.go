package goget

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	dl "ex.com/goget/internal/download"
	"ex.com/goget/internal/printers"
	"ex.com/goget/internal/utils"
	"ex.com/goget/internal/writers"
)

type CommonFlags struct {
	Help           bool
	LogDisabled bool
	LogFile     string
	LogFormat   string
}

var (
	Flags      CommonFlags
	ErrNoInput = errors.New("Please input command or arg")
)

const helpMsg string = `About goget: goget - util to download files
syntax:
    goget [flags] <From> <To>
or:
    pipe | goget [flags]

flags:
    --help                        | this message
    --log-disabled  [true|false]  | disable log
    --log-format    [string|json] | log format
    --log-file      <filepath>    | log file
`

func parseFlags(args []string) ([]string, error) {
	flagSet := flag.NewFlagSet("common", flag.PanicOnError)
	flagSet.BoolVar(&Flags.Help, "help", false, "Help")
	flagSet.BoolVar(&Flags.LogDisabled, "log-disabled", false, "Disable log")
	flagSet.StringVar(&Flags.LogFormat, "log-format", "string", "Log format")
	flagSet.StringVar(&Flags.LogFile, "log-file", "", "Log file")

	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}
	return flagSet.Args(), nil
}
func createWriter() writers.Writer {
	var w writers.Writer
	switch {
	case Flags.LogDisabled:
		w = writers.NewQuiteWriter()
	case Flags.LogFile != "":
		w = writers.NewFileWriter(Flags.LogFile)
	default:
		w = writers.NewConsoleWriter()
	}
	return w
}
func createPrinter(w writers.Writer) printers.Printer {
	var p printers.Printer
	switch {
	case Flags.LogFormat == "json":
		p = printers.NewJsonPrinter(w)
	default:
		p = printers.NewStringPrinter(w)
	}
	return p
}
func parseArgs(args []string) (string, string, error) {
	argsLen := len(args)
	switch argsLen {
	case 0:
		return "", "", fmt.Errorf("To few args to create download")
	case 1:
		return args[0], ".", nil
	case 2:
		return args[0], args[1], nil
	default:
		return "", "", fmt.Errorf("To many args to create download")
	}
}
func createDownload(args []string) dl.Download {
	rawUrl, rawPath, err := parseArgs(args)
	if err != nil {
		return dl.NewBadDownload(nil, "", err)
	}

	url_, err := url.Parse(rawUrl)
	if err != nil {
		return dl.NewBadDownload(nil, "", err)
	}

	var path_ string
	if utils.DirExist(rawPath) {
		path_ = path.Join(rawPath, path.Base(url_.Path))
	} else {
		path_ = rawPath
	}
	if utils.FileExist(path_) {
		return dl.NewBadDownload(url_, path_, fmt.Errorf("file %v exist", path_))
	}

	switch url_.Scheme {
	case "http", "https":
		return dl.NewHTTPDownload(url_, path_)
	case "ftp":
		return dl.NewFTPDownload(url_, path_)
	default:
		return dl.NewBadDownload(url_, path_, fmt.Errorf("unsupported protocol '%v'", url_.Scheme))
	}
}
func readPipe(manager *dl.DownloadManager) {
	defer manager.Close()
	if !utils.PipeExist() {
		return
	}
	r := bufio.NewReader(os.Stdin)
	for {
		b, _, err := r.ReadLine()
		str := string(b)
		if err == io.EOF {
			break
		} else if err != nil {
			download := dl.NewBadDownload(nil, "", fmt.Errorf("Error while read from pipe: %v", err))
			manager.Add <- download
			break
		}
		str = strings.TrimSpace(str)
		if str == "" {
			break
		}
		args := strings.Split(str, " ")
		manager.Add <- createDownload(args)
	}
}
func Run() {
	args, err := parseFlags(os.Args[1:])
	if err != nil || Flags.Help {
		print(helpMsg)
		return
	}

	writer := createWriter()
	writer.Start()
	defer writer.Stop()
	printer := createPrinter(writer)

	manager := dl.NewDownloadManager()
	manager.Run()
	// add download from command args
	if len(args) > 0 {
		manager.Add <- createDownload(args)
	}
	// add downloads from pipe
	go readPipe(manager)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		printer.WatchDownloadManager(manager)
		wg.Done()
	}()
	go func() {
		manager.Wait()
		wg.Done()
	}()
	wg.Wait()
}
