package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/zwcway/castserver-go/common"
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/control"
	"github.com/zwcway/castserver-go/detector"
	"github.com/zwcway/castserver-go/mutexer"
	"github.com/zwcway/castserver-go/pusher"
	"github.com/zwcway/castserver-go/receiver"
	"github.com/zwcway/castserver-go/utils"
	"github.com/zwcway/castserver-go/web"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	options         map[string]string
	version         bool
	help            bool
	daemon          bool
	is6             bool
	logFile         string
	configFile      string
	netInterface    string
	detectInterface string
)

func init() {
	flag.StringVar(&configFile, "c", "", "specify configuration file")
	flag.StringVar(&netInterface, "i", "", "listen interface")
	flag.StringVar(&detectInterface, "detect-interface", "", "detect listen interface")
	flag.StringVar(&logFile, "l", "", "log file")
	flag.BoolVar(&version, "v", false, "show current version of clash")
	flag.BoolVar(&daemon, "D", false, "running in background")
	flag.BoolVar(&help, "h", false, "show this message")
	flag.BoolVar(&is6, "6", false, "use IPV6 net")
	flag.Parse()

	options = make(map[string]string)
	flag.Visit(func(f *flag.Flag) { options[f.Name] = f.Value.String() })
}

func exit(code int, usage bool, format string, val ...any) {
	fmt.Printf(format, val...)
	if usage {
		os.Stdout.Sync()
		flag.Usage()
	}
	os.Exit(code)
}

func initLogger() (log *zap.Logger, close func()) {
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	topicDebugging := zapcore.AddSync(io.Discard)
	topicErrors := zapcore.AddSync(io.Discard)
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	FileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	cores := []zapcore.Core{}

	if len(logFile) > 0 {
		// TODO 接收信号重新打开日志文件
		var (
			sink zapcore.WriteSyncer
			err  error
		)
		sink, close, err = zap.Open(logFile)
		if err != nil {
			close()
			exit(1, true, "open log file error %v", err)
		}
		cores = append(
			cores,
			zapcore.NewCore(FileEncoder, sink, highPriority),
			zapcore.NewCore(FileEncoder, sink, lowPriority),
		)
	} else {
		cores = append(cores,
			zapcore.NewCore(FileEncoder, topicErrors, highPriority),
			zapcore.NewCore(FileEncoder, topicDebugging, lowPriority),
		)
	}

	if !daemon {
		cores = append(cores,
			zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
			zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
		)
	}

	log = zap.New(zapcore.NewTee(cores...))

	return
}

func main() {
	if version {
		exit(0, false, "Version: %s", config.VERSION)
	}
	if help {
		exit(0, true, "")
	}

	log, logClose := initLogger()

	// 用于通知主程序退出
	signalChannel := make(chan os.Signal, 2)

	// 用于通知子协程退出
	rootCtx, ctxCancel := utils.NewContext().WithSignal(signalChannel).WithLogger(log).WithCancel()

	err := config.FromOptions(log, options)
	if err != nil {
		exit(2, false, err.Error())
	}

	debug.SetMaxThreads(config.RuntimeThreads)
	common.Init(rootCtx)

	mods := []Module{
		mutexer.Module,
		detector.Module,
		control.Module,
		pusher.Module,
		receiver.Module,
		web.Module,
	}

	for _, f := range mods {
		err = f.Init(rootCtx)
		if err != nil {
			exit(255, false, err.Error())
		}
	}

	// 阻塞
	signal.Notify(signalChannel,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
	)

	<-signalChannel
	ctxCancel()
	close(signalChannel)

	log.Info("exit")

	for _, f := range mods {
		f.DeInit()
	}

	if logClose != nil {
		logClose()
	}

	os.Exit(0)
}
