package common

import (
	"encoding/hex"
	"log/slog"

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
	Debug(b ...bool)
	SetLevel(level int)
}

func NewLogger() *logger {
	return &logger{
		Logger: slog.Default(),
		encode: func(p []byte) string { return string(p) },
		debug:  true,
	}
}

var _ Logger = (*logger)(nil)

type logger struct {
	*slog.Logger
	debug  bool
	level  int
	encode func(p []byte) string
}

func (l *logger) Readln(prefix string, p []byte) {
	if !l.debug {
		return
	}
	if l.level > LevelDebug {
		return
	}
	s := l.encode(p)
	color.Blue("[读取] %s%s\n", prefix, s)
}

func (l *logger) Writeln(prefix string, p []byte) {
	if !l.debug {
		return
	}
	if l.level > LevelDebug {
		return
	}
	s := l.encode(p)
	color.Blue("[写入] %s%s\n", prefix, s)
}

func (l *logger) Infof(format string, v ...interface{}) {
	if !l.debug {
		return
	}
	if l.level > LevelInfo {
		return
	}
	color.Cyan("[信息] "+format, v...)
}

func (l *logger) Errorf(format string, v ...interface{}) {
	if !l.debug {
		return
	}
	if l.level > LevelError {
		return
	}
	color.Red("[错误] "+format, v...)
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

func (l *logger) Debug(b ...bool) {
	l.debug = len(b) == 0 || b[0]
}

func (l *logger) SetLevel(level int) {
	l.level = level
}
