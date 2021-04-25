package check

import (
	"context"
	"sync"

	"github.com/s1mo0n/shitpool/pkg/config"
	"github.com/s1mo0n/shitpool/pkg/logger"
	"github.com/s1mo0n/shitpool/pkg/shit"
)

func CheckProxy(ip *shit.IP, ctx ...context.Context) {
	logger.Debug("[CheckProxy] Check IP = %s Type = %s", ip.Data, ip.Type)
	if CheckIP(ip, ctx...) {
		logger.Debug("[CheckProxy] Alive IP = %s Type = %s Speed = %dms FuckWall = %v", ip.Data, ip.Type, ip.Speed, ip.Fuckwall)
	}
}

func CheckProxyWork(ctx ...context.Context) {
	x := shit.Len("all")
	logger.Info("[CheckProxyWork] Before check: %d records.", x)
	ips := shit.GetAll("all")

	worker := config.Global.Worker
	if len(ips) < worker {
		worker = len(ips)
	}

	var wg sync.WaitGroup

	var c context.Context

	if len(ctx) > 0 {
		c = ctx[0]
	} else {
		c = context.TODO()
	}

	work_ch := make(chan *shit.IP, worker)
	for i := 0; i < worker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-c.Done():
					return
				case ip, ok := <-work_ch:
					if !ok {
						return
					}
					if !CheckIP(ip, c) {
						shit.Del(ip.Data)
					} else {
						logger.Debug("[CheckProxyWork] Alive IP = %s Type = %s Speed = %dms FuckWall = %v", ip.Data, ip.Type, ip.Speed, ip.Fuckwall)
					}
				}
			}
		}()
	}

	for _, ip := range ips {
		select {
		case work_ch <- ip:
		case <-c.Done():
			close(work_ch)
			wg.Wait()
			return
		}
	}
	close(work_ch)
	wg.Wait()

	x = shit.Len("all")
	logger.Info("[CheckProxyWork] After check: %d records.", x)
}

func CheckType(ip *shit.IP) bool {
	switch ip.Type {
	case "http", "https", "socks4", "socks5":
		return true
	default:
		logger.Warn("[CheckType] Error = unsupport proxy protocol %s", ip.Type)
		return false
	}
}

func CheckIP(ip *shit.IP, ctx ...context.Context) bool {
	var err error

	if !CheckType(ip) {
		return false
	}

	for _, f := range checkFuncs {
		if !f(ip, ctx...) {
			return false
		}
	}

	if err = shit.Add(ip); err != nil {
		logger.Warn("[CheckIP] Add IP = %v Error = %v", *ip, err)
	}

	return true
}
