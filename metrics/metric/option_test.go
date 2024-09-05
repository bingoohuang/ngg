package metric_test

import (
	"os"
	"testing"
	"time"

	"github.com/bingoohuang/ngg/metrics/metric"
	"github.com/bingoohuang/ngg/ss"
	"github.com/stretchr/testify/assert"
)

func TestEnvOption(t *testing.T) {
	println(ss.Pick1(os.Getwd()))

	var o metric.Option

	metric.EnvOption("../testdata/golden.env")(&o)
	assert.Equal(t, metric.Option{
		AppName:         "bingoohuangapp",
		MetricsInterval: 3 * time.Second,
		HBInterval:      20 * time.Second,
		ChanCap:         123,
		LogPath:         "/tmp/metricslog",
		MaxBackups:      7,
	}, o)
}
