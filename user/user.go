package user

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gracew/widget-proxy/config"
	"github.com/pkg/errors"
)

type CreateRes struct {
	CreatedAt string `json:"createdAt"`
	ObjectID  string `json:"objectId"`
}

func GetUserId(parseToken string) (string, error) {
	parseURL, err := parseURL("users/me")
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("GET", parseURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Parse-Application-Id", "appId")
	req.Header.Add("X-Parse-Session-Token", parseToken)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch information for current user")
	}
	var parseRes CreateRes
	err = json.NewDecoder(res.Body).Decode(&parseRes)
	if err != nil {
		return "", errors.Wrap(err, "failed to json decode response")
	}
	return parseRes.ObjectID, nil
}

func parseURL(path string) (string, error) {
	parseURL, err := url.Parse(config.ParseURL)
	if err != nil {
		return "", errors.Wrapf(err, "could not parse PARSE_URL as URL: %s", config.ParseURL)
	}
	pathURL, err := url.Parse(path)
	if err != nil {
		return "", errors.Wrapf(err, "could not parse path as URL: %s", path)
	}
	return parseURL.ResolveReference(pathURL).String(), nil
}
