package deepl

import (
	"bahmut.de/pdx-deepl/logging"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

type ApiRequest struct {
	Translate        []string `json:"text"`
	TargetLang       string   `json:"target_lang"`
	SourceLang       string   `json:"source_lang"`
	TagHandling      string   `json:"tag_handling"`
	IgnoreTags       []string `json:"ignore_tags"`
	OutlineDetection bool     `json:"outline_detection"`
}

type ApiResponse struct {
	Translations []*ApiTranslation `json:"translations"`
}

type ApiTranslation struct {
	SourceLang  string `json:"detected_source_language"`
	Translation string `json:"text"`
}

type DeeplApi struct {
	ApiUrl *url.URL
	Token  string
}

func CreateApi(apiUrl *url.URL, token string) *DeeplApi {
	return &DeeplApi{
		ApiUrl: apiUrl,
		Token:  token,
	}
}

func (api DeeplApi) Translate(translate []string, sourceLang string, targetLang string, ignoreTags []string) (*ApiResponse, error) {
	apiRequest := ApiRequest{
		Translate:  translate,
		TargetLang: targetLang,
		SourceLang: sourceLang,
	}

	if ignoreTags != nil && len(ignoreTags) > 0 {
		apiRequest.OutlineDetection = false
		apiRequest.TagHandling = "xml"
		apiRequest.IgnoreTags = ignoreTags
	}

	requestBody, err := json.Marshal(apiRequest)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(
		"POST",
		api.ApiUrl.String(),
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

	if response.StatusCode != 200 {
		logging.Tracef("Deepl Response: %s", string(body))
		return nil, errors.New(response.Status)
	}

	var apiResponse ApiResponse
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
