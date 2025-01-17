package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/bingoohuang/ngg/daemon/autoload"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/gum"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/yaml"
	"github.com/fatedier/frp"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func main() {
	root.RunCmd(nil, "frp", "frp with proxy target", &subCmd{})
}

type subCmd struct {
	ServerPort  int    `short:"P" help:"frp server port"`
	FrpCnf      string `short:"c" help:"FRP yaml config file" default:"~/.frp.yaml"`
	ProxyConfig string `short:"p" help:"YAML config file for proxy target" default:"~/.proxytarget.yaml"`

	Version bool `short:"v" help:"show version" version:"1"`
}

type TargetConfig struct {
	Listen     string // TCP 监听 ip:port, e.g. :3001
	ProxyAddr  string // 代理地址, e.g. 127.0.0.1:6001
	TargetAddr string // 目标地址, e.g. 192.168.1.5:8090
	Desc       string // 描述, e.g. gitlab, jenkins, etc.
}

type Config struct {
	ProxyAddr string // 代理地址, e.g. 127.0.0.1:6001
	Proxies   []TargetConfig
}

func (f *subCmd) Run(*cobra.Command, []string) error {
	if f.FrpCnf == "" {
		f.FrpCnf = "~/.frp.yaml"
	}

	f.FrpCnf = ss.ExpandHome(f.FrpCnf)

	tempFile, err := chooseServerPort(f.FrpCnf, f.ServerPort)
	if err != nil {
		return err
	}
	tempFile = ss.Or(tempFile, f.FrpCnf)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		if tempFile != f.FrpCnf {
			os.Remove(tempFile)
		}
		os.Exit(1)
	}()

	proxyConfigFile, _ := ss.ExpandFilename(f.ProxyConfig)
	if ok, _ := ss.Exists(proxyConfigFile); !ok {
		return frp.Run(tempFile)
	}

	if f.ProxyConfig != "" {
		func() {
			yamlFileData, err := os.ReadFile(proxyConfigFile)
			if err != nil {
				log.Printf("E! read yaml config file: %v", err)
				return
			}

			var config Config
			if err := yaml.Unmarshal(yamlFileData, &config); err != nil {
				log.Printf("E! unmarshal yaml config file: %v", err)
				return
			}

			for _, p := range config.Proxies {
				if p.ProxyAddr == "" {
					p.ProxyAddr = config.ProxyAddr
				}

				go func() {
					if err := p.Serve(); err != nil {
						log.Printf("Error serving proxy: %v", err)
					}
				}()
			}
		}()
	}

	if err := frp.Run(tempFile); err != nil {
		log.Printf("E! frp error: %v", err)
	}

	return nil
}

func chooseServerPort(frpFile string, flagServerPort int) (string, error) {
	frpConf, err := os.ReadFile(frpFile)
	if err != nil {
		return "", err
	}

	var configValues map[string]any
	if err := yaml.Unmarshal(frpConf, &configValues); err != nil {
		return "", err
	}
	serverPort := configValues["serverPort"]
	if multiPorts, ok := serverPort.(string); ok {
		if flagServerPort > 0 {
			configValues["serverPort"] = flagServerPort
		} else {
			ports := ss.Split(multiPorts, ",")
			chosen, err := gum.Choose(ports, gum.ChooseLimit(1), gum.ChooseHeader("choose serverPort"),
				gum.ChooseTimeout(3*time.Second), gum.ChooseTimeoutValues([]string{ports[ss.Rand().Intn(len(ports))]}))
			if err != nil {
				return "", err
			}
			log.Printf("choose server port: %v", chosen)

			configValues["serverPort"], _ = ss.Parse[int](chosen[0])
		}

		newConfig, err := yaml.Marshal(configValues)
		if err != nil {
			return "", err
		}
		// Create a temporary file
		file, err := ss.WriteTempFile("", "*.yaml", newConfig, false)
		if err != nil {
			return "", err
		}
		return file, nil
	}

	return frpFile, nil
}

