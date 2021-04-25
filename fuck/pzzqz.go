package fuck

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
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

var PZZQZDelay = 2 * time.Minute

func PZZQZ(ctx context.Context) (result []*shit.IP, delay time.Duration) {
	delay = PZZQZDelay

	var wg sync.WaitGroup
	var lk sync.Mutex

	for _, typ := range []string{"http", "socks4", "socks5"} {
		wg.Add(1)
		go func(typ string) {
			r := pzzqz_work(ctx, typ)
			if len(r) != 0 {
				lk.Lock()
				result = append(result, r...)
				lk.Unlock()
			}
			wg.Done()
		}(typ)
	}
	wg.Wait()

	logger.Info("[pzzqz.com] found %d records", len(result))
	return
}

func pzzqz_work(ctx context.Context, typ string) (result []*shit.IP) {
	u := "https://pzzqz.com"

	opt := client.NewOption()
	opt.Proxy = shit.Get("all").String()
	opt.Timeout = 10
	opt.Redirect = true
	opt.Headers["User-Agent"] = agent.Random()
	opt.Headers["x-requested-with"] = "XMLHttpRequest"

	cli, err := client.NewClient(opt, ctx)
	if err != nil {
		logger.Warn("[PZZQZ] Error = %v", err)
		return
	}

	resp, err := cli.Get(u)
	if err != nil {
		if err != context.Canceled {
			logger.Warn("[PZZQZ] Error = %v", err)
		}
		return
	}
	if resp == nil {
		logger.Warn("[PZZQZ] Error = Response is nil.")
		return
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		if err != context.Canceled {
			logger.Warn("[PZZQZ] Error = %v", err)
		}
		return
	}

	r, err := regexp.Compile(`"X-CSRFToken": "(.+?)"`)
	if err != nil {
		if err != context.Canceled {
			logger.Warn("[PZZQZ] Error = %v", err)
		}
		return
	}

	token := r.FindStringSubmatch(string(body))
	if len(token) < 2 {
		logger.Warn("[PZZQZ] Error = can't find the csrf token")
		return
	}
	cli.SetHeader("X-CSRFToken", token[1])

	d := map[string]string{
		"country": "all",
		"elite":   "on",
		"ping":    "200",
		"ports":   "",
	}
	d[typ] = "on"
	data, err := json.Marshal(d)
	if err != nil {
		logger.Warn("[PZZQZ] Error = can't find the csrf token")
		return
	}

	resp, err = cli.Post(u, "application/json", bytes.NewReader(data))
	if err != nil {
		if err != context.Canceled {
			logger.Warn("[PZZQZ] Error = %v", err)
		}
		return
	}
	if resp == nil {
		logger.Warn("[PZZQZ] Error = Response is nil.")
		return
	}

	result = pzzqz_parse(resp, typ)

	return
}

func pzzqz_parse(resp *http.Response, typ string) (result []*shit.IP) {
	de := json.NewDecoder(resp.Body)
	data := make(map[string]interface{})
	err := de.Decode(&data)
	resp.Body.Close()
	if err != nil {
		logger.Warn("[PZZQZ] Error = can't find the csrf token")
		return
	}

	var buf io.Reader
	if d, ok := data["proxy_html"]; !ok {
		logger.Warn("[PZZQZ] Error = wrong response")
		return
	} else if s, ok := d.(string); !ok {
		logger.Warn("[PZZQZ] Error = response not string")
		return
	} else {
		buf = strings.NewReader("<table>" + s + "</table>")
	}

	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		logger.Warn("[PZZQZ] Error = %v", err)
		return
	}

	if doc == nil {
		logger.Warn("[PZZQZ] Error = html document is nil")
		return
	}

	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		ip := shit.NewIP()
		ip.Source = "https://pzzqz.com"
		for ii, node := range s.Find("td").Nodes {
			text := strings.TrimSpace(s.FindNodes(node).Text())
			if ii == 0 {
				ip.Data = text
			}
			if ii == 1 {
				ip.Data += ":" + text
			}
			if ii == 4 {
				ip.Type = typ
				break
			}
		}
		result = append(result, ip)
	})

	return
}
