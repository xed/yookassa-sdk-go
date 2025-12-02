// Package yookassa implements all the necessary methods for working with YooMoney.
package yookassa

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	BaseURL = "https://api.yookassa.ru/v3/"
)

// Client works with YooMoney API.
type Client struct {
	client    http.Client
	accountId string
	secretKey string
}

func NewClient(accountId string, secretKey string) *Client {
	return &Client{
		client:    http.Client{},
		accountId: accountId,
		secretKey: secretKey,
	}
}

func (c *Client) makeRequest(
	method string,
	endpoint string,
	body []byte,
	params map[string]interface{},
	idempotencyKey string,
) (*http.Response, error) {
	uri := fmt.Sprintf("%s%s", BaseURL, endpoint)

	req, err := http.NewRequest(method, uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if idempotencyKey == "" {
		idempotencyKey = uuid.NewString()
	}

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotence-Key", idempotencyKey)
	}

	req.SetBasicAuth(c.accountId, c.secretKey)

	if params != nil {
		q := req.URL.Query()
		for paramName, paramVal := range params {
			q.Add(paramName, fmt.Sprintf("%v", paramVal))
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// SetSocks5Proxy configures the HTTP client to tunnel requests through the provided SOCKS5 proxy.
// Pass an empty string to disable the proxy entirely.
func (c *Client) SetSocks5Proxy(proxyAddr string) error {
	if proxyAddr == "" {
		c.client.Transport = nil
		return nil
	}

	cfg, err := parseSocks5Proxy(proxyAddr)
	if err != nil {
		return err
	}

	dialer := newSocks5Dialer(cfg.address, cfg.username, cfg.password)
	transport := &http.Transport{
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	transport.Proxy = nil
	transport.DialContext = dialer.DialContext

	c.client.Transport = transport
	return nil
}

type socks5ProxyConfig struct {
	address  string
	username string
	password string
}

func parseSocks5Proxy(raw string) (socks5ProxyConfig, error) {
	normalized := raw
	if !strings.Contains(raw, "://") {
		normalized = "socks5://" + raw
	}

	u, err := url.Parse(normalized)
	if err != nil {
		return socks5ProxyConfig{}, err
	}

	if u.Scheme != "" && u.Scheme != "socks5" {
		return socks5ProxyConfig{}, fmt.Errorf("yookassa: unsupported proxy scheme %q", u.Scheme)
	}

	address := u.Host
	if address == "" {
		address = u.Path
	}
	if address == "" {
		return socks5ProxyConfig{}, fmt.Errorf("yookassa: proxy address is empty")
	}
	if !strings.Contains(address, ":") {
		return socks5ProxyConfig{}, fmt.Errorf("yookassa: proxy address must include port")
	}

	config := socks5ProxyConfig{address: address}
	if u.User != nil {
		config.username = u.User.Username()
		config.password, _ = u.User.Password()
	}

	return config, nil
}
