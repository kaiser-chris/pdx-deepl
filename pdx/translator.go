package pdx

import (
	"bahmut.de/pdx-deepl/logging"
	"bahmut.de/pdx-deepl/translator"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ParadoxTranslator struct {
	Config                *TranslationConfiguration
	LocalizationDirectory string
	Api                   *translator.DeeplApi
	BaseLanguage          *LocalizationLanguage
	TargetLanguages       []*LocalizationLanguage
}

func CreateTranslator(configFile, localizationDirectory string, api *translator.DeeplApi) (*ParadoxTranslator, error) {
	config, err := readConfigFile(configFile)
	if err != nil {
		return nil, err
	}

	return &ParadoxTranslator{
		Config:                config,
		LocalizationDirectory: localizationDirectory,
		Api:                   api,
	}, nil
}

func (translator *ParadoxTranslator) Translate() error {
	baseLanguage, err := readLanguage(translator.LocalizationDirectory, translator.Config.BaseLanguage)
	if err != nil {
		return err
	}

	translator.BaseLanguage = baseLanguage

	for _, targetLanguageName := range translator.Config.TargetLanguages {
		targetLanguage, err := translator.readTargetLanguage(targetLanguageName)
		if err != nil {
			return err
		}
		translatedLanguage, err := translator.translateTargetLanguage(targetLanguage)
		translator.TargetLanguages = append(translator.TargetLanguages, translatedLanguage)
	}

	for _, targetLanguage := range translator.TargetLanguages {
		err := targetLanguage.Write()
		if err != nil {
			return fmt.Errorf("error writing target language (%s): %v", targetLanguage.Name, err)
		}
	}

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
	targetLanguage, err := readLanguage(translator.LocalizationDirectory, language)
	if err != nil {
		return nil, err
	}

	return targetLanguage, nil
}

func (translator *ParadoxTranslator) translateTargetLanguage(targetLanguage *LocalizationLanguage) (*LocalizationLanguage, error) {
	for key, file := range translator.BaseLanguage.Files {
		translatedFile, err := translator.translateTargetFile(file, targetLanguage.Files[key], targetLanguage)
		if err != nil {
			return nil, err
		}
		targetLanguage.Files[key] = translatedFile
	}

	return targetLanguage, nil
}

func (translator *ParadoxTranslator) translateTargetFile(baseFile, targetFile *LocalizationFile, targetLanguage *LocalizationLanguage) (*LocalizationFile, error) {
	var file *LocalizationFile
	if targetFile == nil {
		tag := fmt.Sprintf("l_%s.yml", targetLanguage.Name)
		name := baseFile.Key + tag
		basePath := strings.Replace(filepath.Dir(baseFile.Path), translator.BaseLanguage.Name, targetLanguage.Name, 1)
		path := filepath.Join(basePath, name)
		file = &LocalizationFile{
			Key:           baseFile.Key,
			Path:          path,
			FileName:      name,
			Localizations: make(map[string]*Localization),
		}
	} else {
		file = targetFile
	}

	counter := 0
	for key, localization := range baseFile.Localizations {
		targetLocalization, ok := file.Localizations[key]
		if !ok {
			targetLocalization = &Localization{
				Key:             key,
				CompareChecksum: 1, // Mark as to be translated
			}
		}
		if targetLocalization.CompareChecksum == 0 {
			// Don't touch manual localizations
			// in the target language
			continue
		}
		if localization.Checksum == targetLocalization.CompareChecksum {
			// Localization was already translated
			// and is up to date
			continue
		}
		response, err := translator.Api.Translate([]string{localization.Text}, translator.BaseLanguage.Locale, targetLanguage.Locale)
		if err != nil || len(response.Translations) == 0 {
			// Too many requests
			if err != nil && strings.Contains(err.Error(), "429") {
				logging.Errorf("Stopped translation for file (%s) because of: %v", baseFile.FileName, err)
				break
			}

			// Localization was already translated
			// and is up to date
			logging.Warnf("Ignored localization key (%s) in file (%s) because of an error: %s", key, baseFile.FileName, err)
			continue
		}
		targetLocalization.Text = response.Translations[0].Translation
		targetLocalization.CompareChecksum = localization.Checksum
		file.Localizations[key] = targetLocalization
		counter++
	}

	targetLanguage.Files[baseFile.Key] = file
	if counter > 0 {
		logging.Infof("Translated %d localization keys in file: %s", counter, file.FileName)
	}

	return file, nil
}
