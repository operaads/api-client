package proxy

type RequestBodyType string

const (
	RequestBodyTypeNone          = RequestBodyType("")
	RequestBodyTypeRaw           = RequestBodyType("RAW")
	RequestBodyTypeForm          = RequestBodyType("FORM")
	RequestBodyTypeMultipartForm = RequestBodyType("MULTIPART_FORM")
)
