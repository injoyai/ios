package common

import (
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/fatih/color"
)

const (
	LevelAll   = 0
	LevelDebug = 1
	LevelInfo  = 2
	LevelError = 3
	LevelNone  = 9
)

type Logger interface {
	Readln(prefix string, p []byte)
	Writeln(prefix string, p []byte)
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	SetEncode(f func(p []byte) string)
	WithUTF8()
	WithHEX()
	Enable(b ...bool)
	SetLevel(level int)
}

func NewLogger() *logger {
	return &logger{
		Logger: slog.Default(),
		encode: func(p []byte) string { return string(p) },
		enable: true,
	}
}

var _ Logger = (*logger)(nil)

type logger struct {
	*slog.Logger
	enable bool
	debug  bool
	level  int
	encode func(p []byte) string
}

func (l *logger) Readln(prefix string, p []byte) {
	if !l.enable {
		return
	}
	if l.level > LevelDebug {
		return
	}
	t := time.Now().Format(time.TimeOnly)
	s := l.encode(p)
	color.Blue(t+" [读取] %s%s\n", prefix, s)
}

func (l *logger) Writeln(prefix string, p []byte) {
	if !l.enable {
		return
	}
	if l.level > LevelDebug {
		return
	}
	t := time.Now().Format(time.TimeOnly)
	s := l.encode(p)
	color.Blue(t+" [写入] %s%s\n", prefix, s)
}

func (l *logger) Infof(format string, v ...interface{}) {
	if !l.enable {
		return
	}
	if l.level > LevelInfo {
		return
	}
	t := time.Now().Format(time.TimeOnly)
	color.Cyan(t+" [信息] "+format, v...)
}

func (l *logger) Errorf(format string, v ...interface{}) {
	if !l.enable {
		return
	}
	if l.level > LevelError {
		return
	}
	t := time.Now().Format(time.TimeOnly)
	color.Red(t+" [错误] "+format, v...)
}

func (l *logger) SetEncode(f func(p []byte) string) {
	l.encode = f
}

func (l *logger) WithUTF8() {
	l.SetEncode(func(p []byte) string { return string(p) })
}

func (l *logger) WithHEX() {
	l.SetEncode(func(p []byte) string { return hex.EncodeToString(p) })
}

func (l *logger) Enable(b ...bool) {
	l.enable = len(b) == 0 || b[0]
}

func (l *logger) SetLevel(level int) {
	l.level = level
}
