package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/s1mo0n/shitpool/fuck"

	"github.com/s1mo0n/shitpool/mod"
	"github.com/s1mo0n/shitpool/mod/api"
	"github.com/s1mo0n/shitpool/mod/proxy"

	"github.com/s1mo0n/shitpool/pkg/check"
	"github.com/s1mo0n/shitpool/pkg/config"
	"github.com/s1mo0n/shitpool/pkg/logger"
	"github.com/s1mo0n/shitpool/pkg/shit"
)

type FuckFunc func(ctx context.Context) ([]*shit.IP, time.Duration)

var (
	funcs = []FuckFunc{
		fuck.PZZQZ, // it's too shit, maybe you can delete it.
		fuck.Laravel,
		fuck.FateZero,
		fuck.FreeProxy,
		fuck.Spys,
	}
)

func banner() {
	fmt.Fprintln(os.Stderr, "\x1b[32m ____  _     _ _   ____             _ \n/ ___|| |__ (_) |_|  _ \\ ___   ___ | |\n\\___ \\| '_ \\| | __| |_) / _ \\ / _ \\| |\n ___) | | | | | |_|  __/ (_) | (_) | |\n|____/|_| |_|_|\\__|_|   \\___/ \\___/|_|\n                                      \n\x1b[0m")

}

func usage() {
	fmt.Fprintln(os.Stderr, `
Commands:
	api				api mode
	proxy			proxy mode
	mix				api & proxy mode

Options:
	--data FILE			Data File to load and save
	--worker INT		Concurrent Worker
	--chk-delay DELAY	Check IP delay minute (default: 10 minute)
	--log-level LEVEL	Log Level "ERROR", "WARN", "INFO", "DEBUG", "SKIP" (default: "INFO")
	--help, -h			Show help
	`)
}

func parser() []mod.Server {
	flag.IntVar(&config.Global.Worker, "worker", 100, "")
	flag.IntVar(&config.Global.ChkDelay, "chk-delay", 10, "")
	flag.StringVar(&config.Global.DataFile, "data", "shitpool.json", "")
	flag.StringVar(&config.Global.LogLevel, "log-level", "INFO", "")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
		os.Exit(2)
	}

	logger.SetLevel(config.Global.LogLevel)
	shit.Parse(config.Global.DataFile)

	var server []mod.Server
	switch flag.Arg(0) {
	case "api":
		var addr string
		var logpath string

		subflag := flag.NewFlagSet("api", flag.ExitOnError)
		subflag.StringVar(&addr, "addr", "localhost:3333", "API Listen Address (default: localhost:3333)")
		subflag.StringVar(&logpath, "log", "api.log", "API Log Path")
		subflag.Parse(flag.Args()[1:])

		server = []mod.Server{api.NewAPI(addr, logpath)}

	case "proxy":
		var retry int
		var typ string
		var saddr string
		var haddr string
		var logpath string

		subflag := flag.NewFlagSet("proxy", flag.ExitOnError)
		subflag.IntVar(&retry, "retry", 5, "Proxy IP Retry Count (default: 5)")
		subflag.StringVar(&typ, "type", "all", "Proxy Type ('all', 'fuck', 'http', 'https', 'socks5')")
		subflag.StringVar(&haddr, "httpaddr", "localhost:4444", "HTTP Listen Address (default: localhost:4444)")
		subflag.StringVar(&saddr, "socksaddr", "", "Socks Listen Address (better not to use..., no https...)")
		subflag.StringVar(&logpath, "log", "proxy.log", "Proxy Log Path")

		subflag.Parse(flag.Args()[1:])

		server = []mod.Server{proxy.NewProxy(haddr, saddr, typ, retry, logpath)}

	case "mix":
		var apiaddr string
		var apilogpath string

		var pxyretry int
		var pxytyp string
		var pxysaddr string
		var pxyhaddr string
		var pxylogpath string

		subflag := flag.NewFlagSet("api", flag.ExitOnError)
		subflag.StringVar(&apiaddr, "apiaddr", "localhost:3333", "API Listen Address (default: localhost:3333)")
		subflag.StringVar(&apilogpath, "apilog", "api.log", "API Log Path")

		subflag.IntVar(&pxyretry, "pxyretry", 5, "Proxy IP Retry Count (default: 5)")
		subflag.StringVar(&pxytyp, "pxytype", "all", "Proxy Type ('all', 'fuck', 'http', 'https', 'socks5')")
		subflag.StringVar(&pxyhaddr, "httpaddr", "localhost:4444", "HTTP Listen Address (default: localhost:4444)")
		subflag.StringVar(&pxysaddr, "socksaddr", "", "Socks Listen Address (better not to use...no https...)")
		subflag.StringVar(&pxylogpath, "pxylog", "proxy.log", "Proxy Log Path")

		subflag.Parse(flag.Args()[1:])

		apiserver := api.NewAPI(apiaddr, apilogpath)
		pxyserver := proxy.NewProxy(pxyhaddr, pxysaddr, pxytyp, pxyretry, pxylogpath)

		server = []mod.Server{apiserver, pxyserver}

	default:
		os.Exit(0)
	}

	return server
}

func worker(ctx context.Context) {
	for {
		check.CheckProxyWork(ctx)
		ips := shit.GetAll("all")

		data, err := json.Marshal(&ips)
		if err != nil {
			panic(fmt.Errorf("[main.worker] Error = %v", err))
		}
		os.WriteFile(config.Global.DataFile, data, 0644)

		select {
		case <-ctx.Done():
			logger.Info("[make.worker] Done.")
			return
		case <-time.After(time.Duration(config.Global.ChkDelay) * time.Minute):
		}
	}
}

func fucker(ctx context.Context, fn FuckFunc) {
	for {
		ips, delay := fn(ctx)

		if len(ips) != 0 {
			var wg sync.WaitGroup
			worker := config.Global.Worker / 2
			if len(ips) < worker {
				worker = len(ips)
			}
			work_ch := make(chan *shit.IP)
			for i := 0; i < worker; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						select {
						case <-ctx.Done():
							return
						case ip, ok := <-work_ch:
							if !ok {
								return
							}
							check.CheckProxy(ip, ctx)
						}
					}
				}()
			}
			for _, ip := range ips {
				select {
				case work_ch <- ip:
				case <-ctx.Done():
					return
				}
			}
			close(work_ch)
			wg.Wait()
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}

func runner(ctx context.Context) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		worker(ctx)
		wg.Done()
	}()

	for _, fn := range funcs {
		wg.Add(1)
		go func(fn FuckFunc) {
			fucker(ctx, fn)
			wg.Done()
		}(fn)
	}

	wg.Wait()
}

func main() {
	banner()

	servers := parser()
	for _, server := range servers {
		go server.Run()
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer func() {
		cancel()
		if err := recover(); err != nil {
			logger.Error("[main.runner] Error = %v", err)
		}
	}()

	check.Init()
	runner(ctx)
}
