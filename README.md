---
title: "wps.cn log4go 模块"
date: 2018-08-03
author: HuangChuanTonG@WPS.cn
---

# wps.cn log4go 模块

## 背景和问题

    线上QPS非常高，日志输出过多、过快会严重影响性能。
    如果多个请求(线程)同时写日志，硬盘IO或网络IO都支持不住。
    为了限制写日志速度，并且允许在高压条件下，丢弃日志。
    同时简化写日志代码，并支持输出为json格式。

## 设计原理：

1. 默认只有一个IO 线程在后台执行写操作(代码handler_iothread.go)。
2. 由于只有单个chan接收写入日志请求，有可能会丢日志，包级别提供了接口 SetDropCallback(DropLogCallbackFunc)，当出现chan满，不得不丢日志时，马上回调。
3. 由于写IO的存在，所以log.Info/Warn/Error 等函数都是马上返回，但日志全延迟输出。
4. 可以对Handler 自定义IO线程，参看 SetWriteIOThread() 。
5. 程序退出，必需调用包级别的 log.Close()，让后台线程时把log Flush完。
6. 单条日志的msg限制最大字节数：log.MAX_BYTES_PER_LOG = 1024 * 3


### 类关系：

- Logger    -- 提供输出日志方法，如Info/Warn/Error。
- Handlder  -- Logger可有多个handler，一个handler只能有1个IO线程。
- Formatter -- 日志输出格式化类，纯方法类, 一个Logger只能有一个Formatter。
- HandlerIOThread -- - Handler的io线程，包默认有1个后台线程。一个Handler只能一个io。
- LogInstance -- 单条日志对象(未格式化前)，Logger.Info/Warn/Error函数后生成。


### 正常使用方式：
```
import(
   log "wps.cn/you-gopath/log4go"
)

log.Info("hello world") 
// 默认输出格式：时间 - 级别 - 调用代码文件与行数 - 日志内容
// 2018/08/03 10:26:21 - INFO - go/log/log_test.go:[12] - hello world

Info/Warn/Error等方法，默认支持Sprintf式，如：
log.Info("hello world a=%s b=%v", a, b)

log模块接口是没用python-logging的，包含内部某些功能实现是也抄logging。
所以接口并没像go的stlye: 支持fmt的，使用Infof，f结尾。
```

### 包级别的方法：
```
import(
   log "wps.cn/you-gopath/log4go"
)
// 定义一个丢日志的回调函数：
func dropLogCallback(l *log.LogInstance, sum int64) {
    // io线程chan满了，必需丢弃日志，请通过诸如普罗米修斯进行收集报警
    fmt.Printf("drop-log=%v, drop-sum=%v", l, sum)
}

func main(){
    // 退出前必需Close(),通知IO线程写完盘。
    defer log.Close()

    log.SetLevelS("ERROR") 

    // 给默认的IO线程设置一个丢日志时的回调通知。
    log.SetDropCallback(dropLogCallback) 

    // 默认限制单条日志最大3k，可通过改此值
    log.MAX_BYTES_PER_LOG = 1024 * 2  
}

{// 用法1: 默认包级别的Info/Warn/Error等输出日志方法，都是输出到stdout；
 // 调用包级别方法SetHandler，会改变默认包级Info/Warn/Error 输出到文件中。
    hdlr, err := log.NewRotatingFileHandler(fileName, maxBytes, backupCount)
    log.SetHandler(hdlr)
}

{// 用法2：一条log，多个Handler输出
 // 调用AppendHandler后，包级会有2个Hanlder，这里同时输出到: stdout与文件
    hdlr2, err := log.NewStreamHandler(os.Stdout)
    log.AppendHandler(hdlr2)
    log.Info(".....")    // 这里同时输出到: stdout与文件。
}

{// 用法3，为独立的文件handler设置单独IO线程：
    hdlr3, err := log.NewTimeRotatingFileHandler(baseName, when, interval)

    // 退出前调用 Handler.Close()，会同时Close对应的IOThread。
    // 如果handler没有使用自定义的IO线程，可以不用Close()。
    defer hdlr3.Close()

    ioTh := NewHandleIOWriteThread(name="you-io", chanLength=8192)
    ioTh.SetDropCallback(dropLogCallback)
    hdlr3.SetWriteIOThread(ioTh)
}
```
### 改成 JSON 格式输出方法：
```
//方法一, package级别的WithField或WithFields, 会返回一个新logger实例
log.WithField("k1", "myVal_1").Info("hello JSON")

// 方法二
{
    logger_1 = log.WithField("reqId", 123) // 返回一个新实例
    logger_1.Info("INFO JSON")  // 输出会带上 {"reqId": 123}
    logger_1.Warn("WANR JSON")  // 输出会带上 {"reqId":123}
}

// 方法三，自己构建logger实例：
{
    hdlr, _ := log.NewStreamHandler(os.Stdout)
    logger := log.NewLogger(hdlr, Ltime|Lfile|Llevel)

    //添加Formatter后, 此后这个logger输出的日志，永远是JSON
    logger.SetFormatter(&JSONFormatter{})  

    // NewLogger() 新建出来的logger实例
    // 如果handler没有使用自定义的IO线程，可以不用Close()
    // 反之，有自定IO线程，则需要在退出时调用Close(),
    // 才能保证log能完整输出。
    // logger.Close() 会把内部handlers全Close(), 所以不必再一一close。
    logger.Close()  
}

// 方法四，如果是package级别，添加了JSONFormatter，
// 则刚对所有包级log.Info/Warn等调用，都会变成JSON格式输出。
log.SetFormatter(&JSONFormatter{}) 
log.Info("alway json ....")
```

### 关于 WithField / WithFields 用法

何一个logger只要调用 WithField/WithFields 紧接着的 Info/Warn/Error
输出格式都会自动变成 json格式。

```
import(
   log "wps.cn/you-gopath/log4go"
)

log.WithField("k1", "v1").WithField("k2","v22").Info("I am log msg.")
// 输出：
// {"file":"go/log/log_test.go:[16]","k1":"v1","k2":"v22","level":"INFO","msg":"I am log msg.","time":"2018/08/03 10:59:15"}


log.WithFields(log.Fields{
        "k1": 1,
        "k2": 2,
        "k3": 3,
    }).Info("json INFO log k4=%v", 44)

// 输出: {"file":"go/log/log_test.go:[37]","k1":1,"k2":2,"k3":3,"level":"INFO","msg":"json INFO log k4=44","time":"2018/08/03 11:04:59"}

```

### 工程中，需要对输出 log 打tags的，建议使用模块实例，如
```
import(
   log "wps.cn/you-gopath/log4go"
)

type TB_mgr struct{
    logger *Logger
}

TB_mgr.logger = log.WithField("tags": "TB_mgr")

func (t *TB_mgr)func1(){
    // 输出的log就一定会带上tags:"TB_mgr"
    t.logger.Warn("I'm warning log....")
}

```


### 使用json时，注意：

使用json输出时，默认占用了以下几个key：
即不要使用 log.WithField("file": "xxxx").Info("xxx") 
file对应的值，会被内置替换，无法输出xxxx。

logrus的做法是：给key加前辍；但我暂时不打算修复这个bug。

```
const (
    keyFileNo = "file"
    keyTime   = "time"
    keyMsg    = "msg"
    keyLevel  = "level"
)
```
