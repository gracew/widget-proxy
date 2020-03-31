package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gracew/widget-proxy/config"
	"github.com/gracew/widget-proxy/generated"
	"github.com/gracew/widget-proxy/metrics"
	"github.com/gracew/widget-proxy/model"
	"github.com/gracew/widget-proxy/store"
	"github.com/gracew/widget-proxy/user"
	"github.com/pkg/errors"
)

type Handlers struct {
	Store       store.Store
	Auth        *model.Auth
	CustomLogic *model.AllCustomLogic
}

func (h Handlers) CreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// get the userId
	parseToken := r.Header["X-Parse-Session-Token"][0]
	userID, err := user.GetUserId(parseToken)
	if err != nil {
		panic(err)
	}

	bytes, err := applyBeforeCustomLogic(r, h.CustomLogic.Create, metrics.CREATE)
	if err != nil {
		panic(err)
	}

	// delegate to db
	var obj *generated.Object
	err = json.Unmarshal(bytes, obj)
	if err != nil {
		panic(err)
	}
	obj.CreatedBy = userID

	res, err := h.Store.CreateObject(obj)
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(metrics.CREATE).Inc()
		panic(err)
	}

	err = applyAfterCustomLogic(w, res, h.CustomLogic.Create, metrics.CREATE)
	if err != nil {
		panic(err)
	}
}

func (h Handlers) ReadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// get the userId
	parseToken := r.Header["X-Parse-Session-Token"][0]
	userID, err := user.GetUserId(parseToken)
	if err != nil {
		panic(err)
	}

	// delegate to db
	vars := mux.Vars(r)
	res, err := h.Store.GetObject(vars["id"])
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(metrics.READ).Inc()
		panic(err)
	}

	if h.Auth.Read.Type == model.AuthPolicyTypeCreatedBy {
		if userID != (*res).CreatedBy {
			json.NewEncoder(w).Encode(&unauthorized{Message: "unauthorized"})
			return
		}
	}
	// TODO(gracew): support other authz policies

	json.NewEncoder(w).Encode(&res)
}

func (h Handlers) ListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// get the userId
	parseToken := r.Header["X-Parse-Session-Token"][0]
	userID, err := user.GetUserId(parseToken)
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
	res, err := h.Store.ListObjects(pageSize)
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(metrics.LIST).Inc()
		panic(err)
	}

	// fetch the authorization policy
	// TODO(gracew): parallelize some of these requests
	auth, err := config.Auth(config.AuthPath)
	if err != nil {
		panic(err)
	}

	var filtered []generated.Object
	if auth.Read.Type == model.AuthPolicyTypeCreatedBy {
		for i := 0; i < len(res); i++ {
			if userID == res[i].CreatedBy {
				filtered = append(filtered, res[i])
			}
		}
	}
	// TODO(gracew): support other authz policies

	json.NewEncoder(w).Encode(filtered)
}

// TODO(gracew): authz
func (h Handlers) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	actionName := vars["action"]

	bytes, err := applyBeforeCustomLogic(r, h.CustomLogic.Update[actionName], actionName)
	if err != nil {
		panic(err)
	}

	// delegate to db
	var obj *generated.Object
	err = json.Unmarshal(bytes, obj)
	if err != nil {
		panic(err)
	}
	obj.ID = vars["id"]

	res, err := h.Store.UpdateObject(obj, actionName)
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(actionName).Inc()
		panic(err)
	}

	err = applyAfterCustomLogic(w, res, h.CustomLogic.Update[actionName], actionName)
	if err != nil {
		panic(err)
	}
}

func (h Handlers) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// get the userId
	parseToken := r.Header["X-Parse-Session-Token"][0]
	userID, err := user.GetUserId(parseToken)
	if err != nil {
		panic(err)
	}

	// fetch object first, and enforce authz
	vars := mux.Vars(r)
	res, err := h.Store.GetObject(vars["id"])
	if h.Auth.Delete.Type == model.AuthPolicyTypeCreatedBy {
		if userID != (*res).CreatedBy {
			json.NewEncoder(w).Encode(&unauthorized{Message: "unauthorized"})
			return
		}
	}

	_, err = applyBeforeCustomLogic(r, h.CustomLogic.Delete, metrics.DELETE)
	if err != nil {
		panic(err)
	}

	err = h.Store.DeleteObject(vars["id"])
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(metrics.DELETE).Inc()
		panic(err)
	}

	err = applyAfterCustomLogic(w, res, h.CustomLogic.Delete, metrics.DELETE)
	if err != nil {
		panic(err)
	}
}

type unauthorized struct {
	Message string `json:"message"`
}

func applyBeforeCustomLogic(r *http.Request, customLogic *model.CustomLogic, operation string) ([]byte, error) {
	if customLogic == nil || customLogic.Before == nil {
		ret, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, errors.Wrap(err, "could not read request body")
		}
		return ret, nil
	}

	start := time.Now()
	res, err := http.Post(config.CustomLogicURL+"before"+operation, "application/json", r.Body)
	if err != nil {
		metrics.CustomLogicErrors.WithLabelValues(operation, "before").Inc()
		return nil, errors.Wrap(err, "request to custom logic endpoint failed")
	}
	end := time.Now()
	metrics.CustomLogicSummary.WithLabelValues(operation, "before").Observe(end.Sub(start).Seconds())

	ret, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read custom logic response body")
	}

	return ret, nil
}

func applyAfterCustomLogic(w http.ResponseWriter, input *generated.Object, customLogic *model.CustomLogic, operation string) error {
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
	afterRes, err := http.Post(config.CustomLogicURL+"after"+operation, "application/json", bytes.NewReader(inputBytes))
	if err != nil {
		metrics.CustomLogicErrors.WithLabelValues(operation, "after").Inc()
		return errors.Wrap(err, "request to custom logic endpoint failed")
	}
	end := time.Now()
	metrics.CustomLogicSummary.WithLabelValues(operation, "after").Observe(end.Sub(start).Seconds())

	err = json.NewEncoder(w).Encode(afterRes)
	if err != nil {
		return errors.Wrap(err, "could not encode response")
	}
	return nil
}
