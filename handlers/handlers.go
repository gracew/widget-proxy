package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gracew/widget-proxy/config"
	"github.com/gracew/widget-proxy/model"
	"github.com/gracew/widget-proxy/parse"
	"github.com/gracew/widget-proxy/store"
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

	// execute beforeSave logic
	customLogic, err := store.CustomLogic(config.CustomLogicPath)
	var createCustomLogic *model.CustomLogic
	for _, el := range customLogic {
		if el.OperationType == model.OperationTypeCreate {
			createCustomLogic = &el
		}
	}

	var parseReq map[string]interface{}

	if createCustomLogic == nil || createCustomLogic.BeforeSave == nil {
		err = json.NewDecoder(r.Body).Decode(&parseReq)
		if err != nil {
			panic(err)
		}
	} else {
		res, err := http.Post(config.CustomLogicUrl + "beforeCreate", "application/json", r.Body)
		if err != nil {
			panic(err)
		}
		err = json.NewDecoder(res.Body).Decode(&parseReq)
		if err != nil {
			panic(err)
		}
	}

	// add createdBy to the original req
	parseReq["createdBy"] = userID

	// delegate to parse
	res, err := parse.CreateObject(parseReq)
	if err != nil {
		panic(err)
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
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

type createdBy struct {
	CreatedBy string `json:"createdBy"`
}

type unauthorized struct {
	Message string `json:"message"`
}
