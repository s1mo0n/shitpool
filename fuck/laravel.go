package fuck

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/s1mo0n/shitpool/pkg/agent"
	"github.com/s1mo0n/shitpool/pkg/client"
	"github.com/s1mo0n/shitpool/pkg/logger"
	"github.com/s1mo0n/shitpool/pkg/shit"
)

var LaravelDelay = 2 * time.Minute

func Laravel(ctx context.Context) (result []*shit.IP, delay time.Duration) {
	delay = LaravelDelay

	paths := []string{"/premium", "/stable"}

	var lk sync.Mutex
	var wg sync.WaitGroup

	for _, path := range paths {
		wg.Add(1)
		go func(p string) {
			r := laravel_work(ctx, p)
			lk.Lock()
			result = append(result, r...)
			lk.Unlock()
			wg.Done()
		}(path)
	}
	wg.Wait()

	logger.Info("[Laravel] found %d records.", len(result))
	return
}

func laravel_work(ctx context.Context, p string) (result []*shit.IP) {
	u := "http://39.108.11.65:8085"
	u += p

	opt := client.NewOption()
	opt.Proxy = shit.Get("all").String()
	opt.Timeout = 10
	opt.Redirect = true
	opt.Headers["User-Agent"] = agent.Random()

	cli, err := client.NewClient(opt, ctx)
	if err != nil {
		logger.Warn("[Laravel] Error = %v", err)
		return
	}

	resp, err := cli.Get(u)
	if err != nil {
		if err != context.Canceled {
			logger.Warn("[Laravel] Error = %v", err)
		}
		return
	}
	if resp == nil {
		logger.Warn("[Laravel] Error = Response is nil.")
		return
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	resp.Body.Close()
	if err != nil {
		logger.Warn("[Laravel] Error = %v", err)
		return
	}

	if doc == nil {
		logger.Warn("[Laravel] Error = html document is nil")
		return
	}

	pages := make(map[string]struct{})
	doc.Find("a.page-link").Each(func(i int, s *goquery.Selection) {
		a, ok := s.Attr("href")
		if ok {
			pages[a] = struct{}{}
		}
	})
	result = laravel_parse(doc)

	for page := range pages {
		resp, err := cli.Get(u + page)
		if err != nil {
			if err == context.Canceled {
				return
			}
			logger.Warn("[Laravel] Error = %v", err)
		}
		if resp == nil {
			logger.Warn("[Laravel] Error = Response is nil")
			return
		}

		doc, err := goquery.NewDocumentFromResponse(resp)
		resp.Body.Close()
		if err != nil {
			logger.Warn("[Laravel] Error = %v", err)
		}

		result = append(result, laravel_parse(doc)...)
	}
	return
}

func laravel_parse(doc *goquery.Document) (result []*shit.IP) {
	if doc == nil {
		logger.Warn("[Laravel] Error = html document is nil")
		return
	}

	doc.Find("tbody").Find("tr").Each(func(_ int, s *goquery.Selection) {
		ip := shit.NewIP()
		ip.Source = "http://www.66ip.cn/mo.php"
		s.Find("td").EachWithBreak(func(i int, s *goquery.Selection) bool {
			if i == 0 {
				ip.Data = s.Text()
			}
			if i == 1 {
				ip.Data += ":" + s.Text()
			}

			if i == 3 {
				ip.Type = strings.ToLower(s.Text())
				result = append(result, ip)
				return false
			}
			return true
		})
	})

	return
}
