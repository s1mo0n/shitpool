package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/s1mo0n/shitpool/pkg/logger"
	"github.com/s1mo0n/shitpool/pkg/shit"
)

var apilogger *log.Logger

type API struct {
	addr string
}

func NewAPI(addr string, logpath string) *API {
	writer, err := os.OpenFile(logpath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		logger.Error("[API] Error = %v", err)
	}
	apilogger = log.New(writer, "[API]", log.LstdFlags)

	return &API{
		addr: addr,
	}
}

func now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (api *API) Run() {
	logger.Info("[API] API Server Start")

	mux := http.NewServeMux()

	mux.HandleFunc("/get", ShitHandler("get", "all"))
	mux.HandleFunc("/get/fuck", ShitHandler("get", "fuck"))
	mux.HandleFunc("/get/http", ShitHandler("get", "http"))
	mux.HandleFunc("/get/https", ShitHandler("get", "https"))
	mux.HandleFunc("/get/socks4", ShitHandler("get", "socks4"))
	mux.HandleFunc("/get/socks5", ShitHandler("get", "socks5"))

	mux.HandleFunc("/get_all", ShitHandler("get_all", "all"))
	mux.HandleFunc("/get_all/fuck", ShitHandler("get_all", "fuck"))
	mux.HandleFunc("/get_all/http", ShitHandler("get_all", "http"))
	mux.HandleFunc("/get_all/https", ShitHandler("get_all", "https"))
	mux.HandleFunc("/get_all/socks4", ShitHandler("get_all", "socks4"))
	mux.HandleFunc("/get_all/socks5", ShitHandler("get_all", "socks5"))

	mux.HandleFunc("/count", ShitHandler("count", ""))

	logger.Error("[API] Error = %v", http.ListenAndServe(api.addr, mux))
}

func ShitHandler(fun, typ string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apilogger.Printf("%v <- %s %s", now(), r.RemoteAddr, r.URL.Path)
		if r.Method == "GET" {
			var value interface{}
			switch fun {
			case "get":
				value = shit.Get(typ)
			case "get_all":
				value = shit.GetAll(typ)
			case "count":
				v := make(map[string]int)
				for _, typ := range shit.Types {
					v[typ] = shit.Len(typ)
				}
				value = v
			}

			data, err := json.Marshal(value)
			if err != nil {
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		}
	}
}
