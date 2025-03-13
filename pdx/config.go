package pdx

const DefaultConfigFile = "translation-config.json"

type TranslationConfiguration struct {
	BaseLanguage         string   `json:"base-language"`
	TranslationLanguages []string `json:"translated-languages"`
}
