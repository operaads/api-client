package request

import (
	"github.com/operaads/api-client/interceptor"
	"io"
	"time"
)

type APIRequest struct {
	Method string
	URL    string
	Body   io.Reader

	RequestTimeout time.Duration

	URLInterceptors     []interceptor.URLInterceptor
	RequestInterceptors []interceptor.RequestInterceptor
}

type Option func(*APIRequest)

func WithURLInterceptors(intcps ...interceptor.URLInterceptor) Option {
	return func(r *APIRequest) {
		r.URLInterceptors = intcps
	}
}

func AppendURLInterceptors(intcps ...interceptor.URLInterceptor) Option {
	return func(r *APIRequest) {
		if len(intcps) <= 0 {
			return
		}

		if len(r.URLInterceptors) > 0 {
			r.URLInterceptors = append(r.URLInterceptors, intcps...)
		} else {
			r.URLInterceptors = intcps
		}
	}
}

func WithRequestInterceptors(intcps ...interceptor.RequestInterceptor) Option {
	return func(r *APIRequest) {
		r.RequestInterceptors = intcps
	}
}

func AppendRequestInterceptors(intcps ...interceptor.RequestInterceptor) Option {
	return func(r *APIRequest) {
		if len(intcps) <= 0 {
			return
		}

		if len(r.RequestInterceptors) > 0 {
			r.RequestInterceptors = append(r.RequestInterceptors, intcps...)
		} else {
			r.RequestInterceptors = intcps
		}
	}
}

func WithRequestTimeout(timeout time.Duration) Option {
	return func(r *APIRequest) {
		r.RequestTimeout = timeout
	}
}

func NewAPIRequest(method, url string, body io.Reader, opts ...Option) *APIRequest {
	r := &APIRequest{
		Method: method,
		URL:    url,
		Body:   body,
	}

	for _, o := range opts {
		o(r)
	}

	return r
}
