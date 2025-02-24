package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/gnet"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/rotatefile/stdlog"
	_ "github.com/bingoohuang/ngg/rotatefile/stdlog/autoload"
	"github.com/bingoohuang/ngg/ss"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/basic_auth"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/tracer"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/gzip"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func main() {
	c := root.CreateCmd(nil, "hertz", "hertz 测试服务器", &subCmd{})
	if err := c.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

type subCmd struct {
	Addr       string      `help:"listening address" default:":12123"`
	MaxBody    ss.FlagSize `help:"max request body Size" default:"4M"`
	Gzip       bool        `help:"gzip"`
	Methods    []string    `help:"methods" default:"ANY"`
	UploadPath string      `short:"u" help:"upload path"`
	Auth       string      `help:"basic auth like user:pass"`
	Path       []string    `short:"p" help:"path" default:"/"`
	Body       []string    `short:"b" help:"body string, or @/some/path/file"`
	Procs      int         `short:"t" help:"maximum number of CPUs" default:"runtime.GOMAXPROCS(0)"`
	Version    bool        `version:"1" short:"v" help:"show version and exit"`
}

func (f *subCmd) Run(cmd *cobra.Command, args []string) error {
	if f.Procs > 0 {
		runtime.GOMAXPROCS(f.Procs)
	}

	var trace tracer.Tracer = &Trace{}

	hlog.SetLogger(&defaultLogger{
		Writer: stdlog.RotateWriter,
		level:  hlog.LevelInfo,
	})

	opts := []config.Option{
		server.WithTracer(trace),
		server.WithHostPorts(f.Addr),
		server.WithStreamBody(true),
	}

	if f.MaxBody > 0 {
		opts = append(opts, server.WithMaxRequestBodySize(int(f.MaxBody)))
	}

	h := server.New(opts...)
	h.Use(recovery.Recovery())
	if f.Gzip {
		h.Use(gzip.Gzip(gzip.DefaultCompression))
	}
	if f.Auth != "" {
		user, pass := ss.Split2(f.Auth, ":")
		h.Use(basic_auth.BasicAuth(map[string]string{user: pass}))
	}

	if len(f.Path) == 0 {
		f.Path = []string{"/"}
	}

	for i, path := range f.Path {
		var err error

		if f.UploadPath != "" {
			err = serveUpload(h, path, f.Methods, f.UploadPath)
		} else {
			if len(f.Body) > i {
				body := f.Body[i]
				err = serve(h, path, f.Methods, body)
			} else {
				fn := func(ctx context.Context, c *app.RequestContext) {
					c.JSON(http.StatusOK, serverInfo{
						RemoteAddr:    c.RemoteAddr().String(),
						XRealIP:       string(c.GetHeader("X-Real-IP")),
						XForwardedFor: string(c.GetHeader("X-Forwarded-For")),
						ServerIP:      func() string { ip, _ := gnet.MainIPv4(); return ip }(),
						ServerName:    ss.Pick1(os.Hostname()),
					})
				}
				err = serverMethod(h, f.Methods, path, fn)
			}
		}

		if err != nil {
			return err
		}
	}

	h.Spin()
	return nil
}

type serverInfo struct {
	ServerIP      string `json:"serverIP,omitempty"`
	ServerName    string `json:"serverName,omitempty"`
	RemoteAddr    string `json:"remoteAddr,omitempty"`
	XRealIP       string `json:"xRealIP,omitempty"`
	XForwardedFor string `json:"xForwardedFor,omitempty"`
}

func serveUpload(h *server.Hertz, path string, methods []string, uploadPath string) error {
	if stat, err := os.Stat(uploadPath); err != nil {
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", uploadPath)
	}

	f := func(ctx context.Context, c *app.RequestContext) {
		form, err := c.MultipartForm()
		if err != nil {
			log.Printf("MultipartForm err: %v", err)
			return
		}

		for _, files := range form.File {
			for _, file := range files {
				log.Printf("recv upload file: %s, size: %s", file.Filename, ss.Bytes(uint64(file.Size)))
				if err := c.SaveUploadedFile(file, filepath.Join(uploadPath, file.Filename)); err != nil {
					log.Printf("SaveUploadedFile err: %v", err)
				}
			}
		}
	}

	return serverMethod(h, methods, path, f)
}

func serve(h *server.Hertz, path string, methods []string, body string) error {
	rspBody, err := ss.ExpandAtFile(body)
	if err != nil {
		return fmt.Errorf("read file %s: %w", body, err)
	}

	contentType := lo.If(jj.Valid(rspBody), "application/json; charset=utf-8").Else("text/plain; charset=utf-8")
	f := func(ctx context.Context, c *app.RequestContext) {
		c.Data(consts.StatusOK, contentType, []byte(rspBody))
	}

	return serverMethod(h, methods, path, f)
}

func serverMethod(h *server.Hertz, methods []string, path string, f app.HandlerFunc) error {
	for _, method := range []string{consts.MethodGet, consts.MethodPost, consts.MethodPut, consts.MethodDelete, consts.MethodPatch} {
		prefix := method + " "
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimSpace(path[len(prefix):])
			h.Handle(method, path, f)
			return nil
		}
	}

	if len(methods) == 0 || len(methods) == 1 && slices.Contains(methods, "ANY") {
		h.Any(path, f)
		return nil
	}

	for _, method := range methods {
		h.Handle(method, path, f)
	}

	return nil
}

type Trace struct{}

func (t Trace) Start(ctx context.Context, c *app.RequestContext) context.Context {
	return ctx
}

func (t Trace) Finish(ctx context.Context, c *app.RequestContext) {
	// info := c.GetTraceInfo().Stats()
	// sendSize := info.SendSize()
	// recvSize := info.RecvSize()
}
