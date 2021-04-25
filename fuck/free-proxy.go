package fuck

import (
	"context"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/s1mo0n/shitpool/pkg/agent"
	"github.com/s1mo0n/shitpool/pkg/client"
	"github.com/s1mo0n/shitpool/pkg/logger"
	"github.com/s1mo0n/shitpool/pkg/shit"
)

var FreeProxyDelay = 2 * time.Minute

func FreeProxy(ctx context.Context) (result []*shit.IP, delay time.Duration) {
	delay = FreeProxyDelay

	u := map[string]string{
		"http":  "https://free-proxy-list.net/",
		"socks": "https://www.socks-proxy.net/",
	}

	var lk sync.Mutex
	var wg sync.WaitGroup

	for k, v := range u {
		wg.Add(1)
		go func(typ, u string) {
			r := freeproxy_parse(ctx, typ, u)
			if len(r) > 0 {
				lk.Lock()
				result = append(result, r...)
				lk.Unlock()
			}
			wg.Done()
		}(k, v)
	}

	wg.Wait()

	logger.Info("[FreeProxy] found %d records.", len(result))
	return
}

func freeproxy_parse(ctx context.Context, typ, u string) (result []*shit.IP) {
	opt := client.NewOption()
	opt.Timeout = 10
	opt.Redirect = true
	opt.Headers["User-Agent"] = agent.Random()

	opt.Proxy = shit.Get("fuck").String()

	cli, err := client.NewClient(opt)
	if err != nil {
		logger.Warn("[FreeProxy] Error = %v", err)
		return
	}

	resp, err := cli.Get(u)
	if err != nil {
		if err != context.Canceled {
			logger.Warn("[FreeProxy] Error = %v", err)
		}
		return
	}
	if resp == nil {
		logger.Warn("[FreeProxy] Error = Response is nil.")
		return
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Warn("[FreeProxy] Error = %v", err)
		return
	}

	r, err := regexp.Compile(`<tbody>.+?</tbody>`)
	if err != nil {
		logger.Warn("[FreeProxy] Error = %v", err)
		return
	}

	data := r.FindString(string(body))
	if len(data) == 0 {
		logger.Warn("[FreeProxy] Error = response format wrong")
		return
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<table>" + data + "</table>"))
	if err != nil {
		logger.Warn("[FreeProxy] Error = %v", err)
		return
	}

	if doc == nil {
		logger.Warn("[FreeProxy] Error = html document is nil")
		return
	}

	doc.Find("tbody").Find("tr").Each(func(_ int, s *goquery.Selection) {
		ip := shit.NewIP()
		ip.Source = u
		s.Find("td").EachWithBreak(func(i int, s *goquery.Selection) bool {
			text := strings.TrimSpace(s.Text())
			if i == 0 {
				ip.Data = text
			}
			if i == 1 {
				ip.Data += ":" + text
			}
			if i == 4 && typ == "socks" {
				ip.Type = strings.ToLower(text)
				result = append(result, ip)
				return false
			}
			if i == 6 {
				ip.Type = "http"
				if text == "yes" {
					ip.Type += "s"
				}
				result = append(result, ip)
				return false
			}
			return true
		})
	})

	return
}
