package hertz

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/rotatefile/stdlog"
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

func init() {
	Register(root.Cmd)
}

func (f *Cmd) run(_ []string) error {
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

	if f.MaxBody.Val > 0 {
		opts = append(opts, server.WithMaxRequestBodySize(int(f.MaxBody.Val)))
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

	count := 0
	for i, path := range f.Path {
		if len(f.Body) > i {
			body := f.Body[i]
			if err := handle(h, path, f.Methods, body, f.UploadPath); err != nil {
				return err
			}
			count++
		}
	}

	if count == 0 {
		h.Any("/", func(ctx context.Context, c *app.RequestContext) {
			c.JSON(http.StatusOK, map[string]any{
				"RemoteAddr":      c.RemoteAddr().String(),
				"X-Real-IP":       string(c.GetHeader("X-Real-IP")),
				"X-Forwarded-For": string(c.GetHeader("X-Forwarded-For")),
			})
		})
	}

	h.Spin()
	return nil
}

func handle(h *server.Hertz, path string, methods []string, body, uploadPath string) error {
	if uploadPath != "" {
		if stat, err := os.Stat(uploadPath); err != nil {
			return err
		} else if !stat.IsDir() {
			return fmt.Errorf("%s is not a directory", uploadPath)
		}

		h.Handle(http.MethodPost, path, func(ctx context.Context, c *app.RequestContext) {
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
		})
		return nil
	}

	rspBody, _ := ss.ExpandAtFile(body)

	contentType := lo.If(jj.Valid(rspBody), "application/json; charset=utf-8").
		Else("text/plain; charset=utf-8")

	for _, method := range methods {
		prefix := method + " "
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimSpace(path[len(prefix):])
			h.Handle(method, path, func(ctx context.Context, c *app.RequestContext) {
				c.Data(consts.StatusOK, contentType, []byte(rspBody))
			})
			return nil
		}
	}

	h.Any(path, func(ctx context.Context, c *app.RequestContext) {
		c.Data(consts.StatusOK, contentType, []byte(rspBody))
	})
	return nil
}

type Cmd struct {
	*root.RootCmd
	Addr       string      `help:"listening address" default:":12123"`
	MaxBody    ss.FlagSize `help:"Max request body Size" default:"4M"`
	Gzip       bool        `help:"gzip"`
	Methods    []string    `help:"methods" default:"GET"`
	UploadPath string      `short:"u" help:"Upload path"`
	Auth       string      `help:"basic auth like user:pass"`
	Path       []string    `short:"p" help:"path" default:"/"`
	Body       []string    `short:"b" help:"body"`
	Procs      int         `short:"t" help:"maximum number of CPUs" default:"runtime.GOMAXPROCS(0)"`
}

func Register(rootCmd *root.RootCmd) {
	c := &cobra.Command{
		Use:   "hertz",
		Short: "h",
		Long:  "hertz 测试服务器",
	}

	fc := &Cmd{RootCmd: rootCmd}
	c.Run = func(cmd *cobra.Command, args []string) {
		if err := fc.run(args); err != nil {
			fmt.Println(err)
		}
	}
	root.InitFlags(fc, c.Flags())
	rootCmd.AddCommand(c)
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
