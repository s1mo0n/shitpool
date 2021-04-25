package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	ctx context.Context
	cli *http.Client

	cookies     map[string]string
	headers     map[string]string
	denyHeaders []string
}

func NewClient(option *Option, ctx ...context.Context) (*Client, error) {
	if option == nil {
		option = NewOption()
	}
	proxy := option.Proxy
	timeout := option.Timeout
	redirect := option.Redirect

	proxyFunc := http.ProxyFromEnvironment
	if len(proxy) > 0 {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		proxyFunc = http.ProxyURL(proxyURL)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			Proxy:               proxyFunc,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxIdleConnsPerHost: 50,
		},
	}

	if !redirect {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	}

	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	}

	var c context.Context
	if len(ctx) > 0 {
		c = ctx[0]
	} else {
		c = context.TODO()
	}

	return &Client{
		ctx:         c,
		cli:         client,
		cookies:     option.Cookies,
		headers:     option.Headers,
		denyHeaders: option.DenyHeaders,
	}, nil
}

func (client *Client) SetHeader(k, v string) {
	client.headers[k] = v
}

func (client *Client) SetCookie(k, v string) {
	client.cookies[k] = v
}

func (client *Client) Do(req *http.Request) (*http.Response, error) {
	for name, value := range client.cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	for name, value := range client.headers {
		if strings.ToLower(name) == "host" {
			req.Host = value
		} else {
			req.Header.Add(name, value)
		}
	}

	for _, name := range client.denyHeaders {
		if strings.ToLower(name) == "user-agent" {
			req.Header.Set(name, "")
		} else {
			req.Header.Del(name)
		}
	}

	req = req.WithContext(client.ctx)

	return client.cli.Do(req)
}

func (client *Client) Get(target string) (*http.Response, error) {
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func (client *Client) Head(target string) (*http.Response, error) {
	req, err := http.NewRequest("HEAD", target, nil)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func (client *Client) Post(target, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", target, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return client.Do(req)
}

func (client *Client) PostForm(target string, data url.Values) (*http.Response, error) {
	return client.Post(target, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

type File struct {
	Name    string            `json:"name" yaml:"name"`
	Content interface{}       `json:"content" yaml:"content"`
	Headers map[string]string `json:"headers" yaml:"headers"`
}

func (client *Client) PostFile(target string, files map[string]File) (*http.Response, error) {
	var err error
	buf := new(bytes.Buffer)

	writer := multipart.NewWriter(buf)

	for field, file := range files {
		if len(file.Name) == 0 {
			content, ok := file.Content.(string)
			if !ok {
				return nil, fmt.Errorf("Client: File Content Error.")
			}
			err = writer.WriteField(field, content)
			if err != nil {
				return nil, err
			}
		} else {
			header := make(textproto.MIMEHeader)
			header.Set("Content-Disposition", fmt.Sprintf("form-data; name=\"%s\"; filename=\"%s\"", field, file.Name))
			for name, value := range file.Headers {
				header.Set(name, value)
			}
			partWriter, err := writer.CreatePart(header)
			if err != nil {
				return nil, err
			}

			switch file.Content.(type) {
			case string:
				partWriter.Write([]byte(file.Content.(string)))
			case []byte:
				partWriter.Write(file.Content.([]byte))
			case io.Reader:
				io.Copy(partWriter, file.Content.(io.Reader))
			default:
				return nil, fmt.Errorf("Uncontent file content type")
			}

		}
	}

	writer.Close()

	return client.Post(target, writer.FormDataContentType(), buf)
}
