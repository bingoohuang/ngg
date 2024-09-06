package proxytarget

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
	"strconv"
	"strings"
	"sync"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	Register(root.Cmd)
}

func (f *Cmd) initFlags(p *pflag.FlagSet) {
	p.StringVarP(&f.listen, "listen", "l", "", "Listen address, e.g. :3001")
	p.StringVarP(&f.proxyAddr, "proxy", "p", "127.0.0.1:6001", "Address of the proxy server to connect to")
	p.StringVarP(&f.targetAddr, "target", "t", "", "Target address to send to the proxy server")
	p.StringVarP(&f.yamlConfigFile, "config", "c", "", "YAML config file")
}

type Cmd struct {
	*root.RootCmd

	listen         string
	proxyAddr      string
	targetAddr     string
	yamlConfigFile string
}

func Register(rootCmd *root.RootCmd) {
	c := &cobra.Command{
		Use:   "proxytarget",
		Short: "a proxy to frp with target specified",
	}

	fc := &Cmd{RootCmd: rootCmd}
	c.Run = func(cmd *cobra.Command, args []string) {
		if err := fc.run(args); err != nil {
			fmt.Println(err)
		}
	}
	fc.initFlags(c.Flags())
	rootCmd.AddCommand(c)
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

func (f *Cmd) run(args []string) error {
	if f.yamlConfigFile != "" {
		yamlFile, err := ss.ExpandFilename(f.yamlConfigFile)
		if err != nil {
			return fmt.Errorf("expand yaml config file: %w", err)
		}

		yamlFileData, err := os.ReadFile(yamlFile)
		if err != nil {
			return fmt.Errorf("read yaml config file: %w", err)
		}

		var config Config
		if err = yaml.Unmarshal(yamlFileData, &config); err != nil {
			return fmt.Errorf("unmarshal yaml config file: %w", err)
		}

		var wg sync.WaitGroup

		for _, p := range config.Proxies {
			if p.ProxyAddr == "" {
				p.ProxyAddr = config.ProxyAddr
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := p.Serve(); err != nil {
					log.Printf("Error serving proxy: %v", err)
				}
			}()
		}

		wg.Wait()

		return nil
	}

	t := &TargetConfig{
		Listen:     f.listen,
		ProxyAddr:  f.proxyAddr,
		TargetAddr: f.targetAddr,
	}
	if err := t.Serve(); err != nil {
		log.Printf("Error serving proxy: %v", err)
		return err
	}

	return nil
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

	log.Printf("Listening on port: %d, desc: %s", listener.Addr().(*net.TCPAddr).Port, t.Desc)

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
		if targetURL.Scheme == "http" {
			targetHost += ":80"
		} else {
			targetHost += ":443"
		}
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

	log.Printf("Connected to proxy server at %s", proxy)

	// 构造要发送到代理服务器的目标地址字符串
	targetMessage := "TARGET " + target + ";"

	// 发送目标地址到代理服务器
	_, err = proxyConn.Write([]byte(targetMessage))
	if err != nil {
		log.Printf("Failed to send target address to proxy: %v", err)
		return
	}

	log.Printf("Sent target address to proxy: %s", target)

	// 将客户端连接和代理服务器连接衔接起来
	go io.Copy(proxyConn, conn)
	io.Copy(conn, proxyConn)
}
