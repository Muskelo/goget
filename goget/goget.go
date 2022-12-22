package goget

import (
	"errors"
	"flag"
	"sync"

	dl "ex.com/goget/goget/download"
	"ex.com/goget/goget/printers"
	"ex.com/goget/goget/writers"
)

type CommonFlags struct {
	Help           bool
	OutputDisabled bool
	OutputFormat   string
	OutputFile     string
}

var (
	Flags      CommonFlags
	ErrNoInput = errors.New("Please input command or arg")
)

const helpMsg string = `About goget: goget - util to download files
Download file:
    goget [URL]
Multiple download:
    goget [URL-1] [output-1] [URL-2] [output-2] ...
Help:
    goget help
`

func parseFlags(args []string) ([]string, error) {
	if len(args) < 1 {
		return []string{}, ErrNoInput
	}

	flagSet := flag.NewFlagSet("common", flag.PanicOnError)
	flagSet.BoolVar(&Flags.Help, "help", false, "Help")
	flagSet.BoolVar(&Flags.OutputDisabled, "output-disabled", false, "Disable output")
	flagSet.StringVar(&Flags.OutputFormat, "output-format", "string", "Output formating output")
	flagSet.StringVar(&Flags.OutputFile, "output-file", "", "Output file output")

	if err := flagSet.Parse(args); err != nil {
		return []string{}, err
	}
	return flagSet.Args(), nil
}
func getWriter() writers.Writer {
	var w writers.Writer
	switch {
	case Flags.OutputDisabled:
		w = writers.NewQuiteWriter()
	case Flags.OutputFile != "":
		w = writers.NewFileWriter(Flags.OutputFile)
	default:
		w = writers.NewConsoleWriter()
	}
	return w
}
func getPrinter(w writers.Writer) printers.Printer {
	var p printers.Printer
	switch {
	case Flags.OutputFormat == "json":
		p = printers.NewJsonPrinter(w)
	default:
		p = printers.NewStringPrinter(w)
	}
	return p
}
func Run(args []string) {
	restArgs, err := parseFlags(args)

	w := getWriter()
	w.Start()
	defer w.Stop()

	printer := getPrinter(w)

	if err == ErrNoInput || Flags.Help {
		printer.Msg(helpMsg)
		return
	} else if err != nil {
		printer.Err(err)
		return
	}

	manager, err := dl.NewDownloadManager(restArgs)
	if err != nil {
		printer.Err(err)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		printer.WatchDownloadManager(manager)
		wg.Done()
	}()
	go func() {
		manager.Run()
		wg.Done()
	}()
	wg.Wait()
}
