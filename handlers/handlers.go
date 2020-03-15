package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/gracew/widget-proxy/config"
	"github.com/gracew/widget-proxy/generated"
	"github.com/gracew/widget-proxy/metrics"
	"github.com/gracew/widget-proxy/model"
	"github.com/gracew/widget-proxy/store"
	"github.com/gracew/widget-proxy/user"
	"github.com/pkg/errors"
)

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// get the userId
	dbToken := r.Header["X-Parse-Session-Token"][0]
	userID, err := user.GetUserId(dbToken)
	if err != nil {
		panic(err)
	}

	customLogic, err := config.CustomLogic(config.CustomLogicPath)
	var createCustomLogic *model.CustomLogic
	for _, el := range customLogic {
		if el.OperationType == model.OperationTypeCreate {
			createCustomLogic = &el
		}
	}

	bytes, err := applyBeforeCustomLogic(r, createCustomLogic)
	if err != nil {
		panic(err)
	}

	// delegate to db
	db := pg.Connect(&pg.Options{User: "postgres", Addr: "localhost:5433"})
	defer db.Close()
	s := store.Store{DB: db}
	res, err := s.CreateObject(bytes, userID)
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(model.OperationTypeCreate.String()).Inc()
		panic(err)
	}

	err = applyAfterCustomLogic(w, res, createCustomLogic)
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
	dbToken := r.Header["X-Parse-Session-Token"][0]
	userID, err := user.GetUserId(dbToken)
	if err != nil {
		panic(err)
	}

	// delegate to db
	vars := mux.Vars(r)
	db := pg.Connect(&pg.Options{User: "postgres", Addr: "localhost:5433"})
	defer db.Close()
	s := store.Store{DB: db}
	res, err := s.GetObject(vars["id"])
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(model.OperationTypeRead.String()).Inc()
		panic(err)
	}

	// fetch the authorization policy
	// TODO(gracew): parallelize some of these requests
	auth, err := config.Auth(config.AuthPath)
	if err != nil {
		panic(err)
	}

	if auth.ReadPolicy.Type == model.AuthPolicyTypeCreatedBy {
		if userID != (*res).CreatedBy {
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
	dbToken := r.Header["X-Parse-Session-Token"][0]
	userID, err := user.GetUserId(dbToken)
	if err != nil {
		panic(err)
	}

	// delegate to db
	pageSizes, ok := r.URL.Query()["pageSize"]
	pageSize := 100
	if ok && len(pageSizes[0]) >= 1 {
		pageSize, err = strconv.Atoi(pageSizes[0])
		if err != nil {
			panic(err)
		}
	}
	db := pg.Connect(&pg.Options{User: "postgres", Addr: "localhost:5433"})
	defer db.Close()
	s := store.Store{DB: db}
	res, err := s.ListObjects(pageSize)
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(model.OperationTypeList.String()).Inc()
		panic(err)
	}

	// fetch the authorization policy
	// TODO(gracew): parallelize some of these requests
	auth, err := config.Auth(config.AuthPath)
	if err != nil {
		panic(err)
	}

	var filtered []generated.Object
	if auth.ReadPolicy.Type == model.AuthPolicyTypeCreatedBy {
		for i := 0; i < len(res); i++ {
			if userID == res[i].CreatedBy {
				filtered = append(filtered, res[i])
			}
		}
	}
	// TODO(gracew): support other authz policies

	json.NewEncoder(w).Encode(filtered)
}

type unauthorized struct {
	Message string `json:"message"`
}

func applyBeforeCustomLogic(r *http.Request, customLogic *model.CustomLogic) ([]byte, error) {
	if customLogic == nil || customLogic.Before == nil {
		ret, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, errors.Wrap(err, "could not read request body")
		}
		return ret, nil
	}

	start := time.Now()
	res, err := http.Post(config.CustomLogicUrl + "beforeCreate", "application/json", r.Body)
	if err != nil {
		metrics.CustomLogicErrors.WithLabelValues(model.OperationTypeCreate.String(), "before").Inc()
		return nil, errors.Wrap(err, "request to custom logic endpoint failed")
	}
	end := time.Now()
	metrics.CustomLogicSummary.WithLabelValues(customLogic.OperationType.String(), "before").Observe(end.Sub(start).Seconds())

	ret, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read custom logic response body")
	}

	return ret, nil
}

func applyAfterCustomLogic(w http.ResponseWriter, input *generated.Object, customLogic *model.CustomLogic) error {
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

	start := time.Now()
	afterRes, err := http.Post(config.CustomLogicUrl + "afterCreate", "application/json", bytes.NewReader(inputBytes))
	if err != nil {
		metrics.CustomLogicErrors.WithLabelValues(model.OperationTypeCreate.String(), "after").Inc()
		return errors.Wrap(err, "request to custom logic endpoint failed")
	}
	end := time.Now()
	metrics.CustomLogicSummary.WithLabelValues(customLogic.OperationType.String(), "after").Observe(end.Sub(start).Seconds())

	err = json.NewEncoder(w).Encode(afterRes)
	if err != nil {
		return errors.Wrap(err, "could not encode response")
	}
	return nil
}
