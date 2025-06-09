package pkg

import (
	"time"
)

type CtxVal string

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
