package pdx

import (
	"bahmut.de/pdx-deepl/deepl"
	"bahmut.de/pdx-deepl/logging"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ParadoxTranslator struct {
	Config                *TranslationConfiguration
	LocalizationDirectory string
	Api                   *deepl.Api
	BaseLanguage          *LocalizationLanguage
	TargetLanguages       []*LocalizationLanguage
}

func CreateTranslator(configFile, localizationDirectory string, api *deepl.Api) (*ParadoxTranslator, error) {
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

	for _, targetLanguageConfig := range translator.Config.TargetLanguages {
		targetLanguage, err := translator.readTargetLanguage(targetLanguageConfig.Name)
		if err != nil {
			return err
		}
		translatedLanguage, err := translator.translateTargetLanguage(targetLanguage, targetLanguageConfig.Glossary)
		translator.TargetLanguages = append(translator.TargetLanguages, translatedLanguage)
		err = targetLanguage.Write()
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

func (translator *ParadoxTranslator) translateTargetLanguage(targetLanguage *LocalizationLanguage, glossary string) (*LocalizationLanguage, error) {
	for key, file := range translator.BaseLanguage.Files {
		translatedFile, err := translator.translateTargetFile(file, targetLanguage.Files[key], targetLanguage, glossary)
		if err != nil {
			return nil, err
		}
		targetLanguage.Files[key] = translatedFile
	}

	return targetLanguage, nil
}

func (translator *ParadoxTranslator) translateTargetFile(baseFile, targetFile *LocalizationFile, targetLanguage *LocalizationLanguage, glossary string) (*LocalizationFile, error) {
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
		response, err := translator.translateLocalization(localization.Text, targetLanguage, glossary)
		time.Sleep(500 * time.Millisecond)
		if err != nil {
			// Too many requests
			if strings.Contains(err.Error(), "429") {
				logging.Errorf("Stopped translation for file (%s) because of: %v", baseFile.FileName, err)
				time.Sleep(1000 * time.Millisecond)
				break
			}

			// Localization was already translated
			// and is up to date
			logging.Warnf("Ignored localization key (%s) in file (%s) because of an error: %s", key, baseFile.FileName, err)
			continue
		}
		targetLocalization.Text = response
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

func (translator *ParadoxTranslator) translateLocalization(content string, targetLanguage *LocalizationLanguage, glossary string) (string, error) {
	requestContent := strings.ReplaceAll(content, "[", "<func>")
	requestContent = strings.ReplaceAll(requestContent, "]", "</func>")
	for strings.Contains(requestContent, "$") {
		requestContent = strings.Replace(requestContent, "$", "<ref>", 1)
		requestContent = strings.Replace(requestContent, "$", "</ref>", 1)
	}

	response, err := translator.Api.Translate(
		[]string{requestContent},
		translator.BaseLanguage.Locale,
		targetLanguage.Locale,
		[]string{"func", "ref"},
		glossary,
	)

	if err != nil {
		return "", err
	}

	result := response.Translations[0].Translation
	result = strings.ReplaceAll(result, "<func>", "[")
	result = strings.ReplaceAll(result, "</func>", "]")
	result = strings.ReplaceAll(result, "<ref>", "$")
	result = strings.ReplaceAll(result, "</ref>", "$")

	return result, nil
}
