package ss

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Arg is the application's argument options.
type Arg struct {
	Duration time.Duration `short:"d"`
	MyFlag   myFlag        `flag:"my"`
	Out      []string
	Port     int      `short:"p" val:"1234"`
	Input    string   `short:"i" val:"" required:"true"`
	Version  bool     `val:"false" usage:"Show version"`
	Other    string   `flag:"-"`
	V        int      `short:"v" count:"true"`
	Size     FlagSize `short:"s" val:"10MiB"`
	Pmem     float32
}

type myFlag struct {
	Value string
}

func (i *myFlag) Type() string   { return "my" }
func (i *myFlag) String() string { return i.Value }

func (i *myFlag) Set(value string) error {
	i.Value = value
	return nil
}

func TestParse(t *testing.T) {
	arg := &Arg{}
	ParseArgs(arg, []string{"app", "-i", "5003", "--out", "a", "--out", "b", "--my", "mymy", "-d", "10s", "-vvv", "-s", "2KiB", "--pmem", "0.618"})
	assert.Equal(t, 10*time.Second, arg.Duration)
	assert.Equal(t, myFlag{Value: "mymy"}, arg.MyFlag)
	assert.Equal(t, []string{"a", "b"}, arg.Out)
	assert.Equal(t, 1234, arg.Port)
	assert.Equal(t, "5003", arg.Input)
	assert.Equal(t, 3, arg.V)
	assert.Equal(t, FlagSize(uint64(2*1024)), arg.Size)
	assert.Equal(t, float32(0.618), arg.Pmem)
	// ... use arg
}

// Usage is optional for customized show.
func (a Arg) Usage() string {
	return fmt.Sprintf(`
Usage of pcap (%s):
  -i string HTTP port to capture, or BPF, or pcap file
  -v        Show version
`, a.VersionInfo())
}

// VersionInfo is optional for customized version.
func (a Arg) VersionInfo() string { return "v0.0.2 2021-05-19 08:33:18" }
