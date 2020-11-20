package proxy

import (
	"time"

	"github.com/operaads/api-client/interceptor"
)

type Options struct {
	MaxUploadSize  int64
	RequestTimeout time.Duration

	URLInterceptor     interceptor.URLInterceptor
	RequestInterceptor interceptor.RequestInterceptor

	RequestJSONInterceptor          interceptor.JSONInterceptor
	RequestFormInterceptor          interceptor.FormInterceptor
	RequestMultipartFormInterceptor interceptor.MultipartFormInterceptor

	ResponseJSONInterceptor interceptor.JSONInterceptor
	TransferResponseHeaders []string
}

type Option func(*Options)

func WithMaxUploadSize(size int64) Option {
	return func(o *Options) {
		o.MaxUploadSize = size
	}
}

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

func WithRequestJSONInterceptor(intcp interceptor.JSONInterceptor) Option {
	return func(o *Options) {
		o.RequestJSONInterceptor = intcp
	}
}

func WithRequestFormInterceptor(intcp interceptor.FormInterceptor) Option {
	return func(o *Options) {
		o.RequestFormInterceptor = intcp
	}
}

func WithRequestMultipartFormInterceptor(intcp interceptor.MultipartFormInterceptor) Option {
	return func(o *Options) {
		o.RequestMultipartFormInterceptor = intcp
	}
}

func WithResponseJSONInterceptor(intcp interceptor.JSONInterceptor) Option {
	return func(o *Options) {
		o.ResponseJSONInterceptor = intcp
	}
}

func WithTransferResponseHeaders(headers ...string) Option {
	return func(o *Options) {
		o.TransferResponseHeaders = make([]string, len(headers))

		copy(o.TransferResponseHeaders, headers)
	}
}

func AppendTransferResponseHeaders(headers ...string) Option {
	return func(o *Options) {
		if o.TransferResponseHeaders == nil {
			o.TransferResponseHeaders = make([]string, len(headers))

			copy(o.TransferResponseHeaders, headers)
		} else {
			o.TransferResponseHeaders = append(o.TransferResponseHeaders, headers...)
		}
	}
}
