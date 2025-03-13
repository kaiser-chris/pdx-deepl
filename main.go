package main

import (
	"bahmut.de/pdx-deepl/pdx"
	"bahmut.de/pdx-deepl/translator"
	"bahmut.de/pdx-deepl/util/logging"
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

	translator.Api = translator.DeeplApi{
		Token:  *token,
		ApiUrl: apiUrl,
	}

	translation, err := translator.Api.Translate(
		[]string{"[AddTextIf( Character.IsHistorical, Concatenate( Localize( 'DATA_CHARACTER_TOOLTIP_HISTORICAL' ), '\n\n' ) )][AddTextIf( Character.IsBusy, Concatenate( Localize( 'CHARACTER_IS_BUSY' ), '\n' ) )][Concept('concept_character_role', '$concept_character_roles$')]: [Character.GetAllRoleNames|v][ConcatIfNeitherEmpty(' #v and#! ', AddLocalizationIf(Character.MakeScope.Var('is_magic_researcher').IsSet, 'gg_character_role_researcher'))]$DATA_CHARACTER_NAME_TOOLTIP_COMMANDER_DETAILS$[ConcatIfNeitherEmpty('\n', AddLocalizationIf( Character.IsAgitator, 'AGITATOR_POLITICAL_MOVEMENT_IN_COUNTRY'))]\n\n[Concept('concept_character_trait', '$concept_character_traits$')]: [Character.GetTraitsDesc]\n[concept_interest_group]: [Character.GetInterestGroup.GetName]\n[concept_ideology]: [Character.GetIdeology.GetName]\n\n[SelectLocalization(Character.IsInExilePool, 'DATA_CHARACTER_NAME_TOOLTIP_LOCATION_EXILE', 'DATA_CHARACTER_NAME_TOOLTIP_LOCATION_COUNTRY')][ConcatIfNeitherEmpty('\n', AddLocalizationIf(Character.IsAgitator, 'EXILE_HOME_COUNTRY'))]\n[concept_culture]: [Character.GetCulture.GetName]\n[concept_religion]: [Character.GetReligion.GetName]\n[concept_popularity]: #tooltippable_name #tooltip:[Character.GetTooltipTag],POPULARITY_BREAKDOWN [LabelingHelper.GetLabelForPopularityCFixedPoint(Character.GetPopularity)]#!#! (#tooltippable #tooltip:[Character.GetTooltipTag],POPULARITY_BREAKDOWN [Character.GetPopularity|0]#!#!)\nAge: [Character.GetAge|v]"},
		"EN",
		"DE",
	)
	if err != nil {
		logging.Fatalf("Could not translate translations %s", err.Error())
		os.Exit(1)
	}

	logging.Infof("Translations: %s", translation.Translations)
}
