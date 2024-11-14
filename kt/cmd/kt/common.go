package main

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"
	"unicode/utf16"
)

func listenForInterrupt(q chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	sig := <-signals
	log.Printf("received signal %s\n", sig)
	close(q)
}

func failf(msg string, args ...any) {
	log.Printf(msg+"\n", args...)
	os.Exit(1)
}

func readStdinLines(max int, out chan string) {
	s := bufio.NewScanner(os.Stdin)
	s.Buffer(make([]byte, max), max)

	for s.Scan() {
		out <- s.Text()
	}

	if err := s.Err(); err != nil {
		log.Printf("scanning input err=%v\n", err)
	}
	close(out)
}

// hashCode imitates the behavior of the JDK's String#hashCode method.
// https://docs.oracle.com/javase/7/docs/api/java/lang/String.html#hashCode()
//
// As strings are encoded in utf16 on the JVM, this implementation checks wether
// s contains non-bmp runes and uses utf16 surrogate pairs for those.
func hashCode(s string) (hc int32) {
	for _, r := range s {
		r1, r2 := utf16.EncodeRune(r)
		if r1 == 0xfffd && r2 == 0xfffd {
			hc = hc*31 + r
		} else {
			hc = (hc*31+r1)*31 + r2
		}
	}
	return
}

func kafkaAbs(i int32) int32 {
	switch {
	case i == -2147483648: // Integer.MIN_VALUE
		return 0
	case i < 0:
		return i * -1
	default:
		return i
	}
}

var random = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

func randPartition(partitions int32) int32 {
	return random.Int31n(partitions)
}

func hashCodePartition(key string, partitions int32) int32 {
	if partitions <= 0 {
		return -1
	}

	return kafkaAbs(hashCode(key)) % partitions
}
