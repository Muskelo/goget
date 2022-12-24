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
	"github.com/jlaffaye/ftp"
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

type DownloadInfo struct {
	URL  *url.URL
	Path string

	Size     int64
	Progress int64

	Status int
	Err    error
}

type Download interface {
	Run()
	Info() DownloadInfo
}

func NewDownload(rawUrl string, rawPath string) (Download, error) {
	url_, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	var path_ string
	if utils.DirExist(rawPath) {
		path_ = path.Join(rawPath, path.Base(url_.Path))
	} else {
		path_ = rawPath
	}
	if utils.FileExist(path_) {
		return nil, fmt.Errorf("file %v exist", path_)
	}

	switch url_.Scheme {
	case "http", "https":
		return NewHTTPDownload(url_, path_), nil
	case "ftp":
		return NewFTPDownload(url_, path_), nil
	default:
		return nil, fmt.Errorf("unsupported protocol '%v'", url_.Scheme)
	}
}

// HTTPDownload
// download from web
type HTTPDownload struct {
	url  *url.URL
	path string

	size     int64
	progress int64

	status int
	err    error
}

func NewHTTPDownload(url_ *url.URL, path_ string) *HTTPDownload {
	return &HTTPDownload{url: url_, path: path_}
}

func (dl *HTTPDownload) setErr(err error) {
	dl.err = err
	dl.status = ErrStatus
}

func (dl *HTTPDownload) Info() DownloadInfo {
	return DownloadInfo{
		URL:      dl.url,
		Path:     dl.path,
		Size:     dl.size,
		Progress: dl.progress,
		Status:   dl.status,
		Err:      dl.err,
	}
}

func (dl *HTTPDownload) stream(r io.Reader, w io.Writer) error {
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

		dl.progress += int64(writed)
	}
	return nil
}
func (dl *HTTPDownload) Run() {
	dl.status = InProgresStatus

	res, err := http.Get(dl.url.String())
	if err != nil {
		dl.setErr(err)
		return
	}
	defer res.Body.Close()
	dl.size = res.ContentLength

	file, err := os.Create(dl.path)
	if err != nil {
		dl.setErr(err)
		return
	}
	defer file.Close()

	err = dl.stream(res.Body, file)
	if err != nil {
		dl.setErr(err)
	}

	dl.status = FinishedStatus
}

// FTPDownload
// download from ftp server
type FTPDownload struct {
	url  *url.URL
	path string

	size     int64
	progress int64

	status int
	err    error
}

func NewFTPDownload(url_ *url.URL, path_ string) *FTPDownload {
	// set default port in not present
	if url_.Port() == "" {
		url_.Host = fmt.Sprintf("%v:%v", url_.Host, 21)
	}
	return &FTPDownload{url: url_, path: path_}
}

func (dl *FTPDownload) setErr(err error) {
	dl.err = err
	dl.status = ErrStatus
}

func (dl *FTPDownload) Info() DownloadInfo {
	return DownloadInfo{
		URL:      dl.url,
		Path:     dl.path,
		Size:     dl.size,
		Progress: dl.progress,
		Status:   dl.status,
		Err:      dl.err,
	}
}

func (dl *FTPDownload) getFromFTP() (*ftp.Response, error) {
	conn, err := ftp.Dial(dl.url.Host)
	if err != nil {
		return nil, err
	}

	user := dl.url.User.Username()
	pass, _ := dl.url.User.Password()
	if err := conn.Login(user, pass); err != nil {
		return nil, err
	}

	dl.size, err = conn.FileSize(dl.url.Path)
	if err != nil {
		return nil, err
	}

	return conn.Retr(dl.url.Path)
}
func (dl *FTPDownload) stream(r io.Reader, w io.Writer) error {
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

		dl.progress += int64(writed)
	}
	return nil
}
func (dl *FTPDownload) Run() {
	dl.status = InProgresStatus

	res, err := dl.getFromFTP()
	if err != nil {
		dl.setErr(err)
		return
	}
	defer res.Close()

	file, err := os.Create(dl.path)
	if err != nil {
		dl.setErr(err)
		return
	}
	defer file.Close()

	err = dl.stream(res, file)
	if err != nil {
		dl.setErr(err)
	}

	dl.status = FinishedStatus
}

// DownloadManager
// control downloads
type DownloadManager struct {
	Status    int
	Downloads []Download
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
		go func(download Download) {
			defer wg.Done()
			download.Run()
		}(download)
	}

	manager.Status = InProgresStatus
	wg.Wait()
	manager.Status = FinishedStatus
}
