package interceptor

import (
	"mime/multipart"
	"net/http"
	"net/url"
)

type URLInterceptor func(*url.URL)

type RequestInterceptor func(*http.Request)

type ResponseInterceptor func(response *http.Response)

type JSONInterceptor func(interface{}) (interface{}, error)

type FormInterceptor func(url.Values) (url.Values, error)

type MultipartFormInterceptor func(*multipart.Writer) error
