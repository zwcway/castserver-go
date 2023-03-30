package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/database"
	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/utils"
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

func main() {
	exitCode := 0

	if version {
		exit(exitCode, false, "Version: %s", config.VERSION)
	}
	if help {
		exit(exitCode, true, "")
	}

	log, logClose, err := lg.NewLogger(logFile, daemon)
	if err != nil {
		exit(2, true, err.Error())
	}

	// 用于通知主程序退出
	signalChannel := make(chan os.Signal, 2)

	// 用于通知子协程退出
	rootCtx, ctxCancel := utils.NewContext().WithSignal(signalChannel).WithLogger(log).WithCancel()

	err = config.FromOptions(rootCtx, options)
	if err != nil {
		exit(2, false, err.Error())
	}

	debug.SetMaxThreads(config.RuntimeThreads)

	database.Init(rootCtx, config.DB)

	err = initModules(rootCtx)
	if err != nil {
		exitCode = 255
		goto __exit__
	}
	err = startModules()
	if err != nil {
		exitCode = 254
		goto __exit__
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

__exit__:
	if err != nil {
		fmt.Println(err.Error())
	}

	deinitModules()

	if logClose != nil {
		logClose()
	}

	os.Exit(exitCode)
}
