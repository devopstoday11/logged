package logged

import (
	"io"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	Info  = "info"
	Debug = "debug"
)

type Data map[string]string

type Log interface {
	Info(message string, data Data) error
	Debug(message string, data Data) error
	IsDebug() bool
}

type Config struct {
	Writer        io.Writer
	DebugPackages []string
	Defaults      Data
}

func New(c *Config) Log {
	return &log{
		serializer:    newJSONSerializer(c.Writer),
		debugPackages: c.DebugPackages,
		defaults:      c.Defaults,
	}
}

type log struct {
	mu            sync.Mutex
	serializer    *serializer
	defaults      Data
	debugPackages []string
}

func (l *log) Info(message string, data Data) error {
	return l.write(Info, message, data)
}

func (l *log) Debug(message string, data Data) error {
	if l.IsDebug() {
		return l.write(Debug, message, data)
	}
	return nil
}

func (l *log) IsDebug() bool {
	if len(l.debugPackages) == 0 {
		return false
	}

	pc, _, _, _ := runtime.Caller(2)
	funcName := runtime.FuncForPC(pc).Name()

	for _, pkg := range l.debugPackages {
		if strings.HasPrefix(funcName, pkg) {
			return true
		}
	}

	return false
}

func (l *log) write(level, message string, data Data) error {
	return l.serializer.write(&entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Message:   message,
		Data:      l.mergedData(data),
	})
}

func (l *log) mergedData(data Data) Data {
	if l.defaults == nil || len(l.defaults) == 0 {
		return data
	}

	if data == nil || len(data) == 0 {
		return l.defaults
	}

	merged := make(Data)
	for k, v := range l.defaults {
		merged[k] = v
	}
	for k, v := range data {
		merged[k] = v
	}

	return merged
}