package pdx

import (
	"encoding/json"
	"fmt"
	"os"
)

const DefaultConfigFile = "translation-config.json"

type TranslationConfiguration struct {
	BaseLanguage    string                              `json:"base-language"`
	TargetLanguages []*TranslationConfigurationLanguage `json:"target-languages"`
	IgnoreFiles     []string                            `json:"ignore-files"`
}

type TranslationConfigurationLanguage struct {
	Name     string `json:"name"`
	Glossary string `json:"glossary"`
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

	if len(translationConfiguration.TargetLanguages) == 0 {
		return nil, fmt.Errorf("no target languages found in config file: %s", path)
	}

	return &translationConfiguration, nil
}
