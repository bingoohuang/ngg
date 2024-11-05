package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"image"
	"image/png"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/gnet"
	"github.com/bingoohuang/ngg/ss"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

func main() {
	c := root.CreateCmd(nil, "ggtdl", "文件http下载", &subCmd{})
	if err := c.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

type subCmd struct {
	Files []string `short:"f" help:"files path"`
	Port  int      `short:"p" help:"listen port"`
}

func (f *subCmd) Run(cmd *cobra.Command, args []string) error {
	if f.Port == 0 {
		var err error

		f.Port, err = getRandomPort()
		if err != nil {
			return err
		}
	}

	addr := fmt.Sprintf(":%d", f.Port)
	ip, _ := gnet.MainIPv4()

	var imgs []DownloadImage

	baseURL := fmt.Sprintf("http://%s%s", ip, addr)
	var dlFiles []string

	for _, file := range f.Files {
		// 检查文件是否存在
		stat, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("stat file %q: %w", file, err)
		}

		if stat.IsDir() {
			err := filepath.WalkDir(file, func(root string, info os.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() {
					if !strings.HasPrefix(info.Name(), ".") {
						dlFiles = append(dlFiles, root)
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			dlFiles = append(dlFiles, file)
		}
	}

	for i, file := range dlFiles {
		routerPath := fmt.Sprintf("/%d", i+1)

		// 创建HTTP服务器的handler
		http.HandleFunc(routerPath, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, file)
		})

		fileURL := fmt.Sprintf("%s%s", baseURL, routerPath)
		fmt.Printf("Serving %s on %s\n", file, fileURL)
		RenderString(fileURL, false)

		qrImg := RenderImage(fileURL)
		qrBase64, err := ConvertImageToBase64(qrImg)
		if err != nil {
			return err
		}

		imgs = append(imgs, DownloadImage{
			File:   file,
			Title:  fileURL,
			QRCode: qrBase64,
		})
	}

	// 创建HTTP服务器的handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.New("webpage").Parse(tpl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, imgs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	go ss.OpenInBrowser(baseURL)

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

// RenderString as a QR code
// Copy from https://github.com/claudiodangelis/qrcp
func RenderString(s string, inverseColor bool) {
	q, err := qrcode.New(s, qrcode.Medium)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(q.ToSmallString(inverseColor))
}

// RenderImage returns a QR code as an image.Image
func RenderImage(s string) image.Image {
	q, err := qrcode.New(s, qrcode.Medium)
	if err != nil {
		log.Fatal(err)
	}
	return q.Image(256)
}

type DownloadImage struct {
	File   string
	Title  string
	QRCode string // Base64 字符串
}

// ConvertImageToBase64 将 image.Image 对象转换为 Base64 字符串
func ConvertImageToBase64(img image.Image) (string, error) {
	// 创建一个字节缓冲区
	buf := new(bytes.Buffer)

	// 将图像写入缓冲区，使用 PNG 格式
	err := png.Encode(buf, img)
	if err != nil {
		return "", err
	}

	// 将字节缓冲区中的数据编码为 Base64
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

const tpl = `
<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>扫码下载</title>
    <style>
        body {
            display: flex;
            flex-wrap: wrap;
            justify-content: center;
            padding: 20px;
            background-color: #f4f4f4;
        }
        .card {
            display: flex;
            flex-direction: column;
            align-items: center;
            margin: 10px;
            padding: 20px;
            width: 150px;
            background: white;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
            border-radius: 8px;
        }
        img {
            margin-top: 10px;
			width: 100px; /* 设置二维码图片宽度 */
            height: 100px; /* 设置二维码图片高度 */
        }
    </style>
</head>
<body>
    {{range .}}
		<div class="card">
			<div>{{.File}}</div>
			<img src="data:image/png;base64,{{.QRCode}}" title="{{.Title}}">
		</div>
    {{end}}
</body>
</html>
`
