package api_client

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/operaads/api-client/interceptor"
	"github.com/operaads/api-client/request"
	"github.com/operaads/api-client/response"
	"golang.org/x/oauth2/jwt"
)

type Client struct {
	*http.Client

	APIBaseURL     *url.URL
	RequestTimeout time.Duration

	URLInterceptor     interceptor.URLInterceptor
	RequestInterceptor interceptor.RequestInterceptor
}

func NewJWTClient(jwtConfig *jwt.Config, baseURL string, opts ...Option) *Client {
	if jwtConfig == nil {
		panic("jwtConfig is nil")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	opt := &Options{
		RequestTimeout: 10 * time.Second,
	}

	for _, o := range opts {
		o(opt)
	}

	return &Client{
		Client:             jwtConfig.Client(context.Background()),
		APIBaseURL:         u,
		RequestTimeout:     opt.RequestTimeout,
		URLInterceptor:     opt.URLInterceptor,
		RequestInterceptor: opt.RequestInterceptor,
	}
}

func (c *Client) DoAPIRequest(req *request.APIRequest) (*response.APIResponse, error) {
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, err
	}

	var fullURL *url.URL
	if u.Scheme != "" {
		fullURL = u
	} else {
		fullURL = &url.URL{
			Scheme:   c.APIBaseURL.Scheme,
			Host:     c.APIBaseURL.Host,
			Path:     path.Join(c.APIBaseURL.Path, u.Path),
			RawQuery: u.RawQuery,
			Fragment: u.Fragment,
		}
	}

	if c.URLInterceptor != nil {
		c.URLInterceptor(fullURL)
	}
	for _, intcp := range req.URLInterceptors {
		intcp(fullURL)
	}

	requestTimeout := c.RequestTimeout
	if req.RequestTimeout > 0 {
		requestTimeout = req.RequestTimeout
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(timeoutCtx, req.Method, fullURL.String(), req.Body)
	if err != nil {
		return nil, err
	}

	if c.RequestInterceptor != nil {
		c.RequestInterceptor(httpReq)
	}
	for _, intcp := range req.RequestInterceptors {
		intcp(httpReq)
	}

	res, err := c.Do(httpReq)
	if err != nil {
		return nil, err
	}

	return &response.APIResponse{Response: res}, nil
}
