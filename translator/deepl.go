package translator

import (
	"bahmut.de/pdx-deepl/util/logging"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type ApiRequest struct {
	Translate  []string `json:"text"`
	TargetLang string   `json:"target_lang"`
	SourceLang string   `json:"source_lang"`
}

type ApiResponse struct {
	Translations []ApiTranslation `json:"translations"`
}

type ApiTranslation struct {
	SourceLang  string `json:"detected_source_language"`
	Translation string `json:"text"`
}

type DeeplApi struct {
	ApiUrl *url.URL
	Token  string
}

var Api DeeplApi

func (api DeeplApi) Translate(translate []string, sourceLang string, targetLang string) (ApiResponse, error) {

	apiRequest := ApiRequest{
		Translate:  translate,
		TargetLang: targetLang,
		SourceLang: sourceLang,
	}

	requestBody, err := json.Marshal(apiRequest)
	if err != nil {
		return ApiResponse{}, err
	}

	request, err := http.NewRequest(
		"POST",
		api.ApiUrl.String(),
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return ApiResponse{}, err
	}
	request.Header.Set("Authorization", "DeepL-Auth-Key "+api.Token)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return ApiResponse{}, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return ApiResponse{}, err
	}
	logging.Tracef("Deepl Response: %s", string(body))

	var apiResponse ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return ApiResponse{}, err
	}

	err = response.Body.Close()
	if err != nil {
		return ApiResponse{}, err
	}

	return apiResponse, err
}
