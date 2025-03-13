package pdx

import (
	"bahmut.de/pdx-deepl/logging"
	"bufio"
	"hash/crc32"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var localizationRegex = regexp.MustCompile("^\\s*(?P<locKey>.+):\\s*\"(?P<loc>.*)\"\\s*(?P<hash>#deepl:.*)?(?:#.*)?$")
var crc32q = crc32.MakeTable(0xD5828281)

type LocalizationLanguage struct {
	Name   string
	Locale string
	Files  []*LocalizationFile
}

type LocalizationFile struct {
	FileName      string
	Localizations []*Localization
}

type Localization struct {
	Key             string
	Text            string
	Checksum        uint32
	CompareChecksum uint32
}

func readLanguage(localizationDirectory string, name string) (*LocalizationLanguage, error) {
	language := LocalizationLanguage{
		Name:   name,
		Locale: Languages[name],
	}

	languageDirectory := path.Join(localizationDirectory, name)

	err := filepath.WalkDir(languageDirectory, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		localizations := findAll(localizationRegex, string(content))
		if len(localizations) == 0 {
			localization, err := readLocalizationFile(path, info.Name())
			if err != nil {
				return err
			}
			language.Files = append(language.Files, localization)
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return &language, nil
}

func readLocalizationFile(path, filename string) (*LocalizationFile, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	file := &LocalizationFile{
		FileName:      filename,
		Localizations: make([]*Localization, 0),
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		matches := findAll(localizationRegex, scanner.Text())
		checksum := crc32.Checksum([]byte(matches["loc"]), crc32q)

		localization := &Localization{
			Key:      matches["locKey"],
			Text:     matches["loc"],
			Checksum: checksum,
		}
		if matches["hash"] != "" {
			pureHash, _ := strings.CutPrefix(matches["hash"], "#deepl:")
			checksum, err := strconv.Atoi(pureHash)
			if err != nil {
				logging.Warnf("Could not parse existsing compare checksum (%s) in file: %s", matches["hash"], path)
				break
			} else {
				localization.Checksum = uint32(checksum)
			}
		}
		file.Localizations = append(file.Localizations, localization)
	}

	if len(file.Localizations) == 0 {
		logging.Infof("Nothing found for localization file %s", path)
		return nil, nil
	}

	return file, nil
}

func findAll(expression *regexp.Regexp, content string) (matches map[string]string) {
	match := expression.FindStringSubmatch(content)
	matches = make(map[string]string)
	for i, name := range expression.SubexpNames() {
		if i > 0 && i <= len(match) {
			matches[name] = match[i]
		}
	}
	return matches
}
