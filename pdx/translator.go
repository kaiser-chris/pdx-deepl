package pdx

import "bahmut.de/pdx-deepl/translator"

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
	readLanguage(translator.LocalizationDirectory, "english")
	return nil
}
