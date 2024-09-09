package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/bingoohuang/ngg/cmd"
)

func main() {
	flag.Parse()
	args := flag.Args()

	var (
		err     error
		timeout time.Duration
	)

	var options []cmd.Option
	if env := os.Getenv("TIMEOUT"); env != "" {
		timeout, err = time.ParseDuration(env)
		if err != nil {
			log.Fatalf("parse $TIMEOUT=%s: %v", env, err)
		}
		options = append(options, cmd.WithTimeout(timeout))
	}

	if env := os.Getenv("CWD"); env != "" {
		options = append(options, cmd.WithWorkingDir(env))
	}

	if env := os.Getenv("LINES"); env == "1" {
		options = append(options, cmd.WithStdout(cmd.NewLineStream(func(line string) {
			log.Printf("line: %s", line)
		})))
	}

	shell := cmd.ShellQuoteMust(args...)

	if env := os.Getenv("NOSH"); env == "1" {
		shell = ""
		options = append(options, cmd.WithCmd(exec.Command(args[0], args[1:]...)))
	}

	if shell != "" {
		log.Printf("shell: %q", shell)
	}

	c := cmd.New(shell, options...)
	if err := c.Run(context.TODO()); err != nil {
		log.Fatalf("error: %v", err)
	}

	log.Printf("stdout: %s", c.Stdout())
	log.Printf("stderr: %s", c.Stderr())
	log.Printf("exitCode: %d", c.ExitCode())
}
