package main

import (
	"bahmut.de/pdx-deepl/logging"
	"bahmut.de/pdx-deepl/pdx"
	"bahmut.de/pdx-deepl/translator"
	"flag"
	"fmt"
	"net/url"
	"os"
)

const (
	ApiFree = "free"
	ApiPaid = "paid"
)

const (
	FlagApiType      = "api-type"
	FlagApiToken     = "api-token"
	FlagConfig       = "config"
	FlagLocalization = "localization"
)

func main() {
	localizationLocation := flag.String(FlagLocalization, ".", "Optional: Path to localization directory of your mod")
	apiType := flag.String(FlagApiType, ApiFree, "Optional: Whether to use free or paid Deepl API")
	token := flag.String(FlagApiToken, "", "Required: Deepl API Token")
	config := flag.String(FlagConfig, pdx.DefaultConfigFile, "Optional: Path to translation config file")
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

	logging.Infof("%sTranslation Config:%s %s", logging.AnsiBoldOn, logging.AnsiAllDefault, resolvedConfigFile)

	var resolvedLocalizationDirectory string
	if localizationLocation == nil || *localizationLocation == "" {
		resolvedLocalizationDirectory = "."
	} else {
		resolvedLocalizationDirectory = *localizationLocation
	}

	logging.Infof("%sLocalization Directory:%s %s", logging.AnsiBoldOn, logging.AnsiAllDefault, resolvedLocalizationDirectory)

	var apiUrl *url.URL

	var resolvedApiType string

	if apiType == nil {
		resolvedApiType = ApiFree
	} else {
		resolvedApiType = *apiType
	}

	switch resolvedApiType {
	case ApiFree:
		parsedUrl, err := url.Parse("https://api-free.deepl.com/v2/translate")
		if err != nil {
			logging.Fatal("Could not parse free api url")
		}
		apiUrl = parsedUrl
	case ApiPaid:
		parsedUrl, err := url.Parse("https://api.deepl.com/v2/translate")
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

	translatorApi := translator.CreateApi(apiUrl, *token)
	translatorPdx, err := pdx.CreateTranslator(resolvedConfigFile, resolvedLocalizationDirectory, translatorApi)
	if err != nil {
		logging.Fatalf("Could not initialize %sPDX Translator%s: %s", logging.AnsiBoldOn, logging.AnsiAllDefault, err.Error())
		os.Exit(1)
	}
	err = translatorPdx.Translate()
	if err != nil {
		logging.Fatalf("Could not run %sPDX Translator%s: %s", logging.AnsiBoldOn, logging.AnsiAllDefault, err.Error())
		os.Exit(1)
	}

	logging.Infof("%sTranslation was run successfully%s", logging.AnsiBoldOn, logging.AnsiAllDefault)
}
