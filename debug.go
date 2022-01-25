package zinc

import (
	"io"
	"log"
	"os"
	"time"
)

const DEBUG = true

type logWriter struct {
	io.Writer
	timeFormat string
}

func (w logWriter) Write(b []byte) (n int, err error) {
	tyme := time.Now().Format(w.timeFormat)
	return w.Writer.Write(append([]byte(tyme), b...))
}

var zlog = log.New(&logWriter{os.Stdout, "2006-01-02 15:04:05 "}, "", 0)

func ZPrintf(format string, v ...interface{}) {
	if !DEBUG {
		return
	}
	zlog.Printf(format, v...)
}

func ZErrorf(format string, v ...interface{}) {
	if !DEBUG {
		return
	}
	zlog.SetPrefix("[ERROR] ")
	zlog.Printf(format, v...)
	zlog.SetPrefix("")
}