func (t *TargetConfig) Serve() error {
	if t.TargetAddr == "" {
		return fmt.Errorf("target parameter is required")
	}

	// -proxy :6001 时, 补全为 127.0.0.1:6001
	if strings.HasPrefix(t.ProxyAddr, ":") {
		t.ProxyAddr = "127.0.0.1" + t.ProxyAddr
	}

	listener, err := net.Listen("tcp", t.Listen)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", t.Listen, err)
	}
	defer listener.Close()

	listenPort := listener.Addr().(*net.TCPAddr).Port
	log.Printf("Listening on http://127.0.0.1:%d, desc: %s", listenPort, t.Desc)

	if strings.HasPrefix(t.TargetAddr, "http") {
		return serveHTTP(listener, t.ProxyAddr, t.TargetAddr)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accepting connection: %w", err)
		}

		go handleConnection(conn, t.ProxyAddr, t.TargetAddr)
	}
}

func serveHTTP(l net.Listener, proxy, target string) error {
	targetURL, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("parse URL target: %s error: %v", target, err)
	}

	targetHost := targetURL.Host
	if targetURL.Port() == "" {
		targetHost += lo.Switch[string, string](targetURL.Scheme).
			Case("http", ":80").
			Case("https", ":443").
			Default(":80")
	}

	// 忽略证书验证的 HTTPS 客户端
	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	dialContext := tr.DialContext
	tr.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		conn, err := dialContext(ctx, network, proxy)
		if err == nil {
			// 构造要发送到代理服务器的目标地址字符串
			targetMessage := "TARGET " + targetHost + ";"
			// 发送目标地址到代理服务器
			_, err = conn.Write([]byte(targetMessage))
		}
		if err != nil {
			log.Printf("E! dial error: %v", err)
		}

		return conn, err
	}

	h := httputil.NewSingleHostReverseProxy(targetURL)
	h.Transport = tr
	director := h.Director

	h.Director = func(r *http.Request) {
		baseURL := "http://" + r.Host
		director(r)
		r.Host = targetURL.Host
		r.Header.Del("Accept-Encoding")
		r.Header.Set("Origin-Base-Url", baseURL)
	}

	targetBytes := []byte(target)

	h.ModifyResponse = func(resp *http.Response) error {
		replacedUrl := resp.Request.Header.Get("Origin-Base-Url")

		location := resp.Header.Get("Location")
		if location != "" {
			location = strings.Replace(location, target, replacedUrl, 1)
			resp.Header.Set("Location", location)
		}

		contentType := resp.Header.Get("Content-Type")
		if resp.Request != nil && strings.Contains(contentType, "text/html") {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			resp.Body.Close()

			body = bytes.ReplaceAll(body, targetBytes, []byte(replacedUrl))
			resp.Header.Set("Content-Length", strconv.Itoa(len(body)))
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		return nil
	}

	server := &http.Server{Handler: h}
	return server.Serve(l)
}

func handleConnection(conn net.Conn, proxy, target string) {
	defer conn.Close()

	proxyConn, err := net.Dial("tcp", proxy)
	if err != nil {
		log.Printf("Error dialing proxy %s: %v", proxy, err)
		return
	}
	defer proxyConn.Close()

	// log.Printf("Connected to proxy server at %s", proxy)

	// 构造要发送到代理服务器的目标地址字符串
	targetMessage := "TARGET " + target + ";"

	// 发送目标地址到代理服务器
	if _, err = proxyConn.Write([]byte(targetMessage)); err != nil {
		log.Printf("Failed to send target address to proxy: %v", err)
		return
	}

	// log.Printf("Sent target address to proxy: %s", target)

	// 将客户端连接和代理服务器连接衔接起来
	go io.Copy(proxyConn, conn)
	io.Copy(conn, proxyConn)
}
