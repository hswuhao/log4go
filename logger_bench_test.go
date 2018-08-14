//
//  go test -bench=. -benchmem -run=none
//   -benchtime=3s
/*

2018-08-06 11:11 跑的结果如下（win10 工作机）：
    admins@MICROSO-U4Q5SRI MINGW64 /i/GitHub/qing/src/wps.cn/lib/go/log (master)
        $ go test -bench=. -benchmem -run=none
        goos: windows
        goarch: amd64
        pkg: wps.cn/lib/go/log
        BenchmarkMutilLogger-4            300000              5409 ns/op            1216 B/op          9 allocs/op
        BenchmarkMutilJsonLogger-4        200000              9434 ns/op            1489 B/op         29 allocs/op
        BenchmarkJsonLogger-4             100000             19617 ns/op            6322 B/op         61 allocs/op
        BenchmarkTexLineLogger-4          300000              7072 ns/op             815 B/op          9 allocs/op
        PASS
        ok      wps.cn/lib/go/log       9.365s

    admins@MICROSO-U4Q5SRI MINGW64 /i/GitHub/qing/src/wps.cn/lib/go/log (master)
        $ go test -bench=. -benchmem -run=none
        goos: windows
        goarch: amd64
        pkg: wps.cn/lib/go/log
        BenchmarkMutilLogger-4            300000              5250 ns/op            1031 B/op          9 allocs/op
        BenchmarkMutilJsonLogger-4        200000             10993 ns/op            1539 B/op         29 allocs/op
        BenchmarkJsonLogger-4             100000             17709 ns/op            6322 B/op         61 allocs/op
        BenchmarkTexLineLogger-4          200000              5186 ns/op             558 B/op          9 allocs/op
        PASS
        ok      wps.cn/lib/go/log       7.960s

    admins@MICROSO-U4Q5SRI MINGW64 /i/GitHub/qing/src/wps.cn/lib/go/log (master)
        $ go test -bench=. -benchmem -run=none
        goos: windows
        goarch: amd64
        pkg: wps.cn/lib/go/log
        BenchmarkMutilLogger-4            200000              5431 ns/op            1065 B/op          9 allocs/op
        BenchmarkMutilJsonLogger-4        100000             11313 ns/op            1485 B/op         29 allocs/op
        BenchmarkJsonLogger-4             100000             18978 ns/op            6322 B/op         61 allocs/op
        BenchmarkTexLineLogger-4          300000              5639 ns/op             662 B/op          9 allocs/op
        PASS
        ok      wps.cn/lib/go/log       7.532s

---- 邪恶之分隔线 ------------------------------------------------------------------------------
2018-08-06 使用 Buffer = sync.Pool 后， 第3列的每次内存数申请少了。
    admins@MICROSO-U4Q5SRI MINGW64 /i/GitHub/qing/src/wps.cn/lib/go/log (master)
        $ go test -bench=. -benchmem -run=none
        goos: windows
        goarch: amd64
        pkg: wps.cn/lib/go/log
        BenchmarkMutilLogger-4            300000              5696 ns/op             369 B/op          9 allocs/op
        BenchmarkMutilJsonLogger-4        200000              9099 ns/op            1046 B/op         27 allocs/op
        BenchmarkJsonLogger-4             100000             19008 ns/op            4963 B/op         58 allocs/op
        BenchmarkTexLineLogger-4          300000              4936 ns/op             392 B/op          9 allocs/op
        PASS
        ok      wps.cn/lib/go/log       8.072s


---- 邪恶之分隔线 ------------------------------------------------------------------------------
2018-08-13 使用单一IO线程版本后：
$ go test -bench=. -benchmem -run=none
goos: windows
goarch: amd64
pkg: wps.cn/lib/go/log
BenchmarkMutilLogger-8            500000              3843 ns/op             379 B/op          9 allocs/op
BenchmarkMutilJsonLogger-8        300000              4357 ns/op             986 B/op         24 allocs/op
BenchmarkJsonLogger-8             200000             10093 ns/op            4162 B/op         50 allocs/op
BenchmarkTexLineLogger-8          300000              3627 ns/op             396 B/op          9 allocs/op
BenchmarkTestMutilLogger-8        200000              8614 ns/op            3243 B/op         52 allocs/op
PASS
[globalLogIOThread] IoThread.Close() writeCnt=[1625968] dropCnt=[532565] dropRate=[24.67%]
.....TestMain exit..... CPU= 4
ok      wps.cn/lib/go/log       9.371s
日志的处理能力： QPS=writeCnt=[1625968]/9.371s ~= 17万/s
*/
package log4go_test

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"

	log "github.com/kingsoft-wps/log4go"
)

