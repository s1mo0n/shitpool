package proxy

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/elazarl/goproxy"
	"github.com/s1mo0n/shitpool/pkg/logger"
	"github.com/s1mo0n/shitpool/pkg/shit"
)

var pxylogger *log.Logger

type Proxy struct {
	saddr string
	haddr string
	retry int
	pxtyp string
}

func NewProxy(haddr, saddr, pxtyp string, retry int, logpath string) *Proxy {
	writer, err := os.OpenFile(logpath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		logger.Error("[Proxy] Error = %v", err)
	}
	pxylogger = log.New(writer, "[Proxy]", log.LstdFlags)

	return &Proxy{
		saddr: saddr,
		haddr: haddr,
		pxtyp: pxtyp,
		retry: retry,
	}
}

func (p *Proxy) Run() {
	go p.HTTP()

	if len(p.saddr) != 0 {
		go p.Socks()
	}
}

func (p *Proxy) HTTP() {
	logger.Info("[Proxy] HTTP Server Start")
	err := http.ListenAndServe(p.haddr, p)
	logger.Error("[Proxy] Error = %v", err)
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.Logger = pxylogger

	for i := 0; i < p.retry; i++ {
		ip := shit.Get(p.pxtyp)

		if len(ip.Data) == 0 {
			logger.Warn("[Proxy] Error = shitpool have no shit")
			w.WriteHeader(500)
			w.Write([]byte("no shit"))
			return
		}

		u, err := url.Parse(ip.String())
		if err == nil {
			proxy.Tr.Proxy = http.ProxyURL(u)
			break
		}

		shit.Del(ip.Data)
		logger.Warn("[Proxy] IP = %s, Error = %v", ip, err)
	}

	proxy.ServeHTTP(w, r)
}

func (p *Proxy) Socks() {
	logger.Info("[Proxy] Socks5 Server Start")

	listener, err := net.Listen("tcp", p.saddr)
	if err != nil {
		logger.Error("[Proxy] Socks Error = %v", err)
	}

	for {
		client, err := listener.Accept()
		if err != nil {
			pxylogger.Printf("[Socks] Error = %v", err)
			continue
		}

		go func(client net.Conn) {
			defer client.Close()

			proxy, err := net.Dial("tcp", p.haddr)
			if err != nil {
				pxylogger.Printf("[Socks] Error = %v", err)
				return
			}
			defer proxy.Close()

			err = Socks5Auth(client)
			if err != nil {
				pxylogger.Printf("[Socks] Error = %v", err)
				return
			}

			dest, err := Socks5Connect(client)
			if err != nil {
				pxylogger.Printf("[Socks] Error = %v", err)
				return
			}

			if dest.port == 443 {
				client.Close()
				pxylogger.Print("[Socks] Error = Unsupport HTTPS")
				return
			}

			pxylogger.Printf("[Socks] %s => %s", client.RemoteAddr(), dest)

			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				defer wg.Done()
				request_reader := bufio.NewReader(client)

				for {
					request, err := http.ReadRequest(request_reader)
					if err != nil {
						pxylogger.Printf("[Socks] Error = %v", err)
						return
					}

					request.URL.Path = dest.URL() + request.URL.Path
					//request.URL.Path = dest.Scheme() + "://" + request.Host + request.URL.Path
					request.Header.Set("Proxy-Connection", "Keep-Alive")

					err = request.WriteProxy(proxy)
					if err != nil {
						pxylogger.Printf("[Socks] Error = %v", err)
						return
					}
				}
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := io.Copy(client, proxy)
				if err != nil {
					pxylogger.Printf("[Socks] Error = %v", err)
					return
				}
			}()

			wg.Wait()

		}(client)
	}
}
