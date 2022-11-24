# 日志记录器

## 概述

以太坊官方提供了**6**种级别的日志记录模式，分别是：

- Trace
- Debug
- Info
- Warn
- Error
- Crit

以上六个级别的日志等级从上到下逐渐递增。

此外，还支持将日志信息重定向到三种输出通道：

- 控制台
- 文件
- 网络连接

最后，支持两种日志输出格式：

- 控制台格式
- JSON格式
- 普通日志格式

## 使用方法

### 控制台格式输出日志

如果我们是在包外使用本包中定义的日志记录器，首先需要导入本包，然后按照下面的代码实例化一个日志记录器：

```go
l := log.New("blockchain", "ethereum")
l.SetHandler(StreamHandler(os.Stdout, TerminalFormat(true)))
```

上面代码里的`"blockchain"`和`"ethereum"`作为是一对键值对，以后每次使用`logger`输出日志时，都会打印这对键值对，然后第二行代码是用来设置输出
日志的处理器，这里我们设置将日志信息输出到操作系统的标准输出里，并且以控制台显示的格式输出，然后对于不同日志等级还会显式不同的颜色：

```go
l.Info("start service")
```

>输出：
> 
>INFO [01-01|00:00:00.000] start service                            blockchain=ethereum
> 
>ERROR[01-01|00:00:00.000] start service                            blockchain=ethereum

### JSON格式输出日志

实例化一个以JSON格式输出日志信息的日志记录器：

```go
l := New("blockchain", "ethereum")
l.SetHandler(StreamHandler(os.Stdout, JSONFormat()))
l.Info("start service")
l.Error("start service")
```

>输出
> 
>{"blockchain":"ethereum","lvl":"info","msg":"start service","t":"0001-01-01T00:00:00Z"}
> 
>{"blockchain":"ethereum","lvl":"eror","msg":"start service","t":"0001-01-01T00:00:00Z"}

### 普通日志格式

`LogfmtFormat()`函数定义了将日志信息按照普通日志格式打印的逻辑：

```go
l := New("blockchain", "ethereum")
l.SetHandler(StreamHandler(os.Stdout, LogfmtFormat()))
l.Trace("trace logger")
```

>输出：
> 
> t=0001-01-01T00:00:00Z lvl=trce msg="trace logger" blockchain=ethereum

### 将日志信息打印到文件里

```go
l := New("blockchain", "ethereum")
file, _ := os.OpenFile("text.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
l.SetHandler(StreamHandler(file, TerminalFormat(false)))
l.Info("start service")
l.Error("start service")
```

结果：

![image-20221124143154963](https://gitee.com/Sagaya815/assets/raw/master/image-20221124143154963.png)

### 将日志打印到网络连接通道里

这里我们利用`net`包建立了一对连接，然后在服务端把日志信息发送给客户端，客户端接收到以后再打印到控制台上：

```go
l := New("blockchain", "ethereum")
stopc := make(chan struct{})
server, err := net.Listen("tcp", "0.0.0.0:8080")
assert.Nil(t, err)
go func() {
	for {
		conn, err := server.Accept()
		assert.Nil(t, err)
		l.SetHandler(StreamHandler(conn, TerminalFormat(true)))
		l.Trace("welcome")
	}
}()

go func() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	assert.Nil(t, err)
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		assert.Nil(t, err)
		fmt.Println(string(buf[:n]))
		stopc <- struct{}{}
	}
}()
<-stopc
```

>输出：
> 
> client: TRACE[01-01|00:00:00.000] welcome                                  blockchain=ethereum

### 设置打印日志的级别

在下面的例子里，我们要求最多只打印`Warn`这一级别的日志，也就是说，`Trace Debug Info`这三个级别的日志不会被打印

```go
l := New("blockchain", "ethereum")
l.SetHandler(LvlFilterHandler(LvlWarn, StreamHandler(os.Stdout, TerminalFormat(true))))
l.Info("info logger")
l.Warn("warn logger")
l.Error("error logger")
```

>输出：
> 
>WARN [01-01|00:00:00.000] warn logger                              blockchain=ethereum
> 
>ERROR[01-01|00:00:00.000] error logger                             blockchain=ethereum