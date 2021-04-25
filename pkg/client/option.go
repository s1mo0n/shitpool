package client

type Option struct {
	Proxy    string `json:"proxy" yaml:"proxy"`
	Timeout  int    `json:"timeout" yaml:"timeout"`
	Redirect bool   `json:"redirect" yaml:"redirect"`

	Cookies     map[string]string `json:"cookies" yaml:"cookies"`
	Headers     map[string]string `json:"headers" yaml:"headers"`
	DenyHeaders []string          `json:"denyheaders" yaml:"denyheaders"`
}

func NewOption() *Option {
	opt := new(Option)

	opt.Headers = make(map[string]string)
	opt.Cookies = make(map[string]string)

	return opt
}
