package response

import "net/http"

type APIResponse struct {
	*http.Response
}

