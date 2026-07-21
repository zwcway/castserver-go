package i18n

import (
	"embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const defaultLocaleDir = "locales"

var (
	currentLanguage  = "en"
	currentBundle    *i18n.Bundle
	currentLocalizer *i18n.Localizer
)

//go:embed locales/*
var localeFS embed.FS

func init() {
	currentBundle = newBundle()
	if err := Load(defaultLocaleDir); err != nil {
		currentBundle = newBundle()
		_ = loadEmbeddedLocaleFiles()
		currentLocalizer = i18n.NewLocalizer(currentBundle, currentLanguage)
	}
}

func newBundle() *i18n.Bundle {
	b := i18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	return b
}

func Load(localeDir string) error {
	currentBundle = newBundle()
	if localeDir == "" {
		localeDir = defaultLocaleDir
	}

	if !filepath.IsAbs(localeDir) {
		cwd, err := os.Getwd()
		if err == nil {
			localeDir = filepath.Join(cwd, localeDir)
		}
	}

	if localeDir != "" {
		if info, err := os.Stat(localeDir); err == nil && info.IsDir() {
			if err := loadLocaleFilesFromDir(localeDir); err != nil {
				return err
			}
			currentLocalizer = i18n.NewLocalizer(currentBundle, currentLanguage)
			return nil
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	if err := loadEmbeddedLocaleFiles(); err != nil {
		return err
	}

	currentLocalizer = i18n.NewLocalizer(currentBundle, currentLanguage)
	return nil
}

func SetLanguage(lang string) {
	if lang == "" {
		lang = "en"
	}

	switch strings.ToLower(lang) {
	case "zh", "zh-cn", "zh-hans", "zh-hant":
		currentLanguage = "zh"
	default:
		currentLanguage = "en"
	}

	if currentBundle == nil {
		currentBundle = newBundle()
		_ = loadEmbeddedLocaleFiles()
	}
	currentLocalizer = i18n.NewLocalizer(currentBundle, currentLanguage)
}

func Translate(msg string) string {
	if currentLocalizer == nil {
		return msg
	}

	translated, err := currentLocalizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{ID: msg, Other: msg},
	})
	if err != nil || translated == "" {
		return msg
	}
	return translated
}

var T = Translate

func loadEmbeddedLocaleFiles() error {
	entries, err := localeFS.ReadDir(defaultLocaleDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !isLocaleFile(entry.Name()) {
			continue
		}
		filePath := filepath.Join(defaultLocaleDir, entry.Name())
		data, err := localeFS.ReadFile(filePath)
		if err != nil {
			return err
		}
		if err := parseLocaleBytes(data, filePath); err != nil {
			return err
		}
	}
	return nil
}

func loadLocaleFilesFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !isLocaleFile(entry.Name()) {
			continue
		}
		filePath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if err := parseLocaleBytes(data, filePath); err != nil {
			return err
		}
	}
	return nil
}

func isLocaleFile(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".toml")
}

func parseLocaleBytes(data []byte, filePath string) error {
	_, err := currentBundle.ParseMessageFileBytes(data, filePath)
	return err
}
