package kt

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
)

type PrintContext struct {
	Output     any
	Done       chan struct{}
	MessageNum int
	ValueSize  int
}

func ParseBrokers(argBrokers string) []string {
	brokers := strings.Split(argBrokers, ",")
	for i, b := range brokers {
		if !strings.Contains(b, ":") {
			brokers[i] = b + ":9092"
		}
	}

	return brokers
}

func ParseKafkaVersion(s string) (sarama.KafkaVersion, error) {
	if s == "" {
		return sarama.V2_0_0_0, nil
	}

	v, err := sarama.ParseKafkaVersion(strings.TrimPrefix(s, "v"))
	if err != nil {
		return v, fmt.Errorf("failed to parse kafka KafkaVersion %s, error %q", s, err)
	}

	return v, nil
}

func PrintOut(in <-chan PrintContext) {
	PrintOutStats(in)
}

func PrintOutStats(in <-chan PrintContext) {
	messageNum := 0
	valueSize := 0
	start := time.Now()
	defer func() {
		cost := time.Since(start)
		fmt.Printf("total messages %d, size %s, cost: %s, TPS: %f message/s %s/s\n",
			messageNum, ss.Bytes(uint64(valueSize)), cost,
			float64(messageNum)/cost.Seconds(),
			ss.Bytes(uint64(float64(valueSize)/cost.Seconds())))
	}()

	for ctx := range in {
		messageNum += ctx.MessageNum
		valueSize += ctx.ValueSize

		buf, err := json.Marshal(ctx.Output)
		if err != nil {
			log.Printf("E! marshal Output %#v: %v", ctx.Output, err)
		}

		fmt.Println(string(jj.Color(buf, nil, nil)))
		close(ctx.Done)
	}
}

func LogClose(name string, c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("failed to close %s err=%v", name, err)
	}
}

// FirstNotNil returns the first non-nil string.
func FirstNotNil[T any](a ...*T) T {
	for _, i := range a {
		if i != nil {
			return *i
		}
	}

	var zero T
	return zero
}

func CurrentUserName() string {
	usr, err := user.Current()
	if err != nil {
		log.Printf("Failed to read current user err %v", err)
		return "unknown"
	}

	return sanitizeUsername(usr.Username)
}

var invalidClientIDCharactersRegExp = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func sanitizeUsername(u string) string {
	// Windows user may have format "DOMAIN|MACHINE\username", remove domain/machine if present
	s := strings.Split(u, "\\")
	u = s[len(s)-1]
	// Windows account can contain spaces or other special characters not supported
	// in client ID. Keep the bare minimum and ditch the rest.
	return invalidClientIDCharactersRegExp.ReplaceAllString(u, "")
}
