package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gracew/widget-proxy/config"
	"github.com/gracew/widget-proxy/metrics"
	"github.com/gracew/widget-proxy/model"
	"github.com/pkg/errors"
)

type CreateRes struct {
	CreatedAt string `json:"createdAt"`
	ObjectID  string `json:"objectId"`
}

type ObjectRes = map[string]interface{}

type ListRes struct {
	Results []ObjectRes `json:"results"`
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

func CreateObject(req map[string]interface{}) (*CreateRes, error) {
	start := time.Now()
	res, err := createObject(req)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(model.OperationTypeCreate.String()).Observe(end.Sub(start).Seconds())
	return res, err
}

func createObject(req map[string]interface{}) (*CreateRes, error) {
	parseURL, err := parseURL(fmt.Sprintf("classes/%s", parseClassName()))
	if err != nil {
		return nil, err
	}
	marshaled, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	parseReq, err := http.NewRequest("POST", parseURL, bytes.NewReader(marshaled))
	if err != nil {
		return nil, err
	}
	parseReq.Header.Add("X-Parse-Application-Id", "appId")
	parseReq.Header.Add("Content-type", "application/json")
	client := &http.Client{}
	res, err := client.Do(parseReq)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var parseRes CreateRes
	err = json.Unmarshal(bytes, &parseRes)
	if err != nil {
		return nil, err
	}
	return &parseRes, nil
}

func GetObject(objectID string) (*ObjectRes, error) {
	start := time.Now()
	res, err := getObject(objectID)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(model.OperationTypeRead.String()).Observe(end.Sub(start).Seconds())
	return res, err
}

func getObject(objectID string) (*ObjectRes, error) {
	parseURL, err := parseURL(fmt.Sprintf("classes/%s/%s", parseClassName(), objectID))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", parseURL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-Parse-Application-Id", "appId")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var parseRes ObjectRes
	err = json.Unmarshal(bytes, &parseRes)
	if err != nil {
		return nil, err
	}
	return &parseRes, nil
}

func ListObjects(pageSize string) (*ListRes, error) {
	start := time.Now()
	res, err := listObjects(pageSize)
	end := time.Now()
	metrics.DatabaseSummary.WithLabelValues(model.OperationTypeList.String()).Observe(end.Sub(start).Seconds())
	return res, err
}

func listObjects(pageSize string) (*ListRes, error) {
	parseURL, err := parseURL(fmt.Sprintf("classes/%s", parseClassName()))
	if err != nil {
		return nil, err
	}
	data := "limit=" + pageSize
	req, err := http.NewRequest("GET", parseURL, strings.NewReader(data))
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-Parse-Application-Id", "appId")
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var parseRes ListRes
	err = json.Unmarshal(bytes, &parseRes)
	if err != nil {
		return nil, err
	}

	return &parseRes, nil
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

func parseClassName() string {
	deployID := os.Getenv("DEPLOY_ID")
	// parse class names cannot start with numbers or contain dashes
	return fmt.Sprintf("w%s", strings.Replace(deployID, "-", "", -1))

}
