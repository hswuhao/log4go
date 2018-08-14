package log4go

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type LogInstance struct {
	Flag  int
	Level string
	File  string
	Time  string
	Msg   string
	KV    Fields
}

type Formatter interface {
	Format(writeTobuff *bytes.Buffer, l *LogInstance) (*bytes.Buffer, error)
}

type JSONFormatter struct {
	// TODO: https://jsoniter.com/index.cn.html
}

const (
	keyFileNo = "file"
	keyTime   = "time"
	keyMsg    = "msg"
	keyLevel  = "level"
)

func (j *JSONFormatter) Format(writeTobuff *bytes.Buffer,
	l *LogInstance) (*bytes.Buffer, error) {

	kv := l.KV
	kv[keyFileNo] = l.File
	kv[keyTime] = l.Time
	kv[keyLevel] = l.Level
	kv[keyMsg] = l.Msg

	err := json.NewEncoder(writeTobuff).Encode(kv)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal to JSON, %v",
			err)
	}

	return writeTobuff, nil
}

type TxtLineFormatter struct {
}

func (_ *TxtLineFormatter) Format(writeTobuff *bytes.Buffer,
	l *LogInstance) (*bytes.Buffer, error) {

	if l.Flag&Ltime > 0 {
		writeTobuff.WriteString(l.Time)
		writeTobuff.WriteString(FieldSplit)
	}

	if l.Flag&Llevel > 0 {
		writeTobuff.WriteString(l.Level)
		writeTobuff.WriteString(FieldSplit)
	}

	if l.Flag&Lfile > 0 {
		writeTobuff.WriteString(l.File)
		writeTobuff.WriteString(FieldSplit)
	}

	writeTobuff.WriteString(l.Msg)
	if l.Msg[len(l.Msg)-1] != '\n' {
		writeTobuff.WriteByte('\n')
	}

	return writeTobuff, nil
}
