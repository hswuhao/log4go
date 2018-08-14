package log4go

import (
	"io"
)

//Handler writes logs to somewhere
type Handler interface {
	Write(p []byte) (n int, err error)
	Close() error
	AsyncWrite(fmt Formatter, log *LogInstance)
	SetWriteIOThread(th iHandleIOWriteThread)
}

//StreamHandler writes logs to a specified io Writer, maybe stdout, stderr, etc...
type StreamHandler struct {
	w           io.Writer
	writeThread iHandleIOWriteThread
}

func NewStreamHandler(w io.Writer) (*StreamHandler, error) {
	h := new(StreamHandler)

	h.w = w

	return h, nil
}

func (h *StreamHandler) AsyncWrite(fmt Formatter, log *LogInstance) {
	if h.writeThread != nil {
		h.writeThread.AsyncWrite(h, fmt, log)
	} else {
		globalWriteThread.AsyncWrite(h, fmt, log)
	}
}

func (h *StreamHandler) SetWriteIOThread(th iHandleIOWriteThread) {
	h.writeThread = th
}

func (h *StreamHandler) Write(b []byte) (n int, err error) {
	return h.w.Write(b)
}

func (h *StreamHandler) Close() error {
	if h.writeThread != nil {
		h.writeThread.Close()
	}
	return nil
}

//NullHandler does nothing, it discards anything.
type NullHandler struct {
}

func NewNullHandler() (*NullHandler, error) {
	return new(NullHandler), nil
}

func (h *NullHandler) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (h *NullHandler) AsyncWrite(fmt Formatter, log *LogInstance) {
	return
}

func (h *NullHandler) SetWriteIOThread(th iHandleIOWriteThread) {
	return
}

func (h *NullHandler) Close() error {
	return nil
}
