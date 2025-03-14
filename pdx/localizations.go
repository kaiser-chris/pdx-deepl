package pdx

import (
	"bahmut.de/pdx-deepl/logging"
	"bufio"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var localizationRegex = regexp.MustCompile("^\\s*(?P<locKey>.+):\\s*\"(?P<loc>.*)\"\\s*(?P<hash>#deepl:.*)?(?:#.*)?$")
var crc32q = crc32.MakeTable(0xD5828281)

type LocalizationLanguage struct {
	Name      string
	Locale    string
	Directory string
	Files     map[string]*LocalizationFile
}

type LocalizationFile struct {
	FileName      string
	Path          string
	Localizations map[string]*Localization
}

type Localization struct {
	Key             string
	Text            string
	Checksum        uint32
	CompareChecksum uint32
}

func readLanguage(localizationDirectory string, name string) (*LocalizationLanguage, error) {
	languageDirectory := filepath.Join(localizationDirectory, name)

	if _, err := os.Stat(languageDirectory); os.IsNotExist(err) {
		return nil, fmt.Errorf("language directory could not be found: %s", languageDirectory)
	}

	locale, ok := Languages[name]
	if !ok {
		return nil, fmt.Errorf("language locale not supported: %s", name)
	}

	language := LocalizationLanguage{
		Name:      name,
		Directory: languageDirectory,
		Locale:    locale,
		Files:     make(map[string]*LocalizationFile),
	}

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
			localization, err := readLocalizationFile(path)
			if err != nil {
				return err
			}
			language.Files[localization.FileName] = localization
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return &language, nil
}

func readLocalizationFile(file string) (*LocalizationFile, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(file)

	localizationFile := &LocalizationFile{
		FileName:      filename,
		Path:          file,
		Localizations: make(map[string]*Localization),
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
				logging.Warnf("Could not parse existsing compare checksum (%s) in file: %s", matches["hash"], file)
				break
			} else {
				localization.Checksum = uint32(checksum)
			}
		}
		localizationFile.Localizations[localization.Key] = localization
	}

	if len(localizationFile.Localizations) == 0 {
		logging.Infof("Nothing found for localization file %s", file)
		return nil, nil
	}

	return localizationFile, nil
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
