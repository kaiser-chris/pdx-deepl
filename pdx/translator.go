package pdx

import (
	"bahmut.de/pdx-deepl/logging"
	"bahmut.de/pdx-deepl/translator"
	"os"
	"path/filepath"
)

type ParadoxTranslator struct {
	Config                *TranslationConfiguration
	LocalizationDirectory string
	Translator            *translator.DeeplApi
}

func CreateTranslator(configFile, localizationDirectory string, translator *translator.DeeplApi) (*ParadoxTranslator, error) {
	config, err := readConfigFile(configFile)
	if err != nil {
		return nil, err
	}

	return &ParadoxTranslator{
		Config:                config,
		LocalizationDirectory: localizationDirectory,
		Translator:            translator,
	}, nil
}

func (translator *ParadoxTranslator) Translate() error {
	baseLanguage, err := readLanguage(translator.LocalizationDirectory, translator.Config.BaseLanguage)
	if err != nil {
		return err
	}

	targetLanguages := make([]*LocalizationLanguage, len(translator.Config.TargetLanguages))
	for index, targetLanguage := range translator.Config.TargetLanguages {
		targetLanguages[index], err = translator.readTargetLanguage(targetLanguage)
		if err != nil {
			return err
		}
		logging.Info(targetLanguages[index])
	}

	logging.Info(baseLanguage)
	return nil
}

func (translator *ParadoxTranslator) readTargetLanguage(language string) (*LocalizationLanguage, error) {
	languagePath := filepath.Join(translator.LocalizationDirectory, language)
	if _, err := os.Stat(languagePath); os.IsNotExist(err) {
		err := os.MkdirAll(languagePath, 0755)
		if err != nil {
			return nil, err
		}
	}
	return readLanguage(translator.LocalizationDirectory, language)
}
