package model

type HttpError struct {
	Message string
	Code    int
}

// 讓 HttpError 符合 Go 的 error 介面
func (err *HttpError) Error() string {
	return err.Message
}
