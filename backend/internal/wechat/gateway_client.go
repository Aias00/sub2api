package wechat

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GatewayClient struct {
	httpClient *http.Client
}

func NewGatewayClient(httpClient *http.Client) *GatewayClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &GatewayClient{httpClient: httpClient}
}

func (c *GatewayClient) Do(ctx context.Context, method string, path string, query url.Values, form url.Values, cookieHeader string) (*http.Response, error) {
	return c.DoWithReferer(ctx, method, path, query, form, cookieHeader, "https://mp.weixin.qq.com/")
}

func (c *GatewayClient) DoWithReferer(ctx context.Context, method string, path string, query url.Values, form url.Values, cookieHeader string, referer string) (*http.Response, error) {
	if c == nil {
		c = NewGatewayClient(nil)
	}
	endpoint := url.URL{Scheme: "https", Host: "mp.weixin.qq.com", Path: path}
	endpoint.RawQuery = query.Encode()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	if strings.TrimSpace(referer) == "" {
		referer = "https://mp.weixin.qq.com/"
	}
	req.Header.Set("Referer", referer)
	req.Header.Set("Origin", "https://mp.weixin.qq.com")
	req.Header.Set("Accept-Encoding", "identity")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	}
	if strings.TrimSpace(cookieHeader) != "" {
		req.Header.Set("Cookie", cookieHeader)
	}
	return c.httpClient.Do(req)
}
