package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/zwcway/castserver-go/common/speaker"
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
	options       map[string]string
	version       bool
	help          bool
	daemon        bool
	is6           bool
	logLevel      string
	multicastIP   string
	multicastPort int64
	configFile    string
	netInterface  string
)

func init() {
	flag.StringVar(&multicastIP, "multicast-ip", "", "listen ip")
	flag.Int64Var(&multicastPort, "multicast-port", 0, "listen port")
	flag.StringVar(&configFile, "c", "", "specify configuration file")
	flag.StringVar(&netInterface, "i", "", "listen interface")
	flag.StringVar(&logLevel, "l", "", "log level")
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

func initLogger() *zap.Logger {
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

	FileEncoder := zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig())
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(FileEncoder, topicErrors, highPriority),
		zapcore.NewCore(FileEncoder, topicDebugging, lowPriority),
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
	)
	log := zap.New(core)

	return log
}

func main() {
	log := initLogger()

	// 用于通知主程序退出
	signalChannel := make(chan os.Signal, 2)

	// 用于通知子协程退出
	rootCtx, ctxCancel := utils.NewContext().WithSignal(signalChannel).WithLogger(log).WithCancel()

	err := config.FromOptions(log, options)
	if err != nil {
		exit(2, false, err.Error())
	}

	debug.SetMaxThreads(config.RuntimeThreads)

	mods := []Module{
		mutexer.Module,
		detector.Module,
		control.Module,
		pusher.Module,
		receiver.Module,
		web.Module}

	speaker.Init()
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
		syscall.SIGABRT)

	<-signalChannel
	ctxCancel()
	close(signalChannel)

	log.Info("exit")

	for _, f := range mods {
		f.DeInit()
	}

	os.Exit(0)
}
