package fuck

import (
	"context"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/s1mo0n/shitpool/pkg/agent"
	"github.com/s1mo0n/shitpool/pkg/client"
	"github.com/s1mo0n/shitpool/pkg/logger"
	"github.com/s1mo0n/shitpool/pkg/shit"
)

var SpysDelay = 1 * time.Hour

func Spys(ctx context.Context) (result []*shit.IP, delay time.Duration) {
	delay = SpysDelay
	u := "https://spys.me/proxy.txt"

	opt := client.NewOption()
	opt.Proxy = shit.Get("all").String()
	opt.Timeout = 10
	opt.Redirect = false
	opt.Headers["User-Agent"] = agent.Random()

	cli, err := client.NewClient(opt, ctx)
	if err != nil {
		logger.Warn("[Spys] Error = %v", err)
		return
	}

	resp, err := cli.Get(u)
	if err != nil {
		if err != context.Canceled {
			logger.Warn("[Spys] Error = %v", err)
		}
		return
	}
	if resp == nil {
		logger.Warn("[Spys] Error = Response is nil.")
		return
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Warn("[Spys] Error = %v", err)
		return
	}

	r, err := regexp.Compile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:\d{1,5} \w{2}-\w{1}(-S)*`)
	if err != nil {
		logger.Warn("[Spys] Error = %v", err)
		return
	}

	data := r.FindAllString(string(body), -1)
	if len(data) == 0 {
		logger.Warn("[Spys] Error = Can't find anything, maybe it's a shit.")
		return
	}

	for _, s := range data {
		ip := shit.NewIP()
		ip.Source = u
		d := strings.Split(s, " ")
		if len(d) != 2 {
			logger.Warn("[Spys] Error = Something wrong, maybe it's a shit.")
			return
		}
		ip.Data = strings.TrimSpace(d[0])
		ip.Type = "http"
		dd := strings.Split(d[1], "-")
		if len(dd) < 2 {
			logger.Warn("[Spys] Error = Something wrong, maybe it's a shit.")
			return
		}
		if len(dd) > 2 && dd[2] == "S" {
			ip.Type += "s"
		}
		result = append(result, ip)
	}

	return
}
