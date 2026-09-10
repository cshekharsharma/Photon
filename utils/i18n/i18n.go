// Package i18n provides a foundational implementation of a translation service using the Go x/text package.
// This service facilitates text translation for applications, leveraging JSON format for translation files.
// It is designed with the flexibility to be extended to support other translation data sources such as databases
// or remote services in the future.
//
// Key Features:
// - Support for JSON-based translation files.
// - Easy integration with context-based locale determination.
// - Dynamic placeholder substitution in translation strings.
package i18n

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/cshekharsharma/photon/middleware"
	"github.com/cshekharsharma/photon/utils/types"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

var (
	isEnabled        bool     // Flag to indicate if the i18n service is enabled.
	defaultLocale    string   // The default locale to use if no specific locale is determined.
	supportedLocales []string // A list of locales that are supported by the application.

	translators  sync.Map // Concurrent map storing language.Tag to *message.Printer mappings for each locale.
	messageStore sync.Map // Concurrent map storing locale to its corresponding translation key-value pairs.
)

// T translates a key into the corresponding text for the specified locale in the provided context.
// If placeholders are provided, it dynamically replaces them in the translated string.
//
// Parameters:
//   - ctx: Context containing locale information, usually derived from incoming requests.
//   - key: The key corresponding to the text that needs to be translated.
//   - placeholders: A map of placeholder keys and their respective values to be substituted in the translation.
//
// Returns:
//   - The translated string with placeholders substituted, or the key itself if translation cannot be performed.
//
// Usage:
//   - Use this function to fetch a localized version of text based on the user's locale, automatically substituting any dynamic content.
func T(ctx context.Context, key string, placeholders map[string]interface{}) string {
	locale := middleware.GetLocale(ctx) // Extract locale from context using middleware.
	if locale == "" {
		locale = defaultLocale // Fallback to default locale if none is specified.
	}

	localeMessages, found := messageStore.Load(locale)
	if !found {
		return key // Return the key itself if no translations are found for the locale.
	}

	messagesMap, isMap := localeMessages.(map[string]string)
	if !isMap {
		return key // Return the key if stored translations are not in the expected map format.
	}

	messageString, ok := messagesMap[key]
	if !ok {
		return key // Return the key if no translation is found for the specified key.
	}

	if len(placeholders) > 0 {
		for pKey, pVal := range placeholders {
			if types.IsNumber(pVal) {
				tLocale, found := translators.Load(locale)
				if found {
					if tLocale, casted := tLocale.(*message.Printer); casted {
						pVal = formatNumber(tLocale, pVal) // Format numbers to locale-specific representations.
					}
				}
			}
			messageString = strings.ReplaceAll(messageString, pKey, fmt.Sprint(pVal))
		}
	}

	return messageString
}

// Init initializes the internationalization system with the specified configuration.
// It sets up the necessary infrastructure based on the provided settings.
//
// Parameters:
//   - config: A configuration object containing settings such as default locale, supported locales, and translations.
//
// Usage:
//   - Call this function at application startup to configure and enable the internationalization capabilities.
func Init(config *I18nConfig) {
	isEnabled = config.IsEnabled
	if !config.IsEnabled {
		return // Exit initialization if internationalization is disabled.
	}

	defaultLocale = config.DefaultLocale
	supportedLocales = config.SupportedLocales
	populateTranslators(supportedLocales, config.Translations) // Populate translators for each supported locale.
}

// IsI18nEnabled returns a boolean indicating whether the internationalization service is currently enabled.
//
// Returns:
//   - true if i18n is enabled, false otherwise.
//
// Usage:
//   - Use this function to check if internationalization features should be utilized in the application.
func IsI18nEnabled() bool {
	return isEnabled
}

// formatNumber formats a numerical value into a string representation that is appropriate for the given locale.
//
// Parameters:
//   - translator: A *message.Printer associated with the locale.
//   - num: The numerical value to format.
//
// Returns:
//   - A string representation of the number formatted according to locale-specific rules.
func formatNumber(translator *message.Printer, num interface{}) string {
	return translator.Sprintf("%v", number.Decimal(num)) // Format number using the locale-specific formatter.
}

// populateTranslators sets up translators for each of the supported locales using the provided translations.
// This function parses each locale string into a language tag and stores the corresponding *message.Printer and translations.
//
// Parameters:
//   - localeList: A list of locale strings that are supported.
//   - translations: A map from locale strings to their corresponding translation key-value pairs.
//
// Usage:
//   - This function is called internally by Init to prepare the translation infrastructure.
func populateTranslators(localeList []string, translations map[string]map[string]string) {
	for _, locale := range localeList {
		languageTag, cErr := language.Parse(locale)
		if cErr != nil {
			continue // Skip locales that cannot be parsed.
		}

		translators.Store(locale, message.NewPrinter(languageTag)) // Store a new Printer for the locale.

		if localeMsgs, ok := translations[locale]; ok {
			messageStore.Store(locale, localeMsgs) // Store the translations for the locale.
		}
	}
}

// FixLocaleFormat corrects locale string formats, replacing underscores with dashes.
// This standardization ensures consistency in locale identification.
//
// Parameters:
//   - locale: The locale string to correct.
//
// Returns:
//   - A standardized locale string with underscores replaced by dashes.
//
// Usage:
//   - Use this function to ensure locale strings are formatted consistently across the application.
func FixLocaleFormat(locale string) string {
	return strings.ReplaceAll(locale, "_", "-")
}
