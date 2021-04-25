package shit

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/s1mo0n/shitpool/pkg/logger"
)

var (
	lock     sync.RWMutex
	collects = NewCollector()
)

type IP struct {
	Shit
	Data string `json:"data"`
}

func NewIP() *IP {
	ip := new(IP)
	ip.Shit = NewShit()

	return ip
}

func (ip IP) String() string {
	if len(ip.Type) == 0 || len(ip.Data) == 0 {
		return ""
	}
	return ip.Type + "://" + ip.Data
}

func Parse(filename string) error {
	var ips []*IP

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&ips); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	for _, ip := range ips {
		err = Add(ip)
		if err != nil {
			logger.Warn("[Parse] Add IP = %v Error = %v", *ip, err)
		}
	}

	return nil
}

func Len(typ string) int {
	lock.RLock()
	defer lock.RUnlock()

	return collects.Len(typ)
}

func Add(ip *IP) error {
	lock.Lock()
	defer lock.Unlock()

	err := collects.Add(ip.Data, &ip.Shit)
	if err != nil {
		return err
	}

	return nil
}

func Del(ip string) {
	lock.Lock()
	defer lock.Unlock()

	collects.Del(ip)
}

func Get(typ string) *IP {
	lock.RLock()
	defer lock.RUnlock()

	ip := NewIP()
	shitMap := collects.Get(typ)

	ip.Data = shitMap.Random()
	if len(ip.Data) != 0 {
		ip.Shit = *shitMap[ip.Data]
	}

	return ip
}

func GetAll(typ string) []*IP {
	lock.RLock()
	defer lock.RUnlock()

	shitMap := collects.Get(typ)

	ips := make([]*IP, shitMap.Count())
	i := 0
	for k, v := range shitMap {
		ip := NewIP()
		ip.Data = k
		ip.Shit = *v
		ips[i] = ip
		i++
	}

	return ips
}
