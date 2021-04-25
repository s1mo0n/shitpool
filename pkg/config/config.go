package config

import (
	"encoding/json"
	"os"

	"github.com/s1mo0n/shitpool/pkg/logger"
)

var Global = &Config{
	Worker:   100,
	ChkDelay: 10,
	DataFile: "shitpool.json",
	LogLevel: "INFO",
}

type Config struct {
	Worker   int    `json:"worker"`
	ChkDelay int    `json:"chkdelay"`
	DataFile string `json:"datafile"`
	LogLevel string `json:"loglevel"`
}

func ParseConfig(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		logger.Error("[ParseConfig] File = %s Error = %v", filename, err)
	}
	de := json.NewDecoder(f)
	err = de.Decode(Global)
	if err != nil {
		logger.Error("[ParseConfig] File = %s Error = %v", filename, err)
	}
}
