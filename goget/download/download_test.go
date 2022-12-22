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
		changed chan *Download
	}
	type test struct {
		name    string
		args    args
		want    *Download
		wantErr bool
	}
	tests := []test{}

	{
		rawUrl := "http://speedtest.ftp.otenet.gr/files/test1Mb.db"
		rawPath := path.Join(tmpDir, "file.db")
		url_, _ := url.Parse(rawUrl)

		tests = append(tests, test{
			name: "default usage",
			args: args{
				rawUrl,
				rawPath,
				nil,
			},
			want: &Download{
				URL:  url_,
				Path: rawPath,
			},
			wantErr: false,
		})
	}
	{
		rawUrl := "http://speedtest.ftp.otenet.gr/files/test1Mb.db"
		rawPath := tmpDir
		url_, _ := url.Parse(rawUrl)

		tests = append(tests, test{
			name: "default usage",
			args: args{
				rawUrl,
				rawPath,
				nil,
			},
			want: &Download{
				URL:  url_,
				Path: path.Join(rawPath, "test1Mb.db"),
			},
			wantErr: false,
		})
	}
	{
		tests = append(tests, test{
			name: "parse file in non-existent dir",
			args: args{
				"http://speedtest.ftp.otenet.gr/files/test1Mb.db",
				path.Join(tmpDir, "nonExistentDir/file.db"),
				nil,
			},
			wantErr: true,
		})
	}
	{
		rawPath := path.Join(tmpDir, "existingFile.db")
		os.Create(rawPath)

		tests = append(tests, test{
			name: "parse existing file in non-existent dir",
			args: args{
				"http://speedtest.ftp.otenet.gr/files/test1Mb.db",
				rawPath,
				nil,
			},
			wantErr: true,
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

func TestDownload_Start(t *testing.T) {
	// It's very heavy test, enable if need
	// return

	download, err := NewDownload("http://speedtest.ftp.otenet.gr/files/test1Mb.db", t.TempDir())
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	download.Run()

	eq, err := utils.CompareMD5Sum([]string{download.Path, "../../test/test1Mb.db"})
	if !eq {
		t.Errorf("Downloaded and etalon files have differnt md5 sum")
		return
	}
}
