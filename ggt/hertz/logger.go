package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/bingoohuang/ngg/rotatefile/stdlog"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type defaultLogger struct {
	Writer io.Writer
	level  hlog.Level
}

func (ll *defaultLogger) SetLevel(level hlog.Level) {
	ll.level = level
}

func (ll *defaultLogger) SetOutput(writer io.Writer) {
	ll.Writer = writer
}

func HlogLevelString(lv hlog.Level) []byte {
	switch lv {
	case hlog.LevelTrace:
		return []byte("TRACE")
	case hlog.LevelDebug:
		return []byte("DEBUG")
	case hlog.LevelInfo:
		return []byte("INFO")
	case hlog.LevelNotice:
		return []byte("WARN")
	case hlog.LevelWarn:
		return []byte("WARN")
	case hlog.LevelError:
		return []byte("ERROR")
	case hlog.LevelFatal:
		return []byte("FATAL")
	default:
		return []byte("?????")
	}
}

func (ll *defaultLogger) Logf(ctx context.Context, lv hlog.Level, format *string, v ...interface{}) {
	if ll.level > lv {
		return
	}

	var msg string
	if format != nil {
		if len(v) > 0 {
			msg = fmt.Sprintf(*format, v...)
		} else {
			msg = *format
		}
	} else {
		msg = fmt.Sprint(v...)
	}

	levelBytes := HlogLevelString(lv)
	buf := stdlog.GetBuffer()
	defer stdlog.PutBuffer(buf)
	stdlog.WriteLogLine(ll.Writer, 5, levelBytes, []byte(msg), buf)

	if lv == hlog.LevelFatal {
		os.Exit(1)
	}
}

func (ll *defaultLogger) Fatal(v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelFatal, nil, v...)
}

func (ll *defaultLogger) Error(v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelError, nil, v...)
}

func (ll *defaultLogger) Warn(v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelWarn, nil, v...)
}

func (ll *defaultLogger) Notice(v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelNotice, nil, v...)
}

func (ll *defaultLogger) Info(v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelInfo, nil, v...)
}

func (ll *defaultLogger) Debug(v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelDebug, nil, v...)
}

func (ll *defaultLogger) Trace(v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelTrace, nil, v...)
}

func (ll *defaultLogger) Fatalf(format string, v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelFatal, &format, v...)
}

func (ll *defaultLogger) Errorf(format string, v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelError, &format, v...)
}

func (ll *defaultLogger) Warnf(format string, v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelWarn, &format, v...)
}

func (ll *defaultLogger) Noticef(format string, v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelNotice, &format, v...)
}

func (ll *defaultLogger) Infof(format string, v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelInfo, &format, v...)
}

func (ll *defaultLogger) Debugf(format string, v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelDebug, &format, v...)
}

func (ll *defaultLogger) Tracef(format string, v ...interface{}) {
	ll.Logf(context.TODO(), hlog.LevelTrace, &format, v...)
}

func (ll *defaultLogger) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	ll.Logf(ctx, hlog.LevelFatal, &format, v...)
}

func (ll *defaultLogger) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	ll.Logf(ctx, hlog.LevelError, &format, v...)
}

func (ll *defaultLogger) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	ll.Logf(ctx, hlog.LevelWarn, &format, v...)
}

func (ll *defaultLogger) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	ll.Logf(ctx, hlog.LevelNotice, &format, v...)
}

func (ll *defaultLogger) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	ll.Logf(ctx, hlog.LevelInfo, &format, v...)
}

func (ll *defaultLogger) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	ll.Logf(ctx, hlog.LevelDebug, &format, v...)
}

func (ll *defaultLogger) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	ll.Logf(ctx, hlog.LevelTrace, &format, v...)
}
