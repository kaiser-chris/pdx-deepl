package main

import (
	"bahmut.de/pdx-deepl/deepl"
	"bahmut.de/pdx-deepl/logging"
	"bahmut.de/pdx-deepl/pdx"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

const (
	ApiFree = "free"
	ApiPaid = "paid"
)

const (
	FlagApiType      = "api-type"
	FlagApiToken     = "api-token"
	FlagConfig       = "config"
	FlagStatistics   = "stats"
	FlagLocalization = "localization"
)

func main() {
	localizationLocation := flag.String(FlagLocalization, ".", "Optional: Path to localization directory of your mod")
	apiType := flag.String(FlagApiType, ApiFree, "Optional: Whether to use free or paid Deepl API")
	token := flag.String(FlagApiToken, "", "Required: Deepl API Token")
	config := flag.String(FlagConfig, pdx.DefaultConfigFile, "Optional: Path to translation config file")
	stats := flag.Bool(FlagStatistics, false, "Optional: When set produces relevant statistics about the localization like the character count")
	flag.Parse()

	if token == nil || *token == "" {
		fmt.Printf("The parameter %s%s%s is required.\n\n", logging.AnsiBoldOn, FlagApiToken, logging.AnsiAllDefault)
		flag.PrintDefaults()
		os.Exit(1)
	}

	var resolvedConfigFile string
	if config == nil || *config == "" {
		resolvedConfigFile = pdx.DefaultConfigFile
	} else {
		resolvedConfigFile = *config
	}

	configPath, err := filepath.Abs(resolvedConfigFile)
	if err != nil {
		logging.Fatalf("Could not load config file: %s", err)
		os.Exit(1)
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logging.Fatalf("Config file does not exist: %s", configPath)
		os.Exit(1)
	}

	logging.Infof("%sTranslation Config:%s %s", logging.AnsiBoldOn, logging.AnsiAllDefault, configPath)

	var resolvedLocalizationDirectory string
	if localizationLocation == nil || *localizationLocation == "" {
		resolvedLocalizationDirectory = "."
	} else {
		resolvedLocalizationDirectory = *localizationLocation
	}

	localizationPath, err := filepath.Abs(resolvedLocalizationDirectory)
	if err != nil {
		logging.Fatalf("Could not load Localization directory: %s", err)
		os.Exit(1)
	}
	if _, err := os.Stat(localizationPath); os.IsNotExist(err) {
		logging.Fatalf("Localization directory does not exist: %s", localizationPath)
		os.Exit(1)
	}

	logging.Infof("%sLocalization Directory:%s %s", logging.AnsiBoldOn, logging.AnsiAllDefault, localizationPath)

	var apiUrl *url.URL

	var resolvedApiType string

	if apiType == nil {
		resolvedApiType = ApiFree
	} else {
		resolvedApiType = *apiType
	}

	switch resolvedApiType {
	case ApiFree:
		parsedUrl, err := url.Parse("https://api-free.deepl.com/v2/")
		if err != nil {
			logging.Fatal("Could not parse free api url")
		}
		apiUrl = parsedUrl
	case ApiPaid:
		parsedUrl, err := url.Parse("https://api.deepl.com/v2/")
		if err != nil {
			logging.Fatal("Could not parse paid api url")
			os.Exit(1)
		}
		apiUrl = parsedUrl
		break
	default:
		logging.Fatalf("API type %s%s%s unknown please choose one of %s or %s", logging.AnsiBoldOn, resolvedApiType, logging.AnsiAllDefault, ApiFree, ApiPaid)
	}
	logging.Infof("%sAPI Type:%s %s", logging.AnsiBoldOn, logging.AnsiAllDefault, resolvedApiType)

	translatorApi := deepl.CreateApi(apiUrl, *token)
	response, err := translatorApi.Usage()
	if err != nil {
		logging.Fatalf("Could not initialize %sDeepl API%s: %s", logging.AnsiBoldOn, logging.AnsiAllDefault, err.Error())
		os.Exit(1)
	}
	logging.Infof("%sAPI Character Usage:%s %d", logging.AnsiBoldOn, logging.AnsiAllDefault, response.CharacterCount)
	logging.Infof("%sAPI Character Limit:%s %d", logging.AnsiBoldOn, logging.AnsiAllDefault, response.CharacterLimit)

	translatorPdx, err := pdx.CreateTranslator(resolvedConfigFile, resolvedLocalizationDirectory, translatorApi)
	if err != nil {
		logging.Fatalf("Could not initialize %sPDX Translator%s: %s", logging.AnsiBoldOn, logging.AnsiAllDefault, err.Error())
		os.Exit(1)
	}

	if stats != nil && *stats {
		err = translatorPdx.Statistics()
		if err != nil {
			logging.Fatalf("Could not calculate %sStatistics%s: %s", logging.AnsiBoldOn, logging.AnsiAllDefault, err.Error())
			os.Exit(1)
		}
		return
	}

	err = translatorPdx.Translate()
	if err != nil {
		logging.Fatalf("Could not run %sPDX Translator%s: %s", logging.AnsiBoldOn, logging.AnsiAllDefault, err.Error())
		os.Exit(1)
	}

	logging.Infof("%sTranslation was run successfully%s", logging.AnsiBoldOn, logging.AnsiAllDefault)
}
