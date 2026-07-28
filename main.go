package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"

	"github.com/go-ini/ini"
	"github.com/zwcway/castserver-go/common"
	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/i18n"
	"github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/utils"
)

var (
	options       map[string]string
	optVersion    bool
	optHelp       bool
	optDaemon     bool
	optLogFile    string
	optConfigFile string
)

func init() {
	flag.StringVar(&optLogFile, "l", "", "log file")
	flag.BoolVar(&optVersion, "v", false, "show current version of clash")
	flag.BoolVar(&optDaemon, "D", false, "running in background")
	flag.BoolVar(&optHelp, "h", false, "show this message")

	var (
		optIsIPV6          bool
		optNetInterface    string
		optDetectInterface string
	)
	flag.StringVar(&optConfigFile, "c", "", "specify configuration file")
	flag.StringVar(&optNetInterface, "i", "", "listen interface")
	flag.StringVar(&optDetectInterface, "detect-interface", "", "detect listen interface")
	flag.BoolVar(&optIsIPV6, "6", false, "use IPV6 net")
}

func parseConfigFileToOptions(path string, opts map[string]string) error {
	if path == "" {
		return nil
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("The file '%s' not exists.", path)
		}
		return err
	}

	cfg, err := ini.Load(path)
	if err != nil {
		return err
	}

	defaultSection := cfg.Section(ini.DefaultSection)
	for _, key := range defaultSection.Keys() {
		name := key.Name()
		if _, exists := opts[name]; exists {
			continue
		}
		opts[name] = key.Value()
	}

	return nil
}

func exit(code int, usage bool, format string, val ...any) {
	if format != "" {
		fmt.Printf(format, val...)
	}

	if usage {
		os.Stdout.Sync()
		flag.Usage()
	}
	os.Exit(code)
}

func loadConfig() {
	configDir := filepath.Dir(optConfigFile)
	if configDir == "" {
		configDir = "."
	}

	if err := parseConfigFileToOptions(optConfigFile, options); err != nil {
		exit(2, false, "%s", err.Error())
	}
	_ = i18n.Load(filepath.Join(configDir, "locales"))

	if v, ok := options["language"]; ok && v != "" {
		i18n.SetLanguage(v)
	} else {
		i18n.SetLanguage("en")
	}
}

func main() {
	exitCode := 0

	flag.Parse()
	options = make(map[string]string)
	flag.Visit(func(f *flag.Flag) { options[f.Name] = f.Value.String() })

	if optVersion {
		exit(exitCode, false, "Version: %s", config.VERSION)
	}
	if optHelp {
		exit(exitCode, true, "")
	}

	loadConfig()

	logger, logClose, err := log.NewLogger(optLogFile, optDaemon)
	if err != nil {
		exit(2, true, "%s", err.Error())
	}
	// 用于通知主程序退出
	signalChannel := make(chan os.Signal, 2)

	// 用于通知子协程退出
	rootCtx, ctxCancel := utils.NewContext().WithSignal(signalChannel).WithLogger(logger).WithCancel()

	err = config.FromOptions(rootCtx, options)
	if err != nil {
		exit(2, false, "%s", err.Error())
	}

	debug.SetMaxThreads(config.RuntimeThreads)

	common.Init(rootCtx)

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
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
	)

	<-signalChannel
	ctxCancel()
	close(signalChannel)

	logger.Info("exit")

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
