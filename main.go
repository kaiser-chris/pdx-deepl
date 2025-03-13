package main

import (
	"bahmut.de/pdx-deepl/config"
	"bahmut.de/pdx-deepl/translator"
	"bahmut.de/pdx-deepl/util/logging"
	"net/url"
	"os"
)

const (
	ApiFree = "free"
	ApiPaid = "paid"
)

func main() {
	localizationLocation := config.RequireStringEnv("LOCALIZATION")
	logging.Infof("%sLocalization Directory:%s %s", logging.AnsiBoldOn, logging.AnsiAllDefault, localizationLocation)

	apiType := config.OptionalStringEnv("API_TYPE")
	if apiType == "" {
		apiType = "free"
	}
	var apiUrl *url.URL
	switch apiType {
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
		logging.Fatalf("API type %s%s%s unknown please choose one of %s or %s", logging.AnsiBoldOn, apiType, logging.AnsiAllDefault, ApiFree, ApiPaid)
	}
	logging.Infof("%sAPI Type:%s %s", logging.AnsiBoldOn, logging.AnsiAllDefault, apiType)

	translator.Api = translator.DeeplApi{
		Token:  config.RequireStringEnv("API_TOKEN"),
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
