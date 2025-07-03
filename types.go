package web

import "net/http"

type ErrorID string

const (
	ErrInternal   ErrorID = "err-internal"
	ErrBindParams ErrorID = "err-bind-params"
)

type Pagination struct {
	Page      int    `json:"page" query:"page" validate:"min=1" default:"1"`          // 分页
	Size      int    `json:"size" query:"size" validate:"min=1,max=100" default:"10"` // 每页多少条记录
	NextToken string `json:"next_token" query:"next_token"`                           // 下一页标识
}

type PageInfo struct {
	NextToken   string `json:"next_token,omitempty"`
	HasNextPage bool   `json:"has_next_page"`
	TotalCount  int64  `json:"total_count"`
}

type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type Err struct {
	code int
	err  error
	id   ErrorID
	data map[string]any
}

func (e *Err) Error() string {
	return e.err.Error()
}

func (e *Err) Wrap(err error) *Err {
	e.err = err
	return e
}

func (e *Err) WithData(kv ...any) *Err {
	if len(kv)%2 != 0 {
		panic("[WEB] WithData kv must be even")
	}

	m := make(map[string]any)
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	e.data = m
	return e
}

func NewErr(code int, id ErrorID) *Err {
	return &Err{
		code: code,
		id:   id,
	}
}

func NewBadRequestErr(id ErrorID) *Err {
	return &Err{
		code: http.StatusBadRequest,
		id:   id,
	}
}
