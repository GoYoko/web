package web

import "net/http"

type ErrorID string

const (
	ErrInternal   ErrorID = "err-internal"
	ErrBindParams ErrorID = "err-bind-params"
)

type Pagination struct {
	Page      int    `json:"page" query:"page"`             // 分页
	Size      int    `json:"size" query:"size"`             // 每页多少条记录
	NextToken string `json:"next_token" query:"next_token"` // 下一页标识
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
	Status int            // 对应 http status code
	code   int            // 业务内部的错误码, 由业务方定义
	err    error          // 原始错误
	id     ErrorID        // 用于解析 i18n 的 id
	param  map[string]any // 模板参数
	data   any            // 一些错误也需要返回 data 数据
}

func (e *Err) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *Err) Wrap(err error) *Err {
	// 如果传入的就是 *Err,直接返回当前 Err,不再包装
	if _, ok := err.(*Err); ok {
		return e
	}
	e.err = err
	return e
}

func (e *Err) WithData(data any) *Err {
	e.data = data
	return e
}

func (e *Err) WithParam(kv ...any) *Err {
	if len(kv)%2 != 0 {
		panic("[WEB] WithParam kv must be even")
	}

	m := make(map[string]any)
	for i := 0; i < len(kv); i += 2 {
		k, ok := kv[i].(string)
		if !ok {
			panic("[WEB] WithParam key must be string")
		}
		m[k] = kv[i+1]
	}
	e.param = m
	return e
}

func NewErr(status, code int, id ErrorID) *Err {
	return &Err{
		Status: status,
		code:   code,
		id:     id,
	}
}

func NewBadRequestErr(id ErrorID) *Err {
	return &Err{
		Status: http.StatusBadRequest,
		code:   http.StatusBadRequest,
		id:     id,
	}
}
