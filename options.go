package api_client

import (
	"github.com/operaads/api-client/interceptor"
	"time"
)

type Options struct {
	RequestTimeout time.Duration

	URLInterceptor     interceptor.URLInterceptor
	RequestInterceptor interceptor.RequestInterceptor
}

type Option func(*Options)

func WithRequestTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.RequestTimeout = timeout
	}
}

func WithURLInterceptor(intcp interceptor.URLInterceptor) Option {
	return func(o *Options) {
		o.URLInterceptor = intcp
	}
}

func WithRequestInterceptor(intcp interceptor.RequestInterceptor) Option {
	return func(o *Options) {
		o.RequestInterceptor = intcp
	}
}
