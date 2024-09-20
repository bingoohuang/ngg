package gossh

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/atotto/clipboard"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/gossh"
	"github.com/bingoohuang/ngg/ss"
	"github.com/spf13/cobra"
)

type subCmd struct {
	Ssh  bool   `help:"create sshpass ssh for hosts"`
	Scp  bool   `help:"create sshpass scp for hosts"`
	Repl bool   `short:"r" help:"repl mode"`
	Tag  string `short:"t" help:"command prefix tag"`

	PbePwd    string `help:"pbe password"`
	Pbe       string `help:"PrintEncrypt by pbe, string or @file"`
	Ebp       string `help:"PrintDecrypt by pbe, string or @file"`
	PbeChange string `help:"file to be change with another pbes"`
	PbePwdNew string `help:"new pbe pwd"`

	Cnf string `help:"cnf file path" short:"c"`

	gossh.Config `squash:"true"`
}

func init() {
	fc := &subCmd{}
	c := &cobra.Command{
		Use:  "ssh",
		RunE: fc.run,
	}

	root.AddCommand(c, fc)
}

func (fc *subCmd) run(*cobra.Command, []string) error {
	if fc.dealPbePflag() {
		return nil
	}

	if fc.Cnf != "" {
		cnf, err := ss.ExpandFilename(fc.Cnf)
		if err != nil {
			return err
		}
		cnfData, err := ss.ReadFile(cnf)
		if err != nil {
			return err
		}
		meta, err := toml.NewDecoder(bytes.NewReader(cnfData)).Decode(&fc.Config)
		if err != nil {
			return fmt.Errorf("DecodeFile error %w", err)
		}
		undecodedKeys := meta.Undecoded()
		if len(undecodedKeys) > 0 {
			log.Printf("unknown keys: %+v", undecodedKeys)
		}

		if fc.Tag != "" {
			var result map[string]any
			if _, err := toml.NewDecoder(bytes.NewReader(cnfData)).Decode(&result); err != nil {
				return fmt.Errorf("DecodeFile error %w", err)
			}

			if tagCmds, ok := result[fc.Tag+"-cmds"]; ok {
				fc.Cmds = nil
				if tc, ok := tagCmds.([]any); ok {
					for _, cmd := range tc {
						if s, ok := cmd.(string); ok && s != "" {
							fc.Cmds = append(fc.Cmds, s)
						}
					}
				}
			}
		}
	}

	fc.Group = ss.Or(fc.Group, "default")

	if fc.PrintConfig {
		fmt.Printf("Config%s\n", ss.JSONPretty(fc.Config))
	}

	gs := fc.Parse()
	if fc.Ssh {
		gs.Hosts.PrintSSH()
		return nil
	}
	if fc.Scp {
		gs.Hosts.PrintSCP()
		return nil
	}

	if len(gs.Cmds) == 0 {
		fc.Repl = true
	}

	logsDir := ss.ExpandHome("~/.gossh/logs/")
	_ = os.MkdirAll(logsDir, os.ModePerm)

	if fc.Cnf != "" {
		fc.Cnf += "-"
	}

	logFn := filepath.Join(logsDir, fc.Cnf+time.Now().Format("20060102150304")+".log")
	logFile, err := os.Create(logFn)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log file %s, error:%v\n", logFn, err)
	} else {
		fmt.Fprintf(os.Stdout, "log file %s created\n", logFn)
		fmt.Fprintf(logFile, "started at %s\n", time.Now().UTC().Format("2006-01-02 15:03:04"))
	}

	start := time.Now()

	var stdout io.Writer = os.Stdout

	if logFile != nil {
		stdout = io.MultiWriter(os.Stdout, logFile)

		defer func() {
			fmt.Fprintf(logFile, "finished at %s\n", time.Now().UTC().Format("2006-01-02 15:03:04"))
			fmt.Fprintf(logFile, "cost %s\n", time.Since(start))
			fmt.Fprintf(os.Stdout, "log file %s recorded\n", logFn)

			logFile.Close()
		}()
	}

	eo := gossh.ExecOption{}
	switch gs.Config.ExecMode {
	case gossh.ExecModeCmdByCmd:
		gossh.ExecCmds(&gs, gossh.NewExecModeCmdByCmd(), stdout, eo, fc.Group)
	case gossh.ExecModeHostByHost:
		hosts := append([]*gossh.Host{gossh.LocalHost}, gs.Hosts...)
		for _, host := range hosts {
			gossh.ExecCmds(&gs, host, stdout, eo, fc.Group)
		}
	}

	if fc.Repl {
		gs.Config.GlobalRemote = true
		gossh.Repl(&gs, gs.Hosts, stdout, fc.Group)
	}

	gs.Close()
	return nil
}

// dealPbePflag deals the request by the pflags.
func (fc *subCmd) dealPbePflag() bool {
	pbes := fc.Pbe
	ebps := fc.Ebp
	pbechg := fc.PbeChange

	if len(pbes) == 0 && len(ebps) == 0 && pbechg == "" {
		return false
	}

	alreadyHasOutput := false

	gossh.DecryptPassphrase("")
	passStr := ss.GetPbePwd()

	if len(pbes) > 0 {
		ss.PbePrintEncrypt(passStr, pbes)
		if val, err := ss.PbeEncode(pbes); err == nil {
			if err := clipboard.WriteAll(val); err == nil {
				fmt.Printf("Copied to clipboard\n")
			}
		}
		alreadyHasOutput = true
	}

	if len(ebps) > 0 {
		if alreadyHasOutput {
			fmt.Println()
		}

		ss.PbePrintDecrypt(passStr, ebps)

		if val, err := ss.PbeDecode(ebps); err == nil {
			if err := clipboard.WriteAll(val); err == nil {
				fmt.Printf("Copied to clipboard\n")
			}
		}

		alreadyHasOutput = true
	}

	if pbechg != "" {
		if alreadyHasOutput {
			fmt.Println()
		}

		processPbeChgFile(pbechg, passStr, fc.PbePwdNew)
	}

	return true
}

func processPbeChgFile(filename, passStr, pbenew string) {
	filename = ss.ExpandHome(filename)

	file, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	text, err := ss.Pbe{Passphrase: passStr}.Change(string(file), pbenew)
	if err != nil {
		panic(err)
	}

	ft, _ := os.Stat(filename)

	if err := os.WriteFile(filename, []byte(text), ft.Mode()); err != nil {
		panic(err)
	}
}
