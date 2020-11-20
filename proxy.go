package api_client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/operaads/api-client/proxy"
	"github.com/operaads/api-client/request"
)

func (c *Client) ProxyAPI(
	method, path string,
	httpReq *http.Request,
	resWriter http.ResponseWriter,
	reqBodyType proxy.RequestBodyType,
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

	switch reqBodyType {
	case proxy.RequestBodyTypeRaw:
		reqParseFunc = parseRawRequest
	case proxy.RequestBodyTypeForm:
		reqParseFunc = parseFormRequest
	case proxy.RequestBodyTypeMultipartForm:
		reqParseFunc = parseMultipartFormRequest
	default:
		reqParseFunc = func(req *http.Request, opt *proxy.Options) (io.Reader, string, error) {
			return nil, "", nil
		}
	}

	reqBody, reqContentType, err := reqParseFunc(httpReq, opt)
	if err != nil {
		return err
	}

	requestOptions := []request.Option{
		request.WithRequestInterceptors(func(r *http.Request) {
			for k, vv := range httpReq.Header {
				for _, v := range vv {
					r.Header.Add(k, v)
				}
			}

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

	if res.StatusCode == http.StatusNoContent {
		resWriter.WriteHeader(http.StatusNoContent)

		// transfer response headers
		for _, h := range opt.TransferResponseHeaders {
			if vv, ok := res.Header[h]; ok {
				headerValue := make([]string, len(vv))
				copy(headerValue, vv)

				resWriter.Header()[h] = headerValue
			}
		}

		return nil
	}

	defer res.Body.Close()

	resHeaders := make(http.Header)

	// transfer response headers
	for _, h := range opt.TransferResponseHeaders {
		if vv, ok := res.Header[h]; ok {
			headerValue := make([]string, len(vv))
			copy(headerValue, vv)
			resHeaders[h] = headerValue
		}
	}

	var resBody io.Reader

	resContentEncoding := res.Header.Get("Content-Encoding")

	if opt.ResponseJSONInterceptor != nil {
		var obj interface{}

		var reader io.Reader
		switch resContentEncoding {
		case "gzip":
			if gzReader, err := gzip.NewReader(res.Body); err != nil {
				return err
			} else {
				reader = gzReader
			}
		default:
			reader = res.Body
		}

		if err := json.NewDecoder(reader).Decode(&obj); err != nil {
			return err
		}

		if newObj, err := opt.ResponseJSONInterceptor(obj); err != nil {
			return err
		} else {
			obj = newObj
		}

		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(obj); err != nil {
			return err
		}

		resHeaders.Set("Content-Type", "application/json; charset=utf-8")
		resHeaders.Set("Content-Length", strconv.Itoa(buf.Len()))

		resBody = buf
	} else {
		resHeaders.Set("Content-Type", res.Header.Get("Content-Type"))

		if res.ContentLength >= 0 {
			resHeaders.Set("Content-Length", strconv.FormatInt(res.ContentLength, 10))
		}
		if resContentEncoding != "" {
			resHeaders.Set("Content-Encoding", resContentEncoding)
		}

		resBody = res.Body
	}

	for k, vv := range resHeaders {
		resWriter.Header()[k] = vv
	}

	// write status code
	resWriter.WriteHeader(res.StatusCode)

	// copy response
	if _, err := io.Copy(resWriter, resBody); err != nil {
		return err
	}

	return nil
}

func (c *Client) TransparentProxyAPI(httpReq *http.Request, resWriter http.ResponseWriter, requestType proxy.RequestBodyType) error {
	return c.ProxyAPI("", "", httpReq, resWriter, requestType)
}

func (c *Client) ProxyGetAPI(
	path string,
	httpReq *http.Request,
	resWriter http.ResponseWriter,
	opts ...proxy.Option,
) error {
	return c.ProxyAPI("GET", path, httpReq, resWriter, proxy.RequestBodyTypeNone, opts...)
}

func (c *Client) TransparentProxyGetAPI(httpReq *http.Request, resWriter http.ResponseWriter) error {
	return c.ProxyGetAPI("", httpReq, resWriter)
}

func parseRawRequest(req *http.Request, opt *proxy.Options) (io.Reader, string, error) {
	if opt.RequestJSONInterceptor != nil {
		defer req.Body.Close()

		var obj interface{}

		if err := json.NewDecoder(req.Body).Decode(&obj); err != nil {
			return nil, "", err
		}

		if newObj, err := opt.RequestJSONInterceptor(obj); err != nil {
			return nil, "", err
		} else {
			obj = newObj
		}

		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(obj); err != nil {
			return nil, "", err
		}

		return buf, "application/json; charset=utf-8", nil
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
		if newForm, err := opt.RequestFormInterceptor(form); err != nil {
			return nil, "", err
		} else {
			form = newForm
		}
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
		if err := opt.RequestMultipartFormInterceptor(multiWriter); err != nil {
			return nil, "", err
		}
	}

	return reqBody, multiWriter.FormDataContentType(), nil
}
