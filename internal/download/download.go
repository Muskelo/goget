package download

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

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

// struct to get standart info from all downloads structurs
type DownloadInfo struct {
	URL  *url.URL
	Path string

	Size     int64
	Progress int64

	Status int
	Err    error
}

// Download interface
// download something
type Download interface {
	Run()
	Info() DownloadInfo
}

// HTTPDownload
// download using http
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

func (dl *HTTPDownload) setErr(err error) {
	dl.err = err
	dl.status = ErrStatus
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
		return
	}

	dl.status = FinishedStatus
}

// FTPDownload
// download using ftp
type FTPDownload struct {
	url  *url.URL
	path string

	size     int64
	progress int64

	status int
	err    error
}

func NewFTPDownload(url_ *url.URL, path_ string) *FTPDownload {
	// set default port if not present
	if url_.Port() == "" {
		url_.Host = fmt.Sprintf("%v:%v", url_.Host, 21)
	}
	return &FTPDownload{url: url_, path: path_}
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

func (dl *FTPDownload) setErr(err error) {
	dl.err = err
	dl.status = ErrStatus
}

// get response from ftp
func (dl *FTPDownload) ftpGet() (*ftp.Response, error) {
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

	res, err := dl.ftpGet()
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
		return
	}

	dl.status = FinishedStatus
}

// BadDownload
// use when creating download failed
type BadDownload struct {
	url  *url.URL
	path string

	size     int64
	progress int64

	status int
	err    error
}

func NewBadDownload(url_ *url.URL, path string, err error) *BadDownload {
	return &BadDownload{url: url_, path: path, err: err, status: ErrStatus}
}

func (dl *BadDownload) Info() DownloadInfo {
	return DownloadInfo{
		URL:      dl.url,
		Path:     dl.path,
		Size:     dl.size,
		Progress: dl.progress,
		Status:   dl.status,
		Err:      dl.err,
	}
}

func (dl *BadDownload) Run() {
}

// DownloadManager
// control downloads
type DownloadManager struct {
	Status    int
	Downloads []Download
	Add       chan Download

	total     int
	completed int
	closed    bool
}

func NewDownloadManager() *DownloadManager {
	return &DownloadManager{
		Add: make(chan Download),
	}
}

// stops execution until the manager complete downloads
func (manager *DownloadManager) Wait() {
	for {
		if manager.closed && manager.total == manager.completed {
			manager.Status = FinishedStatus
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// used after adding the last download
func (manager *DownloadManager) Close() {
	close(manager.Add)
	manager.closed = true
}

func (manager *DownloadManager) checkFinished() {
	for {
		// if there are no new downloads
		// and this is the last download
		if manager.closed && manager.total == manager.completed {
			manager.Status = FinishedStatus
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// add downloads to manager from manager.Add chan
func (manager *DownloadManager) addDownloads() {
	for download := range manager.Add {
		go func(download Download) {
			manager.Downloads = append(manager.Downloads, download)
			manager.total++
			download.Run()
			manager.completed++

		}(download)
	}
}
func (manager *DownloadManager) Run() {
	manager.Status = InProgresStatus
	go manager.addDownloads()
	go manager.checkFinished()
}
