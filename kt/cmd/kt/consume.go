package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"regexp"

	"github.com/bingoohuang/ngg/kt/pkg/kt"
	"github.com/bingoohuang/ngg/ss"
	"github.com/spf13/cobra"
)

type consumeCmd struct {
	kt.ConsumerConfig `squash:"1"`

	Grep         string `help:"grep message"`
	N            int64  `help:"Max number of messages to consume"`
	Web          bool   `help:"Start web server for HTTP requests and responses event"`
	Context      string `help:"Web server context path if web is enable"`
	Port         int    `help:"Web server port if web is enable"`
	KeyEncoder   string `default:"string" enum:"hex,base64,string"`
	ValueEncoder string `default:"string" enum:"hex,base64,string"`

	sseSender *kt.SSESender
	grepExpr  *regexp.Regexp
}

func (c *consumeCmd) Run(*cobra.Command, []string) (err error) {
	if err := c.CommonArgs.Validate(); err != nil {
		return err
	}

	if c.Grep != "" {
		if c.grepExpr, err = regexp.Compile(c.Grep); err != nil {
			return fmt.Errorf("regex %q is invalid: %v", c.Grep, err)
		}
	}

	c.parseWeb()

	ve := kt.ParseBytesEncoder(c.ValueEncoder)
	ke := kt.ParseBytesEncoder(c.KeyEncoder)
	c.MessageConsumer = kt.NewPrintMessageConsumer(ke, ve, c.sseSender, c.grepExpr, c.N)

	_, err = kt.StartConsume(c.ConsumerConfig)
	return err
}
func (c *consumeCmd) parseWeb() {
	if !c.Web {
		return
	}

	port := c.Port
	if port <= 0 {
		port = ss.Rand().Port(19092)
	}

	stream := kt.NewSSEStream()
	c.sseSender = &kt.SSESender{Stream: stream}
	contextPath := path.Join("/", c.Context)
	log.Printf("contextPath: %s", contextPath)

	http.Handle("/", http.HandlerFunc(kt.SSEWebHandler(contextPath, stream)))
	log.Printf("start to listen on %d", port)
	go func() {
		addr := fmt.Sprintf(":%d", port)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("listen and serve failed: %v", err)
		}
	}()

	go ss.OpenInBrowser(fmt.Sprintf("http://127.0.0.1:%d", port), contextPath)
}

func (c *consumeCmd) LongHelp() string {
	return `Offset can be specified as a comma-separated list of intervals:
  [[partition=Start:End],...]
The default is to consume from the oldest Offset on every partition for the given topic.
 - partition is the numeric identifier for a partition. You can use "all" to
   specify a default OffsetInterval for all partitions.
 - Start is the included Offset where consumption should Start.
 - End is the included Offset where consumption should End.
The following syntax is supported for each Offset:
  (oldest|newest|resume)?(+|-)?(\d+)?
 - "oldest" and "newest" refer to the oldest and newest offsets known for a given partition.
 - "resume" can be used in combination with -group.
 - Use "+" with a numeric value to skip the given number of messages since the oldest Offset. 
   For example, "1=+20" will skip 20 Offset value since the oldest Offset for partition 1.
 - Use "-" with a numeric value to refer to only the given number of messages before the newest Offset. 
   For example, "1=-10" will refer to the last 10 Offset values before the newest Offset for partition 1.
 - Relative offsets are based on numeric values and will not take skipped offsets (e.g. due to compaction) into account.
 - Given only a numeric value, it is interpreted as an absolute Offset value.
More examples:
 - 0=10:20       To consume messages from partition 0 between offsets 10 and 20 (inclusive).
 - all=2:10      To define an OffsetInterval for all partitions use -1 as the partition identifier:
 - all=1-5,2=5-7 Override the offsets for a single partition, in this case 2:
 - 0=4:,2=1:10,6 To consume from multiple partitions:
 - This would consume messages from three partitions: p=0 offset>=4,  p=2 1<=offsets<=10, p=6 all offsets.
 - all=newest    To Start at the latest Offset for each partition
 - newest        Shorter of above
 - newest-10     To consume the last 10 messages
 - -10           Omit "newest", same with above
 - oldest+10     To skip the first 15 messages starting with the oldest Offset
 - +10           Omit "oldest",, same with above
`
}
