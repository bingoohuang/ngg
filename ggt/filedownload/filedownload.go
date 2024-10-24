package filedownload

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/gnet"
	"github.com/spf13/cobra"
)

func init() {
	fc := &subCmd{}
	c := &cobra.Command{
		Use:     "filedownload",
		Aliases: []string{"dl"},
		Long:    "文件下载",
		RunE:    fc.run,
	}
	root.AddCommand(c, fc)
}

type subCmd struct {
	Files []string `short:"f" help:"files path"`
	Port  int      `short:"p" help:"listen port"`
}

func (f *subCmd) run(cmd *cobra.Command, args []string) error {
	if f.Port == 0 {
		var err error

		f.Port, err = getRandomPort()
		if err != nil {
			return err
		}
	}

	addr := fmt.Sprintf(":%d", f.Port)
	ip, _ := gnet.MainIPv4()

	for i, file := range f.Files {
		// 检查文件是否存在
		if _, err := os.Stat(file); err != nil {
			return fmt.Errorf("stat file %q: %w", file, err)
		}

		var routerPath string
		if len(f.Files) == 1 {
			routerPath = "/"
		} else {
			routerPath = fmt.Sprintf("/%d", i+1)
		}
		// 创建HTTP服务器的handler
		http.HandleFunc(routerPath, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, file)
		})

		fmt.Printf("Serving %s on http://%s%s%s", file, ip, addr, routerPath)
	}

	// 启动服务器
	return http.ListenAndServe(addr, nil)
}

// getRandomPort 获取一个可用的随机端口
func getRandomPort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close() // 确保在函数返回后关闭监听器
	return l.Addr().(*net.TCPAddr).Port, nil
}
