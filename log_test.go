package log4go_test

import (
	"math/rand"
	"os"
	"runtime"
	"sync"
	"testing"

	"wps.cn/lib/go/log"
)

func TestStdStreamLog(t *testing.T) {
	h, _ := log.NewStreamHandler(os.Stdout)
	logger := log.NewDefaultLogger(h)

	logger.Info("----line-text Format ----> hello world")

	logger.WithField("testing_field", "valesssss").WithField("k1", "v1").
		WithField("k2", "v22").
		Info("Json-Format: hello world")

	//all json format after set formater
	logger.SetFormatter(&log.JSONFormatter{})

	logger.Info("----All alway Json Format ----> \nhello world")
	logger.Warn("----All alway Json Format ----> hello world")
	logger.Error("----All alway Json Format ----> hello world")

	logger.Close()
	log.Info("Std-Logger line-txt Format hello world")
	log.StdLogger().SetFormatter(&log.JSONFormatter{})
	log.Info("Std-Logger All alway Json Format hello world")
	log.Warn("Std-Logger All alway Json Format hello world")
	log.Error("Std-Logger All alway Json Format hello world")
	log.Info("Std-Logger All alway Json Format hello world")

	log.WithFields(log.Fields{
		"k1": 1,
		"k2": 2,
		"k3": 3,
	}).Info("json INFO log k4=%v", 44)
	log.StdLogger().SetFormatter(&log.TxtLineFormatter{})
	for i := 0; i < 100; i++ {
		if rand.Int31n(100)&0x01 == 0x01 {
			log.Warn("+++++++++++TxtLineFormatter{%v}------------!", i)
		} else {
			log.Info("TxtLineFormatter{%v}------------!", i)
		}
	}
}

func TestMutilLogger(t *testing.T) {
	wg := sync.WaitGroup{}
	test_stdout := func() {
		h, _ := log.NewStreamHandler(os.Stdout)
		logger := log.NewLogger(h, log.StdLogFlag)
		logger.WithField("k1", "vv").
			Info("json: +++++++++++ ABCDEFGHI++++++++++++++++++++++")
		wg.Done()
	}

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go test_stdout()
	}
	wg.Wait()
}

func TestRotatingFileLog(t *testing.T) {
	path := "./test_log"
	os.RemoveAll(path)

	os.Mkdir(path, 0777)
	fileName := path + "/test"

	h, err := log.NewRotatingFileHandler(fileName, 10, 2)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 10)

	h.Write(buf)

	h.Write(buf)

	if _, err := os.Stat(fileName + ".1"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(fileName + ".2"); err == nil {
		t.Fatal(err)
	}

	h.Write(buf)
	if _, err := os.Stat(fileName + ".2"); err != nil {
		t.Fatal(err)
	}

	h.Close()

	os.RemoveAll(path)

}

func TestChangeStdHandler(t *testing.T) {
	fh, err := log.NewFileHandler("/dev/null")
	if err != nil {
		t.Fatalf("err=%v", err.Error())
		panic(err.Error())
	}
	log.SetHandler(fh)
	log.Info("file handler.....")
}

func TestMain(m *testing.M) {

	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	n := m.Run()
	log.Close()
	println(".....TestMain exit..... CPU=", runtime.NumCPU())
	os.Exit(n)
}
