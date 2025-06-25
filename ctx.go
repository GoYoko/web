package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rs/xid"

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

func (c *Context) ErrMsg(id ErrorID, data map[string]any) string {
	return c.locale.Message(c.Request().Header.Get("Accept-Language"), string(id), data)
}

func (c *Context) Failed(code int, id ErrorID, err error) error {
	for _, handle := range c.es {
		if e := handle(err); e != nil {
			return c.failed(e.code, e.id, err, e.data)
		}
	}

	if e, ok := err.(*Err); ok {
		if e.err == nil {
			e.err = errors.New(c.ErrMsg(e.id, e.data))
		}
		return c.failed(e.code, e.id, e.err, e.data)
	}
	return c.failed(code, id, err, nil)
}

func (c *Context) failed(code int, id ErrorID, err error, data map[string]any) error {
	traceID := xid.New().String()
	e := c.ErrMsg(id, data)
	logger.With("trace_id", traceID).With("err", err).Warn("request failed")
	c.Set("err-msg", err.Error())
	return c.JSON(code, Resp{
		Code:    code,
		Message: fmt.Sprintf("%s [trace_id: %s]", e, traceID),
	})
}

func (c *Context) Success(r any) error {
	return c.JSON(http.StatusOK, Resp{
		Code:    0,
		Message: "success",
		Data:    r,
	})
}
