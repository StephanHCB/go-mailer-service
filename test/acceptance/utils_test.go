package acceptance

import (
	"errors"
	"github.com/go-http-utils/headers"
	"io/ioutil"
	"net/http"
)

// placing these here because they are package global

type tstWebResponse struct {
	status      int
	body        string
	contentType string
	location    string
}

func tstWebResponseFromResponse(response *http.Response) (tstWebResponse, error) {
	status := response.StatusCode
	ct := ""
	if val, ok := response.Header[headers.ContentType]; ok {
		ct = val[0]
	}
	loc := ""
	if val, ok := response.Header[headers.Location]; ok {
		loc = val[0]
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return tstWebResponse{}, err
	}
	err = response.Body.Close()
	if err != nil {
		return tstWebResponse{}, err
	}
	return tstWebResponse{
		status:      status,
		body:        string(body),
		contentType: ct,
		location:    loc,
	}, nil
}

func tstPerformGet(relativeUrlWithLeadingSlash string, bearerToken string) (tstWebResponse, error) {
	if ts == nil {
		return tstWebResponse{}, errors.New("test web server was not initialized")
	}
	request, err := http.NewRequest(http.MethodGet, ts.URL+relativeUrlWithLeadingSlash, nil)
	if err != nil {
		return tstWebResponse{}, err
	}
	if bearerToken != "" {
		request.Header.Set(headers.Authorization, "Bearer "+bearerToken)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return tstWebResponse{}, err
	}
	return tstWebResponseFromResponse(response)
}
