package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gracew/widget-proxy/generated"
	"github.com/gracew/widget-proxy/metrics"
	"github.com/gracew/widget-proxy/model"
	"github.com/gracew/widget-proxy/store"
	"github.com/gracew/widget-proxy/user"
	"github.com/pkg/errors"
)

type Handlers struct {
	Store               store.Store
	Auth                model.Auth
	Authenticator       user.Authenticator
	CustomLogic         model.AllCustomLogic
	CustomLogicExecutor CustomLogicExecutor
}

func (h Handlers) CreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// get the userId
	userID, err := h.Authenticator.GetUserId(r.Header)
	if err != nil {
		panic(err)
	}

	obj, err := h.applyBeforeCustomLogic(r.Body, h.CustomLogic.Create, metrics.CREATE)
	if err != nil {
		panic(err)
	}

	// delegate to db
	obj.CreatedBy = userID
	res, err := h.Store.CreateObject(obj)
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(metrics.CREATE).Inc()
		panic(err)
	}

	err = h.applyAfterCustomLogic(w, res, h.CustomLogic.Create, metrics.CREATE)
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
	userID, err := h.Authenticator.GetUserId(r.Header)
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
			h.unauthorizedResponse(w)
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
	userID, err := h.Authenticator.GetUserId(r.Header)
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

	var filtered []generated.Object
	if h.Auth.Read.Type == model.AuthPolicyTypeCreatedBy {
		for i := 0; i < len(res); i++ {
			if userID == res[i].CreatedBy {
				filtered = append(filtered, res[i])
			}
		}
	}
	// TODO(gracew): support other authz policies

	json.NewEncoder(w).Encode(filtered)
}

func (h Handlers) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// get the userId
	userID, err := h.Authenticator.GetUserId(r.Header)
	if err != nil {
		panic(err)
	}

	// fetch object first, and enforce authz
	vars := mux.Vars(r)
	id := vars["id"]
	actionName := vars["action"]
	res, err := h.Store.GetObject(id)
	if h.Auth.Update[actionName].Type == model.AuthPolicyTypeCreatedBy {
		if userID != (*res).CreatedBy {
			h.unauthorizedResponse(w)
			return
		}
	}

	obj, err := h.applyBeforeCustomLogic(r.Body, h.CustomLogic.Update[actionName], actionName)
	if err != nil {
		panic(err)
	}

	// delegate to db
	obj.ID = id
	res, err = h.Store.UpdateObject(obj, actionName)
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(actionName).Inc()
		panic(err)
	}

	err = h.applyAfterCustomLogic(w, res, h.CustomLogic.Update[actionName], actionName)
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
	userID, err := h.Authenticator.GetUserId(r.Header)
	if err != nil {
		panic(err)
	}

	// fetch object first, and enforce authz
	vars := mux.Vars(r)
	obj, err := h.Store.GetObject(vars["id"])
	if h.Auth.Delete.Type == model.AuthPolicyTypeCreatedBy {
		if userID != (*obj).CreatedBy {
			h.unauthorizedResponse(w)
			return
		}
	}

	objBytes, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	_, err = h.applyBeforeCustomLogic(bytes.NewReader(objBytes), h.CustomLogic.Delete, metrics.DELETE)
	if err != nil {
		panic(err)
	}

	err = h.Store.DeleteObject(vars["id"])
	if err != nil {
		metrics.DatabaseErrors.WithLabelValues(metrics.DELETE).Inc()
		panic(err)
	}

	err = h.applyAfterCustomLogic(w, obj, h.CustomLogic.Delete, metrics.DELETE)
	if err != nil {
		panic(err)
	}
}

type errorResponse struct {
	Message string `json:"message"`
}

func (h Handlers) unauthorizedResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(&errorResponse{Message: "unauthorized"})

}

func (h Handlers) applyBeforeCustomLogic(reader io.Reader, customLogic *model.CustomLogic, operation string) (*generated.Object, error) {
	var obj generated.Object
	if customLogic == nil || customLogic.Before == nil {
		err := json.NewDecoder(reader).Decode(&obj)
		if err != nil {
			return nil, errors.Wrap(err, "could not read request body")
		}
		return &obj, nil
	}

	res, err := h.CustomLogicExecutor.Execute(reader, "before", operation)
	if err != nil {
		return nil, errors.Wrap(err, "request to custom logic endpoint failed")
	}

	err = json.NewDecoder(res.Body).Decode(&obj)
	if err != nil {
		return nil, errors.Wrap(err, "could not read custom logic response body")
	}

	return &obj, nil
}

func (h Handlers) applyAfterCustomLogic(w http.ResponseWriter, input *generated.Object, customLogic *model.CustomLogic, operation string) error {
	if customLogic == nil || customLogic.After == nil {
		if operation == metrics.DELETE {
			w.WriteHeader(http.StatusNoContent)
			return nil
		}

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

	res, err := h.CustomLogicExecutor.Execute(bytes.NewReader(inputBytes), "after", operation)
	if err != nil {
		errors.Wrap(err, "request to custom logic endpoint failed")
	}

	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "could not read response from custom logic endpoint")
	}

	_, err = w.Write(resBytes)
	if err != nil {
		return errors.Wrap(err, "could not write response")
	}

	return nil
}
