package pdx

import (
	"bahmut.de/pdx-deepl/logging"
	"regexp"
	"unicode/utf8"
)

var ignoreTagRegex = regexp.MustCompile(`<ignore>.*?</ignore>`)
var referenceTagRegex = regexp.MustCompile(`<ref>.*?</ref>`)

func (translator *ParadoxTranslator) Statistics() error {
	baseLanguage, err := readLanguage(translator.LocalizationDirectory, translator.Config.BaseLanguage)
	if err != nil {
		return err
	}
	logging.Infof("%sBase Language:%s %s", logging.AnsiBoldOn, logging.AnsiAllDefault, baseLanguage.Name)
	logging.Infof("%sLocalization Files:%s %d", logging.AnsiBoldOn, logging.AnsiAllDefault, len(baseLanguage.Files))
	keyCount := 0
	characterCount := 0

	for _, file := range baseLanguage.Files {
		keyCount = keyCount + len(file.Localizations)
		for _, localization := range file.Localizations {
			clean := ignoreTagRegex.ReplaceAllString(localization.Text, "")
			clean = referenceTagRegex.ReplaceAllString(localization.Text, "")
			characterCount = characterCount + utf8.RuneCountInString(clean)
		}
	}

	logging.Infof("%sLocalization Keys:%s %d", logging.AnsiBoldOn, logging.AnsiAllDefault, keyCount)
	logging.Infof("%sTotal Characters:%s %d", logging.AnsiBoldOn, logging.AnsiAllDefault, characterCount)
	return nil
}
