package locale

import (
	"embed"
	"io/fs"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed default.*.toml
var LocaleFS embed.FS

type Localizer struct {
	bundle *i18n.Bundle
}

func NewLocalizer() *Localizer {
	bundle := i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.LoadMessageFileFS(LocaleFS, "default.zh.toml")
	bundle.LoadMessageFileFS(LocaleFS, "default.en.toml")
	return &Localizer{
		bundle: bundle,
	}
}

func NewLocalizerWithFile(defaultLang language.Tag, fs fs.FS, paths []string) *Localizer {
	bundle := i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.LoadMessageFileFS(LocaleFS, "default.zh.toml")
	bundle.LoadMessageFileFS(LocaleFS, "default.en.toml")
	for _, path := range paths {
		bundle.LoadMessageFileFS(fs, path)
	}
	return &Localizer{
		bundle: bundle,
	}
}

func (e *Localizer) Message(lang, id string, data map[string]any) string {
	loc := i18n.NewLocalizer(e.bundle, lang)
	msg, err := loc.Localize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: data,
	})
	if err != nil {
		return err.Error()
	}
	return msg
}
