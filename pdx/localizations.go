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

const skippedHash = "skipped"
const skippedChecksum = 1

var regexLocalization = regexp.MustCompile(`^\s*(?P<locKey>.+):\d*\s*"(?P<loc>.*)"\s*(?P<hash>#deepl:.*)?(?:#.*)?$`)
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

func (file *LocalizationFile) WriteFile(
	baseFile *LocalizationFile,
	baseLanguage *LocalizationLanguage,
	targetLanguage *LocalizationLanguage,
) error {
	// Create target language directories
	if _, err := os.Stat(filepath.Dir(file.Path)); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(file.Path), 0755)
		if err != nil {
			return err
		}
	}

	// Read the baseFile language file
	baseContent, err := os.ReadFile(baseFile.Path)
	if err != nil {
		return err
	}
	targetContent := string(baseContent)

	// Replace language Tag
	targetContent = strings.Replace(
		targetContent,
		"l_"+baseLanguage.Name+":",
		"l_"+targetLanguage.Name+":",
		1,
	)

	keys := make([]string, 0, len(file.Localizations))
	for key := range file.Localizations {
		keys = append(keys, key)
	}

	// Replace relevant lines
	for line := range strings.Lines(targetContent) {
		for _, key := range keys {
			matches := findAll(regexLocalization, line)
			localization := file.Localizations[key]
			if matches["locKey"] != key {
				continue
			}
			var lineBuilder strings.Builder
			lineBuilder.WriteString(" ")
			lineBuilder.WriteString(localization.Key)
			lineBuilder.WriteString(": \"")
			lineBuilder.WriteString(localization.Text)
			lineBuilder.WriteString("\"")
			if localization.CompareChecksum == skippedChecksum {
				lineBuilder.WriteString(" #deepl:")
				lineBuilder.WriteString(skippedHash)
			} else if localization.CompareChecksum != 0 {
				lineBuilder.WriteString(" #deepl:")
				lineBuilder.WriteString(strconv.Itoa(int(localization.CompareChecksum)))
			}
			if strings.HasSuffix(line, "\r\n") {
				lineBuilder.WriteString("\r\n")
			} else {
				lineBuilder.WriteString("\n")
			}

			targetContent = strings.Replace(
				targetContent,
				line,
				lineBuilder.String(),
				1,
			)
			break
		}
	}

	return os.WriteFile(file.Path, []byte(targetContent), 0644)
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
		localization, err := readLocalizationFile(path, &language)
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

func readLocalizationFile(file string, language *LocalizationLanguage) (*LocalizationFile, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	filename := file[len(language.Directory)+1:]

	localizationFile := &LocalizationFile{
		FileName:      filename,
		Path:          file,
		Localizations: make(map[string]*Localization),
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		matches := findAll(regexLocalization, line)
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
		if matches["hash"] != "" && !strings.Contains(matches["hash"], "#deepl:"+skippedHash) {
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
