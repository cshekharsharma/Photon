package i18n

import (
	"context"
	"testing"

	"github.com/cshekharsharma/photon/middleware"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Mock middleware for GetLocale function
func init() {}

func resetI18nState() {
	isEnabled = false
	defaultLocale = ""
	supportedLocales = nil

	translators.Range(func(key, _ interface{}) bool {
		translators.Delete(key)
		return true
	})
	messageStore.Range(func(key, _ interface{}) bool {
		messageStore.Delete(key)
		return true
	})
}

func TestInit(t *testing.T) {
	resetI18nState()
	config := &I18nConfig{
		IsEnabled:        true,
		SupportedLocales: []string{"en-US", "fr-FR"},
		Translations: map[string]map[string]string{
			"en-US": {"hello": "Hello"},
			"fr-FR": {"hello": "Bonjour"},
		},
	}

	Init(config)
	assert.True(t, isEnabled)
	assert.Equal(t, []string{"en-US", "fr-FR"}, supportedLocales)

	loadedTranslation, ok := translators.Load("en-US")
	assert.NotNil(t, loadedTranslation)
	assert.True(t, ok)

	// Test disabled scenario
	config.IsEnabled = false
	Init(config)
	assert.False(t, isEnabled)
}

func TestT(t *testing.T) {
	resetI18nState()
	ctx := context.WithValue(context.Background(), middleware.LocaleKey, "en-US")

	translators.Store("en-US", message.NewPrinter(language.English))
	messageStore.Store("en-US", map[string]string{
		"hello": "Hello, %NAME%!",
		"bye":   "Goodbye, %NAME%!",
	})

	result := T(ctx, "hello", map[string]interface{}{"%NAME%": "John"})
	assert.Equal(t, "Hello, John!", result)

	result = T(ctx, "missing", nil)
	assert.Equal(t, "missing", result)

	numResult := T(ctx, "bye", map[string]interface{}{"%NAME%": 123})
	assert.Equal(t, "Goodbye, 123!", numResult)
}

func TestT_DefaultLocaleFallback(t *testing.T) {
	resetI18nState()
	defaultLocale = "en-US"

	translators.Store("en-US", message.NewPrinter(language.English))
	messageStore.Store("en-US", map[string]string{
		"hello": "Hello, %NAME%!",
	})

	ctx := context.WithValue(context.Background(), middleware.LocaleKey, "")
	result := T(ctx, "hello", map[string]interface{}{"%NAME%": "Sam"})
	assert.Equal(t, "Hello, Sam!", result)
}

func TestT_MissingLocaleAndBadMessageStore(t *testing.T) {
	resetI18nState()
	ctx := context.WithValue(context.Background(), middleware.LocaleKey, "en-US")

	// Missing locale should return key
	assert.Equal(t, "missing", T(ctx, "missing", nil))

	// Wrong messageStore type should return key
	messageStore.Store("en-US", "not-a-map")
	assert.Equal(t, "hello", T(ctx, "hello", nil))
}

func TestT_NumberFormattingWithoutTranslator(t *testing.T) {
	resetI18nState()
	ctx := context.WithValue(context.Background(), middleware.LocaleKey, "en-US")

	messageStore.Store("en-US", map[string]string{
		"bye": "Goodbye, %NAME%!",
	})

	// Missing translator entry
	result := T(ctx, "bye", map[string]interface{}{"%NAME%": 456})
	assert.Equal(t, "Goodbye, 456!", result)

	// Translator entry with wrong type
	translators.Store("en-US", "not-a-printer")
	result = T(ctx, "bye", map[string]interface{}{"%NAME%": 789})
	assert.Equal(t, "Goodbye, 789!", result)
}

func TestPopulateTranslators_InvalidLocale(t *testing.T) {
	resetI18nState()
	populateTranslators([]string{"bad_locale", "en-US"}, map[string]map[string]string{
		"en-US": {"hello": "Hello"},
	})

	_, bad := translators.Load("bad_locale")
	assert.False(t, bad)

	_, good := translators.Load("en-US")
	assert.True(t, good)
}

func TestIsI18nEnabled(t *testing.T) {
	resetI18nState()
	isEnabled = true
	assert.True(t, IsI18nEnabled())

	isEnabled = false
	assert.False(t, IsI18nEnabled())
}

func TestFixLocaleFormat(t *testing.T) {
	resetI18nState()
	assert.Equal(t, "en-US", FixLocaleFormat("en_US"))
}
