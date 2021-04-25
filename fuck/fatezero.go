package fuck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/s1mo0n/shitpool/pkg/agent"
	"github.com/s1mo0n/shitpool/pkg/client"
	"github.com/s1mo0n/shitpool/pkg/logger"
	"github.com/s1mo0n/shitpool/pkg/shit"
)

var FateZeroDelay = 2 * time.Minute

type FateZeroJSON struct {
	Type      string `json:"type"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Anonymity string `json:"anonymity"`
}

func FateZero(ctx context.Context) (result []*shit.IP, delay time.Duration) {
	delay = FateZeroDelay

	u := "http://proxylist.fatezero.org/proxy.list"

	opt := client.NewOption()
	opt.Proxy = shit.Get("all").String()
	opt.Timeout = 10
	opt.Redirect = false
	opt.Headers["User-Agent"] = agent.Random()

	cli, err := client.NewClient(opt, ctx)
	if err != nil {
		logger.Warn("[FateZero] Error = %v", err)
		return
	}

	resp, err := cli.Get(u)
	if err != nil {
		if err != context.Canceled {
			logger.Warn("[FateZero] Error = %v", err)
		}
		return
	}
	if resp == nil {
		logger.Warn("[FateZero] Error = Response is nil")
		return
	}

	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Warn("[FateZero] Error = %v", err)
		return
	}

	r, err := regexp.Compile(`\{.+?\}`)
	if err != nil {
		logger.Warn("[FateZero] Error = %v", err)
		return
	}

	obj := new(FateZeroJSON)

	for _, d := range r.FindAll(data, -1) {
		if err = json.Unmarshal(d, obj); err != nil {
			logger.Warn("[FateZero] Error = %v", err)
			continue
		}

		ip := shit.NewIP()
		ip.Source = u
		ip.Data = fmt.Sprintf("%s:%d", obj.Host, obj.Port)
		ip.Type = strings.ToLower(obj.Type)
		result = append(result, ip)
	}

	logger.Info("[FateZero] found %d records.", len(result))
	return
}
