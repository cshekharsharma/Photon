package i18n

// I18nConfig holds the configuration settings for the internationalization (i18n) service.
// This struct includes settings to enable or disable i18n, specify the default and supported locales,
// and provide the necessary translations for each locale.
//
// Fields:
//
//   - IsEnabled: A boolean indicating whether the i18n service should be enabled. If false, the i18n functionalities
//     will not be initialized or used, and the service may bypass translation logic to optimize performance.
//
//   - DefaultLocale: The default locale to use when no specific locale is provided or detected. This locale's
//     translations are used as fallbacks if a translation is not available in the user's current locale.
//
//   - SupportedLocales: A list of locales that the application explicitly supports. This list is used to validate
//     locale requests and to load the corresponding translations. Only locales listed here will have their translations
//     loaded into the system.
//
//   - Translations: A map where each key is a locale identifier and its value is another map of translation keys
//     and their corresponding localized strings. This nested map structure allows easy access to any specific
//     translation based on a locale and a translation key.
type I18nConfig struct {
	IsEnabled        bool
	DefaultLocale    string
	SupportedLocales []string
	Translations     map[string]map[string]string
}
