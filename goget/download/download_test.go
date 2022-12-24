package download

import (
	"net/url"
	"os"
	"path"
	"reflect"
	"testing"

	"ex.com/goget/goget/utils"
)

func TestNewDownload(t *testing.T) {
	tmpDir := t.TempDir()

	type args struct {
		rawUrl  string
		rawPath string
	}
	type test struct {
		name    string
		args    args
		want    Download
		wantErr bool
	}
	tests := []test{}

	{
		rawUrl := "http://speedtest.ftp.otenet.gr/files/test1Mb.db"
		rawPath := path.Join(tmpDir, "output.db")
		url_, _ := url.Parse(rawUrl)

		tests = append(tests, test{
			name: "create http download",
			args: args{
				rawUrl,
				rawPath,
			},
			want: &HTTPDownload{
				url:  url_,
				path: rawPath,
			},
			wantErr: false,
		})
	}
	{
		rawPath := path.Join(tmpDir, "existingFile.db")
		os.Create(rawPath)

		tests = append(tests, test{
			name: "Try create download that write to existing file",
			args: args{
				"http://speedtest.ftp.otenet.gr/files/test1Mb.db",
				rawPath,
			},
			wantErr: true,
			want:    nil,
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, err := NewDownload(test.args.rawUrl, test.args.rawPath)

			if (err != nil) != test.wantErr {
				t.Errorf("newFile() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !reflect.DeepEqual(file, test.want) {
				t.Errorf("newFile() = %v, want %v", file, test.want)
			}
		})
	}
}

func TestHTTPDownload(t *testing.T) {
	// It's very heavy test, enable only if need
	// return

	download, err := NewDownload("http://localhost/test1MB.db", t.TempDir())
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	download.Run()

	eq, err := utils.CompareMD5Sum([]string{download.Info().Path, "../../test/data/test1MB.db"})
	if !eq {
		t.Errorf("Downloaded and etalon files have differnt md5 sum")
		return
	}
}

func TestFTPDownload(t *testing.T) {
	// It's very heavy test, enable only if need
	// return

    download, err := NewDownload("ftp://testuser:testpass@localhost/test1MB.db", t.TempDir())
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	download.Run()

	eq, err := utils.CompareMD5Sum([]string{download.Info().Path, "../../test/data/test1MB.db"})
	if !eq {
		t.Errorf("Downloaded and etalon files have differnt md5 sum")
		return
	}
}
