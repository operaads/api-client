package interceptor

import (
	"mime/multipart"
	"net/http"
	"net/url"
)

type URLInterceptor func(*url.URL)

type RequestInterceptor func(*http.Request)
type ResponseInterceptor func(response *http.Response)

type ErrorInterceptor func(error) error

type JSONInterceptor func(map[string]interface{})
type FormInterceptor func(url.Values)
type MultipartFormInterceptor func(*multipart.Writer)
