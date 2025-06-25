# web
web 是基于 echo 的 web 框架，提供了一些好用的功能。

- 参数自动绑定和校验
- 错误处理
- 国际化信息提示
- 好看的 Swagger UI

## Getting started

1. 安装依赖
```bash
go get -u github.com/GoYoko/web@latest
```

2. 初始化 web 框架
```go
import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/GoYoko/web"
)

func main() {
	w := web.New()
	if err := w.Run(":8080"); err != nil {
		panic(err)
	}
}
```

3. 无参数的 handler
```go
w.GET("/", web.BaseHandler(Hello))

func Hello(ctx *web.Context) error {
    if err := fn(); err != nil {
        return err
    }

    return ctx.Success("hello, world")
}
```

4. 带参数的 handler
```go
type HelloReq struct {
    ID string `json:"id" form:"id" query:"id"`
}

w.GET("/", web.BindHandler(Hello))

func Hello(ctx *web.Context, req HelloReq) error {
    if err := fn(req.ID); err != nil {
        return err
    }

    return ctx.Success("hello, world")
}
```

5. 配置Swagger
> swaggerJSON 是 swagger.json 文件的内容

```go
w.Swagger("API Title", "/reference", string(swaggerJSON))
```

## 自定义错误
1. 定义错误 i18n 文件。locale.zh.toml
如果需要其它语言，可以定义 locale.en.toml 等。可以参考 locale/default.en.toml 文件。
```
[err-permission]
other = "没有权限"

[err-record-not-found]
other = "记录不存在"
```

2. 注册错误id
```go
import (
    "github.com/GoYoko/web/locale"
)

//go:embed locale.*.toml
var LocaleFS embed.FS

l := locale.NewLocalizerWithFile(language.Chinese, LocaleFS, []string{"locale.zh.toml"})
w.SetLocale(l)
```

3. 定义错误, 这里定义的错误 id 需要和 i18n 文件中的 id 一致
```go
ErrPermission = web.NewErr(http.StatusForbidden, "err-permission")
ErrNotFound   = web.NewErr(http.StatusNotFound, "err-record-not-found")
```

4. 使用错误, 直接返回对应错误。当然可以将上层错误包装方便在日志中查看
```go
func Hello(ctx *web.Context) error {
    if err := fn(); err != nil {
        return ErrNotFound.Wrap(err)
    }

    return ctx.Success("hello, world")
}
```

5. 动态错误信息  

i18n 文件中可以定义模板
```
[err-record-not-found]
other = "记录不存在: {{.name}}"
```

返回错误时，可以传入动态数据
```go
func Hello(ctx *web.Context) error {
    if err := fn(); err != nil {
        return ErrNotFound.WithData("name", "test")
    }

    return ctx.Success("hello, world")
}
```
