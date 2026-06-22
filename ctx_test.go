package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/GoYoko/web/locale"
)

func newTestContext(req *http.Request) *Context {
	e := echo.New()
	rec := httptest.NewRecorder()
	return &Context{
		Context: e.NewContext(req, rec),
		locale:  locale.NewLocalizer(),
	}
}

func TestContextLang(t *testing.T) {
	tests := []struct {
		name       string
		headerLang string
		cookieLang string
		want       string
	}{
		{
			name:       "cookie language has priority",
			headerLang: "zh",
			cookieLang: "en",
			want:       "en",
		},
		{
			name:       "fallback to accept language when cookie is missing",
			headerLang: "en",
			want:       "en",
		},
		{
			name:       "maps cn to zh",
			cookieLang: "cn",
			want:       "zh",
		},
		{
			name:       "fallback to accept language when cookie is blank",
			headerLang: "en",
			cookieLang: "  ",
			want:       "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.headerLang != "" {
				req.Header.Set("Accept-Language", tt.headerLang)
			}
			if tt.cookieLang != "" {
				req.AddCookie(&http.Cookie{Name: "language", Value: tt.cookieLang})
			}

			ctx := newTestContext(req)
			if got := ctx.Lang(); got != tt.want {
				t.Fatalf("Lang() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContextErrMsgUsesCookieLanguage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "zh")
	req.AddCookie(&http.Cookie{Name: "language", Value: "en"})

	ctx := newTestContext(req)
	if got := ctx.ErrMsg(ErrBindParams, nil); got != "Invalid parameters" {
		t.Fatalf("ErrMsg() = %q, want %q", got, "Invalid parameters")
	}
}
