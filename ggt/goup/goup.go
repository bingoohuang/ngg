package goup

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bingoohuang/ngg/ggt/goup/codec"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/ss"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
	"github.com/segmentio/ksuid"
	"github.com/spf13/cobra"
)

func Run() {
	c := root.CreateCmd(nil, "goup", "go upload client and server", &subCmd{})
	if err := c.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

type subCmd struct {
	Chunk      ss.FlagSize       `short:"c" default:"10MiB" help:"Chunk size for client (0 to disable chunks), upload limit size for server."`
	Limit      ss.FlagSize       `short:"L" help:"Limit rate /s, like 10K for limit 10K/s"`
	Coroutines int               `short:"t" help:"Threads (go-routines) for client"`
	Port       int               `short:"p" default:"2110" help:"Listening port for server"`
	Code       ss.FlagStringBool `short:"P" help:"Password for PAKE"`
	Cipher     string            `short:"C" help:"Cipher AES256: AES-256 GCM, C20P1305: ChaCha20 Poly1305"`
	ServerUrl  string            `short:"u" help:"Server upload url for client to connect to"`
	FilePath   string            `short:"f" help:"Upload file path for client"`
	Rename     string            `short:"r" help:"Rename to another filename"`
	Bearer     string            `short:"b" help:"Bearer token for client or server, auto for server to generate a random one"`
	Paths      []string          `flag:"path" help:"Short URLs"`
}

func (c *subCmd) Run(cmd *cobra.Command, args []string) error {
	c.processCode()
	log.Printf("Args: %s", ss.Json(c))

	if c.ServerUrl == "" {
		if c.Bearer == "auto" {
			c.Bearer = BearerTokenGenerate()
			log.Printf("Bearer token %s generated", c.Bearer)
		}

		if err := InitServer(); err != nil {
			return err
		}
		http.HandleFunc("/", Bearer(c.Bearer, ServerHandle(c.Code.String(), c.Cipher,
			uint64(c.Chunk), uint64(c.Limit), c.Paths)))
		log.Printf("Listening on %d", c.Port)
		return http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil)
	}

	g, err := New(c.ServerUrl,
		WithFullPath(c.FilePath),
		WithRename(c.Rename),
		WithBearer(c.Bearer),
		WithChunkSize(uint64(c.Chunk)),
		WithProgress(newSchollzProgressbar()),
		WithCoroutines(c.Coroutines),
		WithCode(c.Code.String()),
		WithCipher(c.Cipher),
	)
	if err != nil {
		return err
	}

	if err := g.Start(); err != nil {
		return err
	}

	g.Wait()
	return nil
}

func (c *subCmd) processCode() {
	if c.Code.Exists && c.Code.Val == "" {
		pwd, err := codec.ReadPassword("Password")
		if err != nil {
			log.Printf("E! read password failed: %v", err)
		}
		_ = c.Code.Set(string(pwd))
	} else if c.Code.Val == "" && c.ServerUrl == "" {
		_ = c.Code.Set(ksuid.New().String())
		log.Printf("password is generate: %s", c.Code.String())
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
