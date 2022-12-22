package download

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"

	"ex.com/goget/goget/utils"
)

const (
	ErrStatus       = -1
	CreatedStatus   = 0
	InProgresStatus = 1
	FinishedStatus  = 2
)

var StatusAliases map[int]string = map[int]string{
	-1: "Error",
	0:  "Created",
	1:  "InProgres",
	2:  "Finished",
}

type Download struct {
	URL  *url.URL
	Path string

	Size     int64
	Progress int64

	Status int
	Err    error
}

func NewDownload(rawUrl string, rawPath string) (*Download, error) {
	URL, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	if URL.Scheme != "http" && URL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported protocol '%v'", URL.Scheme)
	}

	var Path string
	if utils.DirExist(rawPath) {
		Path = path.Join(rawPath, path.Base(URL.Path))
	} else {
		Path = rawPath
		if !utils.ParentDirExist(Path) {
			return nil, fmt.Errorf("parent dir of %v not exist", Path)
		}
	}
	if utils.FileExist(Path) {
		return nil, fmt.Errorf("file %v exist", Path)
	}

	return &Download{URL: URL, Path: Path}, nil
}

func (dl *Download) setErr(err error) {
	dl.Err = err
	dl.Status = ErrStatus
}
func (dl *Download) stream(r io.Reader, w io.Writer) error {
	quit := false

	for !quit {
		buf := make([]byte, 16384)

		readed, err := r.Read(buf)
		if err == io.EOF {
			quit = true
		} else if err != nil {
			return err
		}

		writed, err := w.Write(buf[:readed])
		if err != nil {
			return err
		}

		dl.Progress += int64(writed)
	}
	return nil
}
func (dl *Download) Run() {
	dl.Status = InProgresStatus

	res, err := http.Get(dl.URL.String())
	if err != nil {
		dl.setErr(err)
		return
	}
	defer res.Body.Close()
	dl.Size = res.ContentLength

	file, err := os.Create(dl.Path)
	if err != nil {
		dl.setErr(err)
		return
	}
	defer file.Close()

	err = dl.stream(res.Body, file)
	if err != nil {
		dl.setErr(err)
	}

	dl.Status = FinishedStatus
}

type DownloadManager struct {
	Status    int
	Downloads []*Download
}

func NewDownloadManager(args []string) (*DownloadManager, error) {
	cmd := &DownloadManager{}
	err := cmd.addDownloads(args)
	return cmd, err
}
func (manger *DownloadManager) addDownloads(args []string) error {
	argsLen := len(args)
	switch {
	case argsLen == 0:
		return fmt.Errorf("Expected 1 or more args")

	case argsLen == 1:
		download, err := NewDownload(args[0], ".")
		if err != nil {
			return err
		}
		manger.Downloads = append(manger.Downloads, download)

	case argsLen%2 == 1:
		return fmt.Errorf("Multiple download require even args")

	default:
		for i := 0; i < argsLen; i += 2 {
			download, err := NewDownload(args[i], args[i+1])
			if err != nil {
				return err
			}
			manger.Downloads = append(manger.Downloads, download)
		}
	}
	return nil
}

func (manager *DownloadManager) Run() {
	var wg sync.WaitGroup

	for _, download := range manager.Downloads {
		wg.Add(1)
		go func(download *Download) {
			defer wg.Done()
			download.Run()
		}(download)
	}

	manager.Status = InProgresStatus
	wg.Wait()
	manager.Status = FinishedStatus
}
