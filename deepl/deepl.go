package deepl

import (
	"bahmut.de/pdx-deepl/logging"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const EndpointTranslate = "translate"
const EndpointUsage = "usage"

type TranslationRequest struct {
	Translate        []string `json:"text"`
	TargetLang       string   `json:"target_lang"`
	SourceLang       string   `json:"source_lang"`
	TagHandling      string   `json:"tag_handling"`
	IgnoreTags       []string `json:"ignore_tags"`
	OutlineDetection bool     `json:"outline_detection"`
	Glossary         string   `json:"glossary_id"`
}

type TranslationResponse struct {
	Translations []*ApiTranslation `json:"translations"`
}

type UsageResponse struct {
	CharacterCount int `json:"character_count"`
	CharacterLimit int `json:"character_limit"`
}

type ApiTranslation struct {
	SourceLang  string `json:"detected_source_language"`
	Translation string `json:"text"`
}

type Api struct {
	ApiUrl *url.URL
	Token  string
}

func CreateApi(apiUrl *url.URL, token string) *Api {
	return &Api{
		ApiUrl: apiUrl,
		Token:  token,
	}
}

func (api Api) Usage() (*UsageResponse, error) {
	usageUrl := api.ApiUrl.JoinPath(EndpointUsage)
	request, err := http.NewRequest(
		"GET",
		usageUrl.String(),
		nil,
	)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "DeepL-Auth-Key "+api.Token)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = convertApiError(response.StatusCode)
	if err != nil {
		logging.Tracef("Deepl Response: %s", string(body))
		return nil, err
	}

	var apiResponse UsageResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	err = response.Body.Close()
	if err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func (api Api) Translate(
	translate []string,
	sourceLang string,
	targetLang string,
	ignoreTags []string,
	glossary string,
) (*TranslationResponse, error) {
	apiRequest := TranslationRequest{
		Translate:  translate,
		TargetLang: targetLang,
		SourceLang: sourceLang,
	}

	translateUrl := api.ApiUrl.JoinPath(EndpointTranslate)

	if ignoreTags != nil && len(ignoreTags) > 0 {
		apiRequest.OutlineDetection = false
		apiRequest.TagHandling = "xml"
		apiRequest.IgnoreTags = ignoreTags
	}

	if glossary != "" {
		apiRequest.Glossary = glossary
	}

	requestBody, err := json.Marshal(apiRequest)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(
		"POST",
		translateUrl.String(),
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "DeepL-Auth-Key "+api.Token)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = convertApiError(response.StatusCode)
	if err != nil {
		logging.Tracef("Deepl Response: %s", string(body))
		return nil, err
	}

	var apiResponse TranslationResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	err = response.Body.Close()
	if err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

// API errors are documented here:
// https://developers.deepl.com/docs/best-practices/error-handling
// https://support.deepl.com/hc/en-us/articles/9773964275868-DeepL-API-error-messages
func convertApiError(statusCode int) error {
	switch statusCode {
	case 200:
		return nil
	case 400:
		return errors.New("bad request (please report this issue)")
	case 403:
		return errors.New("invalid token")
	case 404:
		return errors.New("the requested resource could not be found (please report this issue)")
	case 413:
		return errors.New("the request size exceeds the limit (what are you doing?)")
	case 429:
		return errors.New("too many requests (restart the tool to translate skipped keys)")
	case 456:
		logging.Fatal("quota exceeded")
		return nil
	default:
		return fmt.Errorf("%d: internal deepl api error", statusCode)
	}
}
