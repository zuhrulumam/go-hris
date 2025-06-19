package pkg

import (
	"time"
)

type CtxVal string

const (
	HeaderAuthorization         = "Authorization"
	HeaderContentType           = "Content-Type"
	HeaderContentLength         = "Content-Length"
	HeaderXForwardedFor         = "X-Forwarded-For"
	HeaderXForwardedHost        = "X-Forwarded-Host"
	HeaderXClientID             = "X-Client-Id"
	HeaderXClientVersion        = "X-Client-Version"
	HeaderXRequestID            = "X-Request-Id"
	HeaderContentTypeJSONPrefix = "application/json"
)

var (
	TxCtxValue CtxVal = "tx"
)

func BoolPtr(b bool) *bool {
	return &b
}

func TimePtr(b time.Time) *time.Time {
	return &b
}

func StringPtr(b string) *string {
	return &b
}

func UintPtr(b uint) *uint {
	return &b
}
