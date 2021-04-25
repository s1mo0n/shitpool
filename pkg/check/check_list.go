package check

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/s1mo0n/shitpool/pkg/agent"
	"github.com/s1mo0n/shitpool/pkg/client"
	"github.com/s1mo0n/shitpool/pkg/logger"
	"github.com/s1mo0n/shitpool/pkg/shit"
)

var your_ip string

var (
	checkFuncs = []func(*shit.IP, ...context.Context) bool{
		CheckHTTPBIN,
		CheckBaidu,
		//CheckCountry,

		// last check
		CheckFuckWall,
	}
)

type HTTPBINJSON struct {
	Args   map[string]string `json:"args"`
	Origin string            `json:"origin"`
}

func Init() {
	logger.Info("[GetYourIP] Checking your IP...")
	opt := client.NewOption()
	opt.Redirect = false
	opt.Timeout = 10

	cli, err := client.NewClient(opt)
	if err != nil {
		logger.Error("[GetYourIP] Error = %v", err)
	}

	resp, err := cli.Get("https://httpbin.org/get?fuck=123")
	if err != nil {
		logger.Error("[GetYourIP] Error = %v", err)
	}

	var obj HTTPBINJSON
	de := json.NewDecoder(resp.Body)
	err = de.Decode(&obj)
	resp.Body.Close()
	if err != nil {
		logger.Error("[GetYourIP] Error = %v", err)
	}

	if v, ok := obj.Args["fuck"]; !ok && v != "123" {
		logger.Error("[GetYourIP] Error = base check function wrong")
	}

	your_ip = obj.Origin
	logger.Info("[GetYourIP] Your IP = %s", your_ip)
}

func CheckFuckWall(ip *shit.IP, ctx ...context.Context) (b bool) {
	// always return true
	b = true

	checkFuckURL := "http://google.com"

	opt := client.NewOption()
	opt.Proxy = ip.String()
	opt.Redirect = false
	opt.Timeout = 10

	cli, err := client.NewClient(opt, ctx...)
	if err != nil {
		logger.Warn("[CheckFuckWall] IP = %s Error = %v", ip, err)
		return
	}

	resp, err := cli.Head(checkFuckURL)
	if err == nil && resp != nil && (resp.StatusCode == 200 || resp.StatusCode == 301 || resp.StatusCode == 302) {
		resp.Body.Close()
		ip.Fuckwall = true
		return
	}

	ip.Fuckwall = false
	return
}

func CheckHTTPBIN(ip *shit.IP, ctx ...context.Context) bool {
	checkURL := "http://httpbin.org/get?fuck=123"

	opt := client.NewOption()
	opt.Proxy = ip.String()
	opt.Redirect = false
	opt.Timeout = 10

	cli, err := client.NewClient(opt, ctx...)
	if err != nil {
		logger.Debug("[CheckHTTPBIN] IP = %s Error = %v", ip, err)
		return false
	}

	begin := time.Now()

	resp, err := cli.Get(checkURL)
	if err != nil || resp == nil {
		if err == context.Canceled {
			return false
		}

		if ip.Type == "https" {
			logger.Debug("[CheckHTTPBIN] IP = %s, Change Type https -> http", ip)
			ip.Type = "http"
			return CheckHTTPBIN(ip, ctx...)
		}

		logger.Debug("[CheckHTTPBIN] testIP = %s, checkURL = %s: Error = %v", ip, checkURL, err)
		return false
	}

	ip.Speed = time.Now().Sub(begin).Nanoseconds() / 1000 / 1000 // ms

	var obj HTTPBINJSON
	de := json.NewDecoder(resp.Body)
	err = de.Decode(&obj)
	resp.Body.Close()
	if err != nil {
		logger.Debug("[CheckHTTPBIN] testIP = %s, checkURL = %s: Error = %v", ip, checkURL, err)
		return false
	}

	if v, ok := obj.Args["fuck"]; !ok && v != "123" {
		logger.Debug("[CheckHTTPBIN] testIP = %s, checkURL = %s: Error = wrong field response", ip, checkURL)
		return false
	}

	if strings.Contains(obj.Origin, your_ip) {
		return false
	}

	return true
}

func CheckBaidu(ip *shit.IP, ctx ...context.Context) bool {
	checkURL := "http://baidu.com"

	opt := client.NewOption()
	opt.Proxy = ip.String()
	opt.Redirect = false
	opt.Timeout = 10
	opt.Headers["User-Agent"] = agent.Random()

	cli, err := client.NewClient(opt, ctx...)
	if err != nil {
		logger.Debug("[CheckBaidu] IP = %s Error = %v", ip, err)
		return false
	}

	resp, err := cli.Get(checkURL)
	if err != nil {
		if err != context.Canceled {
			logger.Debug("[CheckBaidu] IP = %s Error = %v", ip, err)
		}
		return false
	}
	if resp == nil {
		logger.Warn("[CheckBaidu] IP = %s Error = Response is nil.", ip)
		return false

	}
	resp.Body.Close()

	switch resp.StatusCode {
	case 200, 301, 302:
		return true
	default:
		return false
	}
}

type IPAPIJSON struct {
	Country string `json:"country_code"`
}

func CheckCountry(ip *shit.IP, ctx ...context.Context) bool {
	opt := client.NewOption()
	opt.Proxy = ip.String()
	opt.Redirect = false
	opt.Timeout = 10
	opt.Headers["User-Agent"] = "curl"

	cli, err := client.NewClient(opt, ctx...)
	if err != nil {
		logger.Debug("[CheckCountry] IP = %s Error = %v", ip, err)
		return false
	}

	resp, err := cli.Get("http://ifconfig.io")
	if err != nil {
		if err != context.Canceled {
			logger.Debug("[CheckCountry] IP = %s Error = %v", ip, err)
		}
		return false
	}
	if resp == nil {
		logger.Warn("[CheckCountry] IP = %s Error = Response is nil.", ip)
		return false
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Debug("[CheckCountry] IP = %s Error = %v", ip, err)
		return false
	}

	resp, err = cli.Get("https://ipapi.com/ip_api.php?ip=" + strings.TrimSpace(string(body)))
	if err != nil {
		if err != context.Canceled {
			logger.Debug("[CheckCountry] IP = %s Error = %v", ip, err)
		}
		return false
	}
	if resp == nil {
		logger.Warn("[CheckCountry] IP = %s Error = Response is nil.", ip)
		return false
	}

	var obj IPAPIJSON
	de := json.NewDecoder(resp.Body)
	err = de.Decode(&obj)
	resp.Body.Close()
	if err != nil {
		logger.Debug("[CheckCountry] IP = %s Error = %v", ip, err)
		return false
	}

	ip.Country = obj.Country

	return true
}
