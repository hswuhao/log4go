/*
整个log包，只有一个IO对象:
  如非必要，不用配置多个写线程，默认只有1个IO线程。
  如有多个handler（如需要写到 socket、多个文件)时,
  再配置多个IO线程。

*/
package log4go

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type DropLogCallbackFunc func(l *LogInstance, sum int64)

type iHandleIOWriteThread interface {
	AsyncWrite(h Handler, fmt Formatter, log *LogInstance)

	// 当调用AsyncWrite异步写的chan满了，会直接丢弃log；
	// 丢弃前，通过DropLogCallbackFunc 回调一次，告诉上层应用。
	SetDropCallback(f DropLogCallbackFunc)
	Close()
}

const (
	MAX_WAIT_TIME_ON_EXIT = time.Second * 10
)

type hdlrWriter struct {
	Handler Handler
	Fmt     Formatter
	Log     *LogInstance
}

type HandleIOWriteThread struct {
	name   string
	clsoed bool
	quit   chan bool

	handlerWriterChan   chan *hdlrWriter // 一个IO线程处理多个handler的写
	handlerWriterBuffer *sync.Pool

	writeBuffer *bytes.Buffer

	dropCnt  int64
	writeCnt int64

	wg                  sync.WaitGroup
	dropLogCallbackFunc DropLogCallbackFunc
}

const _8k = 8192
const _4k = 4096 //假定文件系统block size=4k

func NewHandleIOWriteThread(name string, chanLength int) *HandleIOWriteThread {
	self := new(HandleIOWriteThread)

	self.name = name
	self.quit = make(chan bool, 10)
	self.handlerWriterChan = make(chan *hdlrWriter, chanLength)

	self.handlerWriterBuffer = &sync.Pool{
		New: func() interface{} {
			return new(hdlrWriter)
		},
	}

	// use 8k buffer in memory, linux filesys block was 4k
	self.writeBuffer = bytes.NewBuffer(make([]byte, 1024*8))

	go self.run()
	return self
}

func (self *HandleIOWriteThread) SetDropCallback(f DropLogCallbackFunc) {
	self.dropLogCallbackFunc = f
}

func (self *HandleIOWriteThread) AsyncWrite(
	h Handler, fmt Formatter, log *LogInstance) {

	if h == nil || fmt == nil {
		return
	}

	hw := self.handlerWriterBuffer.Get().(*hdlrWriter)
	hw.Handler = h
	hw.Fmt = fmt
	hw.Log = log

	select {
	case self.handlerWriterChan <- hw:
		return
	default:
		// TODO: 当handlerWriterChan满了时，只能丢弃日志，
		//    问题在于怎么通知开发人员，丢日志了，
		//    初步想法可以通过普罗米修斯这类的数据收集，进行告警。
		//    丢日志原因有很多，可能硬盘介质写速度太慢，或满了。
		//    如果是网络发送，也会有慢的时候。
		atomic.AddInt64(&self.dropCnt, 1)
		if self.dropLogCallbackFunc != nil {
			self.dropLogCallbackFunc(log, self.dropCnt)
		}
		// TODO：通知开发人员。
	}
}

func (self *HandleIOWriteThread) doFormat(hw *hdlrWriter,
	buff *bytes.Buffer) {

	defer func() {
		LogInstenceBuffer.Put(hw.Log)
		self.handlerWriterBuffer.Put(hw)

		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "\n[%s] PKG[wps.cn/log] err: %v\n",
				self.name, err)
		}
	}()

	if hw.Fmt == nil {
		hw.Fmt = globalTxtLineFormatter
	}

	if _, e := hw.Fmt.Format(buff, hw.Log); e != nil {
		//TODO: format出错，怎么办？
		if hw.Fmt != globalTxtLineFormatter {
			// TxtLineFormatter 不会返回出错
			hw.Fmt.Format(buff, hw.Log)
		} else {
			os.Stderr.WriteString(e.Error())
			self.dropCnt += 1
			return
		}
	}

	self.writeCnt += 1
}

func (self *HandleIOWriteThread) doWrite(hw *hdlrWriter) {
	pBuff := self.writeBuffer //default 8k
	self.doFormat(hw, pBuff)
	if pBuff.Len() >= _4k {
		hw.Handler.Write(pBuff.Bytes())
		pBuff.Reset()
	}

	for len(self.handlerWriterChan) > 0 {
		hw = <-self.handlerWriterChan
		self.doFormat(hw, pBuff)
		if pBuff.Len() >= _4k {
			hw.Handler.Write(pBuff.Bytes())
			pBuff.Reset()
		}
	}

	if pBuff.Len() > 0 {
		hw.Handler.Write(pBuff.Bytes())
		pBuff.Reset()
	}
}

func (self *HandleIOWriteThread) run() {
	self.wg.Add(1)
	defer self.wg.Done()
	stop := false
	var hw *hdlrWriter
	var quitStartTime time.Time
	for {
		select {
		case hw = <-self.handlerWriterChan:
			self.doWrite(hw)
		case <-self.quit:
			stop = true
			quitStartTime = time.Now()
		}
		if stop {
			if len(self.handlerWriterChan) == 0 {
				return
			}
			if time.Since(quitStartTime) >= MAX_WAIT_TIME_ON_EXIT {
				remain := len(self.handlerWriterChan)
				self.dropCnt += int64(remain)

				if remain > 0 {
					fmt.Fprintf(os.Stdout,
						"%s, but remain logs[%v] do not flush yet.",
						"log package was Closed()", remain)

					if self.dropLogCallbackFunc != nil {
						hw = <-self.handlerWriterChan
						self.dropLogCallbackFunc(hw.Log, self.dropCnt)
					}
				}
				return
			}
		}
	}
}

func (self *HandleIOWriteThread) Close() {
	// TODO Lock
	if self.clsoed {
		return
	}
	self.clsoed = true

	select {
	case self.quit <- true:
		self.wg.Wait()
	default:
		return
	}
}

func (self *HandleIOWriteThread) Stat() (name string, writeSum int64, dropSum int64) {
	return self.name, self.writeCnt, self.dropCnt
}

// 测试用函数
func GlobalIOThreadStat() (name string, writeSum int64, dropSum int64) {
	return globalWriteThread.Stat()
}
