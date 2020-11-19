package api_client

import (
	"bytes"
	"encoding/json"
	"github.com/operaads/api-client/proxy"
	"github.com/operaads/api-client/request"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type ProxyRequestType string

const (
	ProxyRequestTypeNone          = ProxyRequestType("")
	ProxyRequestTypeRaw           = ProxyRequestType("RAW")
	ProxyRequestTypeForm          = ProxyRequestType("FORM")
	ProxyRequestTypeMultipartForm = ProxyRequestType("MULTIPART_FORM")
)

func (c *Client) ProxyAPI(
	method, path string,
	httpReq *http.Request,
	writer http.ResponseWriter,
	requestType ProxyRequestType,
	opts ...proxy.Option,
) error {
	if path == "" {
		u := &url.URL{
			Path:     httpReq.URL.Path,
			RawQuery: httpReq.URL.RawQuery,
			Fragment: httpReq.URL.Fragment,
		}
		path = u.String()
	}

	// if method is empty, set to http's request method
	if method == "" {
		method = httpReq.Method
	}

	opt := &proxy.Options{
		RequestTimeout: c.Timeout,
	}

	for _, o := range opts {
		o(opt)
	}

	var reqParseFunc func(*http.Request, *proxy.Options) (io.Reader, string, error)
	switch requestType {
	case ProxyRequestTypeRaw:
		reqParseFunc = parseRawRequest
	case ProxyRequestTypeForm:
		reqParseFunc = parseFormRequest
	case ProxyRequestTypeMultipartForm:
		reqParseFunc = parseMultipartFormRequest
	default:
		reqParseFunc = func(req *http.Request, opt *proxy.Options) (io.Reader, string, error) {
			return nil, req.Header.Get("Content-Type"), nil
		}
	}

	reqBody, reqContentType, err := reqParseFunc(httpReq, opt)
	if err != nil {
		return err
	}

	requestOptions := []request.Option{
		request.WithRequestInterceptors(func(r *http.Request) {
			copyHeaders(httpReq.Header, r.Header)

			if reqContentType != "" {
				r.Header.Set("Content-Type", reqContentType)
			}
		}),
		request.WithRequestTimeout(opt.RequestTimeout),
	}

	if opt.URLInterceptor != nil {
		requestOptions = append(
			requestOptions,
			request.AppendURLInterceptors(opt.URLInterceptor),
		)
	}
	if opt.RequestInterceptor != nil {
		requestOptions = append(
			requestOptions,
			request.AppendRequestInterceptors(opt.RequestInterceptor),
		)
	}

	apiReq := request.NewAPIRequest(
		method, path, reqBody,
		requestOptions...,
	)

	res, err := c.DoAPIRequest(apiReq)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	var resBody io.Reader = res.Body

	if opt.ResponseJSONInterceptor != nil {
		var m map[string]interface{}

		if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
			opt.ResponseJSONInterceptor(m)
			return err
		}

		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(m); err != nil {
			return err
		}
		resBody = buf
	}

	// write status code
	writer.WriteHeader(res.StatusCode)
	copyHeaders(res.Header, writer.Header())

	// copy response
	_, err = io.Copy(writer, resBody)
	return err
}

func (c *Client) ProxyJSONAPI(
	method, path string,
	httpReq *http.Request,
	writer http.ResponseWriter,
	opts ...proxy.Option,
) error {
	return c.ProxyAPI(method, path, httpReq, writer, ProxyRequestTypeRaw, opts...)
}

func (c *Client) TransparentProxyJSONAPI(httpReq *http.Request, writer http.ResponseWriter) error {
	return c.ProxyJSONAPI("", "", httpReq, writer)
}

func (c *Client) ProxyFormAPI(
	method, path string,
	httpReq *http.Request,
	writer http.ResponseWriter,
	opts ...proxy.Option,
) error {
	return c.ProxyAPI(method, path, httpReq, writer, ProxyRequestTypeForm, opts...)
}

func (c *Client) TransparentProxyFormAPI(httpReq *http.Request, writer http.ResponseWriter) error {
	return c.ProxyAPI("", "", httpReq, writer, ProxyRequestTypeForm)
}

func (c *Client) ProxyMultipartFormAPI(
	method, path string,
	httpReq *http.Request,
	writer http.ResponseWriter,
	opts ...proxy.Option,
) error {
	return c.ProxyAPI(method, path, httpReq, writer, ProxyRequestTypeMultipartForm, opts...)
}

func (c *Client) TransparentProxyMultipartFormAPI(httpReq *http.Request, writer http.ResponseWriter) error {
	return c.ProxyAPI("", "", httpReq, writer, ProxyRequestTypeMultipartForm)
}

func copyHeaders(src, dst http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func parseRawRequest(req *http.Request, opt *proxy.Options) (io.Reader, string, error) {
	if opt.RequestJSONInterceptor != nil {
		defer req.Body.Close()

		var m map[string]interface{}

		if err := json.NewDecoder(req.Body).Decode(&m); err != nil {
			return nil, "", err
		}

		opt.RequestJSONInterceptor(m)

		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(m); err != nil {
			return nil, "", err
		}

		return buf, "", nil
	}

	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return req.Body, contentType, nil
}

func parseFormRequest(req *http.Request, opt *proxy.Options) (io.Reader, string, error) {
	if err := req.ParseForm(); err != nil {
		return nil, "", err
	}

	form := url.Values{}
	for k, vv := range req.PostForm {
		for _, v := range vv {
			form.Add(k, v)
		}
	}

	if opt.RequestFormInterceptor != nil {
		opt.RequestFormInterceptor(form)
	}

	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/x-www-form-urlencoded"
	}

	return strings.NewReader(form.Encode()), contentType, nil
}

func parseMultipartFormRequest(req *http.Request, opt *proxy.Options) (io.Reader, string, error) {
	if err := req.ParseMultipartForm(opt.MaxUploadSize); err != nil {
		return nil, "", err
	}

	reqBody := new(bytes.Buffer)
	multiWriter := multipart.NewWriter(reqBody)

	defer multiWriter.Close()

	for k, vv := range req.MultipartForm.Value {
		for _, v := range vv {
			if err := multiWriter.WriteField(k, v); err != nil {
				return nil, "", err
			}
		}
	}

	for k, vv := range req.MultipartForm.File {
		for _, v := range vv {
			f, err := v.Open()
			if err != nil {
				return nil, "", err
			}
			writer, err := multiWriter.CreateFormFile(k, v.Filename)
			if err != nil {
				return nil, "", err
			}
			if _, err := io.Copy(writer, f); err != nil {
				return nil, "", err
			}

			f.Close()
		}
	}

	if opt.RequestMultipartFormInterceptor != nil {
		opt.RequestMultipartFormInterceptor(multiWriter)
	}

	return reqBody, multiWriter.FormDataContentType(), nil
}
