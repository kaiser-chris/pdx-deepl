package pdx

import (
	"bahmut.de/pdx-deepl/deepl"
	"bahmut.de/pdx-deepl/logging"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

const ignoreTagStart = "<ignore>"
const ignoreTagEnd = "</ignore>"

var regexFormatting = regexp.MustCompile(`#[a-zA-Z]+\s`)

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
	logging.Infof("%sBase Language:%s %s", logging.AnsiBoldOn, logging.AnsiAllDefault, baseLanguage.Name)

	translator.BaseLanguage = baseLanguage

	for _, targetLanguageConfig := range translator.Config.TargetLanguages {
		targetLanguage, err := translator.readTargetLanguage(targetLanguageConfig.Name)
		if err != nil {
			return err
		}
		logging.Infof("%sTranslating:%s %s", logging.AnsiBoldOn, logging.AnsiAllDefault, targetLanguage.Name)
		translatedLanguage, err := translator.translateTargetLanguage(targetLanguage, targetLanguageConfig.Glossary)
		translator.TargetLanguages = append(translator.TargetLanguages, translatedLanguage)
		err = targetLanguage.Write()
		if err != nil {
			return fmt.Errorf("error writing target language (%s): %v", targetLanguage.Name, err)
		}
		logging.Infof("%sTranslated:%s %s", logging.AnsiBoldOn, logging.AnsiAllDefault, targetLanguage.Name)
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
		if translatedFile == nil {
			continue
		}
		targetLanguage.Files[key] = translatedFile
	}

	return targetLanguage, nil
}

func (translator *ParadoxTranslator) translateTargetFile(baseFile, targetFile *LocalizationFile, targetLanguage *LocalizationLanguage, glossary string) (*LocalizationFile, error) {
	if slices.Contains(translator.Config.IgnoreFiles, baseFile.FileName) {
		logging.Warnf("Skipped ignored file: %s", baseFile.FileName)
		return nil, nil
	}

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

	counterManual := 0
	counterUpToDate := 0
	counterTranslated := 0
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
			counterManual++
			continue
		}
		if localization.Checksum == targetLocalization.CompareChecksum {
			// Localization was already translated
			// and is up to date
			counterUpToDate++
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
		counterTranslated++
	}

	targetLanguage.Files[baseFile.Key] = file
	if counterTranslated > 0 {
		logging.Infof(
			"%s%s%s: Translated %s%d%s localization keys",
			logging.AnsiBoldOn, file.FileName, logging.AnsiAllDefault,
			logging.AnsiBoldOn, counterTranslated, logging.AnsiAllDefault,
		)
	}
	if counterManual > 0 {
		logging.Infof(
			"%s%s%s: Found %s%d%s manually translated localization keys",
			logging.AnsiBoldOn, file.FileName, logging.AnsiAllDefault,
			logging.AnsiBoldOn, counterManual, logging.AnsiAllDefault,
		)
	}
	if counterUpToDate > 0 {
		logging.Infof(
			"%s%s%s: Found %s%d%s up to date localization keys",
			logging.AnsiBoldOn, file.FileName, logging.AnsiAllDefault,
			logging.AnsiBoldOn, counterUpToDate, logging.AnsiAllDefault,
		)
	}
	if counterUpToDate == 0 && counterTranslated == 0 && counterManual == 0 {
		logging.Warnf(
			"%s%s%s: Translated %sno%s localization keys",
			logging.AnsiBoldOn, file.FileName, logging.AnsiAllDefault,
			logging.AnsiBoldOn, logging.AnsiAllDefault,
		)
	}

	return file, nil
}

func (translator *ParadoxTranslator) translateLocalization(content string, targetLanguage *LocalizationLanguage, glossary string) (string, error) {
	// Escape functions
	requestContent := strings.ReplaceAll(content, "[", ignoreTagStart+"[")
	requestContent = strings.ReplaceAll(requestContent, "]", "]"+ignoreTagEnd)

	// Escape loc references
	for strings.Contains(requestContent, "$") {
		requestContent = strings.Replace(requestContent, "$", "<ref>", 1)
		requestContent = strings.Replace(requestContent, "$", "</ref>", 1)
	}

	// Escape formatting
	requestContent = strings.ReplaceAll(requestContent, "#!", ignoreTagStart+"#!"+ignoreTagEnd)
	for _, formatting := range regexFormatting.FindAllString(requestContent, -1) {
		requestContent = strings.ReplaceAll(requestContent, formatting, ignoreTagStart+formatting+ignoreTagEnd)
	}

	response, err := translator.Api.Translate(
		[]string{requestContent},
		translator.BaseLanguage.Locale,
		targetLanguage.Locale,
		[]string{"ignore", "ref"},
		glossary,
	)

	if err != nil {
		return "", err
	}

	result := response.Translations[0].Translation
	result = strings.ReplaceAll(result, ignoreTagStart, "")
	result = strings.ReplaceAll(result, ignoreTagEnd, "")
	result = strings.ReplaceAll(result, "<ref>", "$")
	result = strings.ReplaceAll(result, "</ref>", "$")

	return result, nil
}
