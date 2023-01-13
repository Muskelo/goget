package download

import (
	"net/url"
	"path"
	"testing"

	"ex.com/goget/goget/utils"
)


func TestHTTPDownload(t *testing.T) {
	// It's very heavy test, enable only if need
	// return

	url_, _ := url.Parse("http://localhost/test1MB.db")
	path_ := path.Join(t.TempDir(), "test1MD.db")
	download := NewHTTPDownload(url_, path_)
	download.Run()

	eq, err := utils.CompareMD5Sum([]string{download.Info().Path, "../../test/data/test1MB.db"})
	if err != nil {
		t.Errorf("Can't compare downloaded file with etalon %v", err)
		return
	}
	if !eq {
		t.Errorf("Downloaded and etalon files have differnt md5 sum")
		return
	}
}

func TestFTPDownload(t *testing.T) {
	// It's very heavy test, enable only if need
	// return

	url_, _ := url.Parse("ftp://testuser:testpass@localhost/test1MB.db")
	path_ := path.Join(t.TempDir(), "test1MB.db")
	download := NewFTPDownload(url_, path_)
	download.Run()

	eq, err := utils.CompareMD5Sum([]string{download.Info().Path, "../../test/data/test1MB.db"})
	if err != nil {
		t.Errorf("Can't compare downloaded file with etalon: %v", err)
		return
	}
	if !eq {
		t.Errorf("Downloaded and etalon files have differnt md5 sum")
		return
	}
}