func BenchmarkMutilLogger(b *testing.B) {

	const num_instance_of_logger = 100

	logs := make([]*log.Logger, num_instance_of_logger)

	fh, err := log.NewFileHandler("/dev/null")
	if err != nil {
		panic(err.Error())
	}
	if fh == nil {
		panic("fh is nil...............")
	}

	for i := 0; i < num_instance_of_logger; i++ {
		logs[i] = log.NewLogger(fh, log.StdLogFlag)
	}

	r := rand.New(rand.NewSource(99))
	for i := 0; i < b.N; i++ {
		logs[r.Intn(num_instance_of_logger)].Info(
			"xxxjk891p23j44 234 1=-idcxvjkxl;kj %x skdlfjasldkjf09adfjlk %d", i, i)
	}
	for i := 0; i < num_instance_of_logger; i++ {
		logs[i].Close()
	}
}

func BenchmarkMutilJsonLogger(b *testing.B) {

	const num_instance_of_logger = 100

	logs := make([]*log.Logger, num_instance_of_logger)

	// fd, _ := log.NewFileHandler("/dev/null")
	fd, err := log.NewFileHandler("/dev/null")
	if err != nil {
		panic(err.Error())
	}

	js := &log.JSONFormatter{}
	for i := 0; i < num_instance_of_logger; i++ {
		logs[i] = log.NewLogger(fd, log.StdLogFlag)
		logs[i].SetFormatter(js)
	}

	r := rand.New(rand.NewSource(99))
	for i := 0; i < b.N; i++ {
		logs[r.Intn(num_instance_of_logger)].Info(
			"xxxjk891p23j44 234 1=-idcxvjkxl;kj %x skdlfjasldkjf09adfjlk %d", i, i)
	}
	for i := 0; i < num_instance_of_logger; i++ {
		logs[i].Close()
	}
}

func BenchmarkJsonLogger(b *testing.B) {
	fd, err := log.NewFileHandler("/dev/null")
	if err != nil {
		panic(err.Error())
	}

	js := &log.JSONFormatter{}
	logger := log.NewLogger(fd, log.StdLogFlag)
	logger.SetFormatter(js)

	for i := 0; i < b.N; i++ {
		logger.WithField("k1111", "111111").
			WithField("k2222", 111111).
			WithField("k3333", "dfasdfasd").
			WithField("k4444", true).
			WithField("k5555", false).
			WithField("k7777", "7777777").
			Warn("format of [%0X/%0X]", i, b.N)
	}

	logger.Close()
}

func BenchmarkTexLineLogger(b *testing.B) {
	fd, err := log.NewFileHandler("/dev/null")
	if err != nil {
		panic(err.Error())
	}
	logger := log.NewLogger(fd, log.StdLogFlag)

	for i := 0; i < b.N; i++ {
		logger.
			Warn("Text line 9234905u3xcvjklxkj %s [%0X/%0X]",
				"02i34p[2o3ji4 lnxzckn kj sd;klfjsd format of",
				i, b.N)
	}
	logger.Close()
}

func BenchmarkTestMutilLogger(b *testing.B) {
	wg := sync.WaitGroup{}
	test_stdout := func() {
		fd, _ := log.NewFileHandler("/dev/null")
		logger := log.NewLogger(fd, log.StdLogFlag)
		for j := 0; j < b.N; j++ {
			logger.WithField("k1", "vv").
				Info("json: +++++++++++ go test_stdout() ++++++++++++++++++++++")
		}
		wg.Done()
	}
	cpu := runtime.NumCPU()
	if cpu > 1 {
		cpu -= 1
	}
	for i := 0; i < cpu; i++ {
		wg.Add(1)
		go test_stdout()
	}
	// time.Sleep(time.Millisecond * 100)
	wg.Wait()
}
