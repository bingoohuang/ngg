package xlsx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bingoohuang/ngg/xlsx"
	"github.com/bingoohuang/ngg/xlsx/pkg/upload"
	"github.com/stretchr/testify/assert"
)

func TestUpload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		x, err := xlsx.New(xlsx.WithUpload(r, "file"))
		if err != nil {
			w.WriteHeader(500)
			return
		}

		if err := x.Download(w, "file.xlsx"); err != nil {
			w.WriteHeader(500)
		}
	}))

	defer ts.Close()

	buf, fn, err := upload.Upload(context.Background(),
		ts.URL, "testdata/template.xlsx", "file", nil)
	assert.Nil(t, err)

	_ = os.WriteFile("testdata/dl.xlsx", buf.Bytes(), 0600)

	assert.Equal(t, "file.xlsx", fn)
}
