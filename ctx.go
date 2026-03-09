package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/GoYoko/web/locale"
)

var logger = slog.Default()

type Context struct {
	echo.Context
	err    error
	locale *locale.Localizer
	page   *Pagination
	es     []ErrHandle
}

type Option func(*Context) error

func WithPage() Option {
	return func(ctx *Context) error {
		p := ctx.QueryParam("page")
		size := ctx.QueryParam("size")
		nt := ctx.QueryParam("next_token")
		i, _ := strconv.Atoi(p)
		pgi, _ := strconv.Atoi(size)
		page := &Pagination{
			Page:      i,
			Size:      pgi,
			NextToken: nt,
		}
		if page.NextToken == "" && page.Page == 0 {
			page.Page = 1
		}
		if page.Size == 0 {
			page.Size = 10
		}
		ctx.page = page
		return nil
	}
}

func (c *Context) Page() *Pagination {
	return c.page
}

func (c *Context) ErrMsg(id ErrorID, param map[string]any) string {
	return c.locale.Message(c.Request().Header.Get("Accept-Language"), string(id), param)
}

func (c *Context) Failed(status, code int, id ErrorID, err error) error {
	for _, handle := range c.es {
		if e := handle(err); e != nil {
			return c.failed(status, e.code, e.id, err, e.param, nil)
		}
	}

	if e, ok := err.(*Err); ok {
		if e.err == nil {
			e.err = errors.New(c.ErrMsg(e.id, e.param))
		}
		return c.failed(e.Status, e.code, e.id, e.err, e.param, e.data)
	}
	return c.failed(status, code, id, err, nil, nil)
}

func (c *Context) failed(status, code int, id ErrorID, err error, param map[string]any, data any) error {
	traceID := GetTraceID(c.Request().Context())

	e := c.ErrMsg(id, param)
	logger.With("trace_id", traceID).With("err", err).Warn("request failed")
	if err != nil {
		c.Set("err-msg", err.Error())
	}
	r := Resp{
		Code:    code,
		Message: fmt.Sprintf("%s [trace_id: %s]", e, traceID),
	}
	if data != nil {
		r.Data = data
	}
	return c.JSON(status, r)
}

func (c *Context) Success(r any) error {
	return c.JSON(http.StatusOK, Resp{
		Code:    0,
		Message: "success",
		Data:    r,
	})
}
