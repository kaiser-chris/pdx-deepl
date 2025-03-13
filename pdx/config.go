package pdx

import (
	"encoding/json"
	"os"
)

const DefaultConfigFile = "translation-config.json"

type TranslationConfiguration struct {
	BaseLanguage         string   `json:"base-language"`
	TranslationLanguages []string `json:"translated-languages"`
}

func readConfigFile(path string) (*TranslationConfiguration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var translationConfiguration TranslationConfiguration
	err = json.Unmarshal(data, &translationConfiguration)
	if err != nil {
		return nil, err
	}

	return &translationConfiguration, nil
}
