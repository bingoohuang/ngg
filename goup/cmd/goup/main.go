package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bingoohuang/ngg/goup"
	"github.com/bingoohuang/ngg/goup/codec"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/ver"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
	"github.com/segmentio/ksuid"
)

type Arg struct {
	ChunkSize   ss.Size `short:"c" val:"10MiB"`
	LimitRate   ss.Size `short:"L" `
	Coroutines  int     `short:"t"`
	Port        int     `short:"p" val:"2110"`
	Version     bool    `short:"v"`
	Init        bool
	Code        ss.FlagStringBool `short:"P"`
	Cipher      string            `short:"C"`
	ServerUrl   string            `short:"u"`
	FilePath    string            `short:"f"`
	Rename      string            `short:"r"`
	BearerToken string            `short:"b"`
	Paths       []string          `flag:"path"`
}

// Usage is optional for customized show.
func (a Arg) Usage() string {
	return `
Usage of goup:
  -b     string Bearer token for client or server, auto for server to generate a random one
  -c     string Chunk size for client (default 10MB, 0 to disable chunks), upload limit size for server.
  -t     int    Threads (go-routines) for client
  -f     string Upload file path for client
  -p     int    Listening port for server
  -r     string Rename to another filename
  -u     string Server upload url for client to connect to
  -P     string Password for PAKE
  -L     string Limit rate /s, like 10K for limit 10K/s
  -C     string Cipher AES256: AES-256 GCM, C20P1305: ChaCha20 Poly1305
  -v     bool   Show version
  --path /short=/short.zip Short URLs
`
}

// VersionInfo is optional for customized version.
func (a Arg) VersionInfo() string { return ver.Version() }

func main() {
	c := &Arg{}
	ss.ParseFlag(c)
	c.processCode()
	log.Printf("Args: %s", ss.Json(c))

	if c.ServerUrl == "" {
		if c.BearerToken == "auto" {
			c.BearerToken = goup.BearerTokenGenerate()
			log.Printf("Bearer token %s generated", c.BearerToken)
		}

		if err := goup.InitServer(); err != nil {
			log.Fatalf("init goup server: %v", err)
		}
		http.HandleFunc("/", goup.Bearer(c.BearerToken, goup.ServerHandle(c.Code.String(), c.Cipher,
			uint64(c.ChunkSize), uint64(c.LimitRate), c.Paths)))
		log.Printf("Listening on %d", c.Port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil); err != nil {
			log.Printf("E! listen failed: %v", err)
		}
		return
	}

	g, err := goup.New(c.ServerUrl,
		goup.WithFullPath(c.FilePath),
		goup.WithRename(c.Rename),
		goup.WithBearer(c.BearerToken),
		goup.WithChunkSize(uint64(c.ChunkSize)),
		goup.WithProgress(newSchollzProgressbar()),
		goup.WithCoroutines(c.Coroutines),
		goup.WithCode(c.Code.String()),
		goup.WithCipher(c.Cipher),
	)
	if err != nil {
		log.Fatalf("new goup client: %v", err)
	}

	if err := g.Start(); err != nil {
		log.Fatalf("start goup client: %v", err)
	}

	g.Wait()
}

func (a *Arg) processCode() {
	if a.Code.Exists && a.Code.Val == "" {
		pwd, err := codec.ReadPassword("Password")
		if err != nil {
			log.Printf("E! read password failed: %v", err)
		}
		_ = a.Code.Set(string(pwd))
	} else if a.Code.Val == "" && a.ServerUrl == "" {
		_ = a.Code.Set(ksuid.New().String())
		log.Printf("password is generate: %s", a.Code.String())
	}
}

type schollzProgressbar struct {
	bar *progressbar.ProgressBar
}

func (s *schollzProgressbar) Start(value uint64) {
	s.bar = progressbar.NewOptions64(
		int64(value),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Printf("\n")
		}),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}

func (s *schollzProgressbar) Add(value uint64) {
	_ = s.bar.Add64(int64(value))
}

func (s schollzProgressbar) Finish() {
	s.bar.Finish()
}

func newSchollzProgressbar() *schollzProgressbar {
	return &schollzProgressbar{}
}
