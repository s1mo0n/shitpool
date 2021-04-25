package shit

import (
	"fmt"
	"math/rand"
	"time"
)

var (
	Types = []string{
		"all",
		"fuck",
		"http",
		"https",
		"socks4",
		"socks5",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Shit struct {
	Type     string `json:"type"`
	Speed    int64  `json:"speed"`
	Fuckwall bool   `json:"fuckwall"`
	Source   string `json:"source"`
	Country  string `json:"country"`
}

func NewShit() Shit {
	return Shit{}
}

type ShitMap map[string]*Shit

func NewShitMap() ShitMap {
	return make(ShitMap)
}

func (s ShitMap) AddShit(shitName string, shit *Shit) {
	s[shitName] = shit
}

func (s ShitMap) DelShit(shitName string) {
	delete(s, shitName)
}

func (s ShitMap) Count() int {
	return len(s)
}

func (s ShitMap) Random() string {
	if len(s) == 0 {
		return ""
	}

	r := rand.Intn(len(s))
	for k := range s {
		if r == 0 {
			return k
		}
		r--
	}
	return ""
}

func (s ShitMap) ToList() []string {
	shitList := make([]string, len(s))

	i := 0
	for k := range s {
		shitList[i] = k
		i++
	}

	return shitList
}

type Collector struct {
	all    ShitMap
	fuck   ShitMap
	http   ShitMap
	https  ShitMap
	socks4 ShitMap
	socks5 ShitMap
}

func NewCollector() *Collector {
	c := new(Collector)

	c.all = NewShitMap()
	c.fuck = NewShitMap()
	c.http = NewShitMap()
	c.https = NewShitMap()
	c.socks4 = NewShitMap()
	c.socks5 = NewShitMap()

	return c
}

func (c *Collector) Len(typ string) int {
	switch typ {
	case "all":
		return len(c.all)
	case "fuck":
		return len(c.fuck)
	case "http":
		return len(c.http)
	case "https":
		return len(c.https)
	case "socks4":
		return len(c.socks4)
	case "socks5":
		return len(c.socks5)
	default:
		return 0
	}
}

func (c *Collector) Add(shitName string, shit *Shit) error {
	switch shit.Type {
	case "http":
		c.http[shitName] = shit
	case "https":
		c.https[shitName] = shit
	case "socks4":
		c.socks4[shitName] = shit
	case "socks5":
		c.socks5[shitName] = shit
	default:
		return fmt.Errorf("Your shit is too shit, collector doesn't want it.")
	}

	c.all[shitName] = shit
	if shit.Fuckwall {
		c.fuck[shitName] = shit
	}

	return nil
}

func (c *Collector) Del(shitName string) {
	delete(c.all, shitName)
	delete(c.fuck, shitName)
	delete(c.http, shitName)
	delete(c.https, shitName)
	delete(c.socks4, shitName)
	delete(c.socks5, shitName)
}

func (c *Collector) Get(typ string) ShitMap {
	switch typ {
	case "all":
		return c.all
	case "fuck":
		return c.fuck
	case "http":
		return c.http
	case "https":
		return c.https
	case "socks4":
		return c.socks4
	case "socks5":
		return c.socks5
	default:
		return nil
	}
}
