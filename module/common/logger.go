package common

import (
	"encoding/hex"
	"github.com/fatih/color"
	"github.com/injoyai/logs"
)

const (
	LevelNone = 0
	//LevelTrace = 1
	LevelDebug = 2
	LevelWrite = 3
	LevelRead  = 4
	LevelInfo  = 5
	//LevelWarn  = 6
	LevelError = 7
	LevelAll   = 999
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

var DefaultLogger Logger = NewLogger()

func NewLogger() *logger {
	return &logger{
		err:    logs.NewEntity("错误").SetFormatter(logs.TimeFormatter).SetSelfLevel(logs.LevelError).SetColor(color.FgRed),
		info:   logs.NewEntity("信息").SetFormatter(logs.TimeFormatter).SetSelfLevel(logs.LevelInfo).SetColor(color.FgCyan),
		read:   logs.NewEntity("读取").SetFormatter(logs.TimeFormatter).SetSelfLevel(logs.LevelRead).SetColor(color.FgBlue),
		write:  logs.NewEntity("写入").SetFormatter(logs.TimeFormatter).SetSelfLevel(logs.LevelWrite).SetColor(color.FgBlue),
		encode: func(p []byte) string { return string(p) },
		debug:  true,
	}
}

type logger struct {
	write  *logs.Entity
	read   *logs.Entity
	info   *logs.Entity
	err    *logs.Entity
	encode func(p []byte) string
	debug  bool
}

func (this *logger) Debug(b ...bool) {
	this.debug = len(b) == 0 || b[0]
}

func (this *logger) SetLevel(level int) {
	logsLevel := logs.Level(level)
	this.write.SetLevel(logsLevel)
	this.read.SetLevel(logsLevel)
	this.info.SetLevel(logsLevel)
	this.err.SetLevel(logsLevel)
}

func (this *logger) WithUTF8() {
	this.SetEncode(func(p []byte) string { return string(p) })
}

func (this *logger) WithHEX() {
	this.SetEncode(func(p []byte) string { return hex.EncodeToString(p) })
}

func (this *logger) SetEncode(f func(p []byte) string) {
	this.encode = f
}

func (this *logger) Readln(prefix string, p []byte) {
	if !this.debug {
		return
	}
	s := this.encode(p)
	this.read.Printf("%s%s\n", prefix, s)
}

func (this *logger) Writeln(prefix string, p []byte) {
	if !this.debug {
		return
	}
	s := this.encode(p)
	this.write.Printf("%s%s\n", prefix, s)
}

func (this *logger) Infof(format string, v ...interface{}) {
	if !this.debug {
		return
	}
	this.info.Printf(format, v...)
}

func (this *logger) Errorf(format string, v ...interface{}) {
	if !this.debug {
		return
	}
	this.err.Printf(format, v...)
}
