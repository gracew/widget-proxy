package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gracew/widget-proxy/config"
	"github.com/gracew/widget-proxy/metrics"
	"github.com/gracew/widget-proxy/model"
	"github.com/gracew/widget-proxy/parse"
	"github.com/gracew/widget-proxy/store"
	"github.com/pkg/errors"
)

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// get the userId
	parseToken := r.Header["X-Parse-Session-Token"][0]
	userID, err := parse.GetUserId(parseToken)
	if err != nil {
		panic(err)
	}

	customLogic, err := store.CustomLogic(config.CustomLogicPath)
	var createCustomLogic *model.CustomLogic
	for _, el := range customLogic {
		if el.OperationType == model.OperationTypeCreate {
			createCustomLogic = &el
		}
	}

	parseReq, err := applyBeforeCustomLogic(r, createCustomLogic)
	if err != nil {
		metrics.CustomLogicErrors.WithLabelValues(model.OperationTypeCreate.String(), "before").Inc()
		panic(err)
	}

	// add createdBy to the original req
	parseReq["createdBy"] = userID

	// delegate to parse
	res, err := parse.CreateObject(parseReq)
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(model.OperationTypeCreate.String()).Inc()
		panic(err)
	}

	err = applyAfterCustomLogic(w, res, createCustomLogic)
	if err != nil {
		metrics.CustomLogicErrors.WithLabelValues(model.OperationTypeCreate.String(), "after").Inc()
		panic(err)
	}
}

func ReadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// get the userId
	parseToken := r.Header["X-Parse-Session-Token"][0]
	userID, err := parse.GetUserId(parseToken)
	if err != nil {
		panic(err)
	}

	// delegate to parse
	vars := mux.Vars(r)
	res, err := parse.GetObject(vars["id"])
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(model.OperationTypeRead.String()).Inc()
		panic(err)
	}

	// fetch the authorization policy
	// TODO(gracew): parallelize some of these requests
	auth, err := store.Auth(config.AuthPath)
	if err != nil {
		panic(err)
	}

	if auth.ReadPolicy.Type == model.AuthPolicyTypeCreatedBy {
		if userID != (*res)["createdBy"] {
			json.NewEncoder(w).Encode(&unauthorized{Message: "unauthorized"})
			return
		}
	}
	// TODO(gracew): support other authz policies

	json.NewEncoder(w).Encode(&res)
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// get the userId
	parseToken := r.Header["X-Parse-Session-Token"][0]
	userID, err := parse.GetUserId(parseToken)
	if err != nil {
		panic(err)
	}

	// delegate to parse
	pageSizes, ok := r.URL.Query()["pageSize"]
	pageSize := "100"
	if ok && len(pageSizes[0]) >= 1 {
		pageSize = pageSizes[0]
	}
	res, err := parse.ListObjects(pageSize)
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(model.OperationTypeList.String()).Inc()
		panic(err)
	}

	// fetch the authorization policy
	// TODO(gracew): parallelize some of these requests
	auth, err := store.Auth(config.AuthPath)
	if err != nil {
		panic(err)
	}

	var filtered []parse.ObjectRes
	if auth.ReadPolicy.Type == model.AuthPolicyTypeCreatedBy {
		for i := 0; i < len(res.Results); i++ {
			if userID == res.Results[i]["createdBy"] {
				filtered = append(filtered, res.Results[i])
			}
		}
	}
	// TODO(gracew): support other authz policies

	json.NewEncoder(w).Encode(&parse.ListRes{Results: filtered})
}

type unauthorized struct {
	Message string `json:"message"`
}

func applyBeforeCustomLogic(r *http.Request, customLogic *model.CustomLogic) (map[string]interface{}, error) {
	start := time.Now()
	res, err := applyBeforeCustomLogicUninstrumented(r, customLogic)
	end := time.Now()
	metrics.CustomLogicSummary.WithLabelValues(customLogic.OperationType.String(), "before").Observe(end.Sub(start).Seconds())
	return res, err
}

func applyBeforeCustomLogicUninstrumented(r *http.Request, customLogic *model.CustomLogic) (map[string]interface{}, error) {
	var result map[string]interface{}

	if customLogic == nil || customLogic.Before == nil {
		err := json.NewDecoder(r.Body).Decode(&result)
		if err != nil {
			return nil, errors.Wrap(err, "could not decode request as json")
		}
		return result, nil
	}

	res, err := http.Post(config.CustomLogicUrl + "beforeCreate", "application/json", r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "request to custom logic endpoint failed")
	}

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode response from custom logic endpoint")
	}
	return result, nil
}

func applyAfterCustomLogic(w http.ResponseWriter, input *parse.CreateRes, customLogic *model.CustomLogic) error {
	start := time.Now()
	err := applyAfterCustomLogicUninstrumented(w, input, customLogic)
	end := time.Now()
	metrics.CustomLogicSummary.WithLabelValues(customLogic.OperationType.String(), "after").Observe(end.Sub(start).Seconds())
	return err
}

func applyAfterCustomLogicUninstrumented(w http.ResponseWriter, input *parse.CreateRes, customLogic *model.CustomLogic) error {
	if customLogic == nil || customLogic.After == nil {
		err := json.NewEncoder(w).Encode(input)
		if err != nil {
			return errors.Wrap(err, "could not encode response")
		}
		return nil
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return errors.Wrap(err, "could not marshal custom logic input")
	}
	afterRes, err := http.Post(config.CustomLogicUrl + "afterCreate", "application/json", bytes.NewReader(inputBytes))
	if err != nil {
		return errors.Wrap(err, "request to custom logic endpoint failed")
	}
	err = json.NewEncoder(w).Encode(afterRes)
	if err != nil {
		return errors.Wrap(err, "could not encode response")
	}
	return nil
}
