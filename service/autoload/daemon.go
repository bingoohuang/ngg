package autoload

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/bingoohuang/ngg/service"
	"github.com/bingoohuang/ngg/ss"
)

type daemon struct {
	productName string
	running     atomic.Bool
	proc        *os.Process
}

func (d *daemon) Start(s service.Service) error {
	log.Println("daemon start")
	return nil
}

func (d *daemon) Stop(s service.Service) error {
	log.Println("service stop")
	d.running.Store(false)
	if d.proc != nil {
		log.Println("stop worker")
		d.proc.Kill()
	}
	if service.Interactive() {
		log.Println("stop daemon")
		os.Exit(0)
	}
	return nil
}

func (d *daemon) run() {
	log.Println("daemon run start")
	defer log.Println("daemon run end")
	d.running.Store(true)
	binPath, _ := os.Executable()
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	log.Println(mydir)
	conf := &service.Config{
		Name:        d.productName,
		DisplayName: d.productName,
		Description: d.productName,
		Executable:  binPath,
	}

	s, _ := service.New(d, conf)
	go s.Run()
	args := os.Args

	envs := []string{}
	for _, env := range os.Environ() {
		if !ss.HasPrefix(env, "SRV=") {
			envs = append(envs, env)
		}
	}

	for {
		// start worker
		tmpDump := filepath.Join("log", "dump.log.tmp")
		dumpFile := filepath.Join("log", "dump.log")
		f, err := os.Create(filepath.Join(tmpDump))
		if err != nil {
			log.Printf("E! start worker error: %s", err)
			return
		}
		log.Println("start worker process, args:", args)
		execSpec := &os.ProcAttr{Env: envs, Files: []*os.File{os.Stdin, os.Stdout, f}}
		lastRebootTime := time.Now()
		p, err := os.StartProcess(binPath, args, execSpec)
		if err != nil {
			log.Printf("E! start worker error: %s", err)
			return
		}
		d.proc = p
		_, _ = p.Wait()
		_ = f.Close()
		time.Sleep(time.Second)
		if err := os.Rename(tmpDump, dumpFile); err != nil {
			log.Printf("E! rename dump error: %s", err)
		}
		if !d.running.Load() {
			return
		}
		if time.Since(lastRebootTime) < time.Second*10 {
			log.Println("E! worker stop, restart it after 10s")
			time.Sleep(time.Second * 10)
		}

	}
}

func (d *daemon) Control(ctrlComm string, exeAbsPath string, args []string) error {
	svcConfig := &service.Config{
		Name:        d.productName,
		DisplayName: d.productName,
		Description: d.productName,
		Executable:  exeAbsPath,
		Arguments:   args,
	}

	s, e := service.New(d, svcConfig)
	if e != nil {
		return e
	}
	e = service.Control(s, ctrlComm)
	if e != nil {
		return e
	}

	return nil
}
