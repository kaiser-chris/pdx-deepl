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

var utf8Bom = []byte{0xEF, 0xBB, 0xBF}
var localizationRegex = regexp.MustCompile(`^\s*(?P<locKey>.+):\d*\s*"(?P<loc>.*)"\s*(?P<hash>#deepl:.*)?(?:#.*)?$`)
var crc32q = crc32.MakeTable(0xD5828281)

type LocalizationLanguage struct {
	Name      string
	Locale    string
	Directory string
	Files     map[string]*LocalizationFile
}

type LocalizationFile struct {
	Key           string
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

func (language *LocalizationLanguage) Write() error {
	for _, file := range language.Files {
		err := file.Write(language)
		if err != nil {
			return err
		}
	}
	return nil
}

func (file *LocalizationFile) Write(language *LocalizationLanguage) error {
	var fileBuilder strings.Builder
	fileBuilder.WriteString("l_")
	fileBuilder.WriteString(language.Name)
	fileBuilder.WriteString(":\n")
	for _, localization := range file.Localizations {
		fileBuilder.WriteString(" ")
		fileBuilder.WriteString(localization.Key)
		fileBuilder.WriteString(": \"")
		fileBuilder.WriteString(localization.Text)
		fileBuilder.WriteString("\"")
		if localization.CompareChecksum != 0 {
			fileBuilder.WriteString(" #deepl:")
			fileBuilder.WriteString(strconv.Itoa(int(localization.CompareChecksum)))
			fileBuilder.WriteString("\n")
		} else {
			fileBuilder.WriteString("\n")
		}
	}

	if _, err := os.Stat(filepath.Dir(file.Path)); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(file.Path), 0755)
		if err != nil {
			return err
		}
	}

	content := append(utf8Bom, []byte(fileBuilder.String())...)

	err := os.WriteFile(file.Path, content, 0644)
	if err != nil {
		return err
	}

	return nil
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
		localization, err := readLocalizationFile(path)
		if err != nil {
			return err
		}
		if localization != nil {
			tag := fmt.Sprintf("l_%s.yml", language.Name)
			key, found := strings.CutSuffix(localization.FileName, tag)
			if !found {
				return fmt.Errorf("language tag (%s) in filename could not be found: %s", tag, localization.FileName)
			}
			localization.Key = key
			language.Files[key] = localization
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
		line := scanner.Text()
		matches := findAll(localizationRegex, line)
		checksum := crc32.Checksum([]byte(matches["loc"]), crc32q)
		if len(matches) == 0 {
			// skip line when there is no valid localization
			continue
		}

		localization := &Localization{
			Key:      matches["locKey"],
			Text:     matches["loc"],
			Checksum: checksum,
		}
		if matches["hash"] != "" {
			pureHash, _ := strings.CutPrefix(matches["hash"], "#deepl:")
			checksum, err := strconv.Atoi(pureHash)
			if err == nil {
				localization.CompareChecksum = uint32(checksum)
			} else {
				logging.Warnf("Could not parse existsing compare checksum (%s) in file: %s", matches["hash"], file)
			}
		}
		localizationFile.Localizations[localization.Key] = localization
	}

	if len(localizationFile.Localizations) == 0 {
		logging.Tracef("Nothing found for localization file %s", file)
		return localizationFile, nil
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
