package sqliter

import (
	"bytes"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

type syncWriter struct {
	wr bytes.Buffer
	m  sync.Mutex
}

func (sw *syncWriter) Write(data []byte) (n int, err error) {
	sw.m.Lock()
	n, err = sw.wr.Write(data)
	sw.m.Unlock()
	return
}

func (sw *syncWriter) String() string {
	sw.m.Lock()
	defer sw.m.Unlock()
	return sw.wr.String()
}

func newBufLogger(sw *syncWriter) cron.Logger {
	return cron.PrintfLogger(log.New(sw, "", log.LstdFlags))
}

// Many tests schedule a job for every second, and then wait at most a second
// for it to run.  This amount is just slightly larger than 1 second to
// compensate for a few milliseconds of runtime.
const OneSecond = 1*time.Second + 50*time.Millisecond

func TestFuncPanicRecovery(t *testing.T) {
	var buf syncWriter
	secondParser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	cron := cron.New(cron.WithParser(secondParser),
		cron.WithChain(cron.Recover(newBufLogger(&buf))))
	cron.Start()
	defer cron.Stop()
	// @midnight
	// @every 5m
	// 每秒: * * * * * ?
	// 每5分钟: 0 5 * * * *", every5min(time.Local)},
	cron.AddFunc("@every 1s", func() {
		panic("YOLO")
	})

	<-time.After(OneSecond)
	if !strings.Contains(buf.String(), "YOLO") {
		t.Error("expected a panic to be logged, got none")
	}
}
