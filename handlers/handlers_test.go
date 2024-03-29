// +build test

package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/gracew/widget-proxy/generated"
	"github.com/gracew/widget-proxy/metrics"
	"github.com/gracew/widget-proxy/mocks"
	"github.com/gracew/widget-proxy/model"
	"github.com/gracew/widget-proxy/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HandlersTestSuite struct {
	suite.Suite
	store         *mocks.MockStore
	executor      *mocks.MockCustomLogicExecutor
	authenticator *mocks.MockAuthenticator
}

var h Handlers

func (suite *HandlersTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())
	defer mockCtrl.Finish()
	suite.store = mocks.NewMockStore(mockCtrl)
	suite.executor = mocks.NewMockCustomLogicExecutor(mockCtrl)
	suite.authenticator = mocks.NewMockAuthenticator(mockCtrl)
	suite.authenticator.EXPECT().GetUserId(gomock.Any()).Return("userID", nil).AnyTimes()
	h = Handlers{
		Store:               suite.store,
		CustomLogic:         model.AllCustomLogic{},
		CustomLogicExecutor: suite.executor,
		Authenticator:       suite.authenticator,
		Auth: model.Auth{
			Read: &model.AuthPolicy{Type: model.AuthPolicyTypeCreatedBy},
			Update: map[string]*model.AuthPolicy{
				"action": &model.AuthPolicy{Type: model.AuthPolicyTypeCreatedBy},
			},
			Delete: &model.AuthPolicy{Type: model.AuthPolicyTypeCreatedBy},
		},
	}
}

func (suite *HandlersTestSuite) TestCreate() {
	input := generated.Object{ID: "1"}
	storeInput := generated.Object{ID: "1", CreatedBy: "userID"}
	storeOutput := generated.Object{ID: "2"}

	suite.executor.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	suite.store.EXPECT().CreateObject(&storeInput).Return(&storeOutput, nil)

	rr := httptest.NewRecorder()
	h.CreateHandler(rr, suite.request(input))

	assert.Equal(suite.T(), storeOutput, suite.decode(rr.Body))
}

func (suite *HandlersTestSuite) TestCreateCustomLogic() {
	customLogic := "something"
	h.CustomLogic = model.AllCustomLogic{Create: &model.CustomLogic{Before: &customLogic, After: &customLogic}}

	input := generated.Object{ID: "1"}
	beforeCustomLogicOutput := generated.Object{ID: "2"}
	storeInput := generated.Object{ID: "2", CreatedBy: "userID"}
	storeOutput := generated.Object{ID: "3"}
	afterCustomLogicOutput := generated.Object{ID: "4"}

	suite.executor.EXPECT().Execute(gomock.Any(), "before", metrics.CREATE).
		Times(1).
		Return(suite.response(beforeCustomLogicOutput), nil)
	suite.store.EXPECT().CreateObject(&storeInput).Return(&storeOutput, nil)
	suite.executor.EXPECT().Execute(gomock.Any(), "after", metrics.CREATE).
		Times(1).
		Return(suite.response(afterCustomLogicOutput), nil)

	rr := httptest.NewRecorder()
	h.CreateHandler(rr, suite.request(input))

	assert.Equal(suite.T(), afterCustomLogicOutput, suite.decode(rr.Body))
}

func (suite *HandlersTestSuite) TestRead() {
	storeOutput := generated.Object{ID: "1", CreatedBy: "userID"}
	suite.store.EXPECT().GetObject("1").Return(&storeOutput, nil)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(suite.T(), err)
	h.ReadHandler(rr, mux.SetURLVars(req, map[string]string{"id": "1"}))

	assert.Equal(suite.T(), storeOutput, suite.decode(rr.Body))
}

func (suite *HandlersTestSuite) TestReadUnauthorized() {
	storeOutput := generated.Object{ID: "1", CreatedBy: "anotherUserID"}
	suite.store.EXPECT().GetObject("1").Return(&storeOutput, nil)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(suite.T(), err)
	h.ReadHandler(rr, mux.SetURLVars(req, map[string]string{"id": "1"}))

	assert.Equal(suite.T(), http.StatusForbidden, rr.Result().StatusCode)
}

func (suite *HandlersTestSuite) TestListDefaultPageSize() {
	storeOutput := []generated.Object{generated.Object{ID: "1", CreatedBy: "userID"}}
	suite.store.EXPECT().ListObjects(100, nil).Return(storeOutput, nil)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(suite.T(), err)
	h.ListHandler(rr, req)

	var res []generated.Object
	err = json.NewDecoder(rr.Body).Decode(&res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), storeOutput, res)
}

func (suite *HandlersTestSuite) TestListPageSizeQuery() {
	storeOutput := []generated.Object{generated.Object{ID: "1", CreatedBy: "userID"}}
	suite.store.EXPECT().ListObjects(50, nil).Return(storeOutput, nil)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(suite.T(), err)
	q := req.URL.Query()
	q.Add("pageSize", "50")
	req.URL.RawQuery = q.Encode()
	h.ListHandler(rr, req)

	var res []generated.Object
	err = json.NewDecoder(rr.Body).Decode(&res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), storeOutput, res)
}

func (suite *HandlersTestSuite) TestListUnauthorized() {
	storeOutput := []generated.Object{generated.Object{ID: "1", CreatedBy: "anotherUserID"}}
	suite.store.EXPECT().ListObjects(100, nil).Return(storeOutput, nil)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(suite.T(), err)
	h.ListHandler(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Result().StatusCode)
	var res []generated.Object
	err = json.NewDecoder(rr.Body).Decode(&res)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), res)
}

func (suite *HandlersTestSuite) TestListFilter() {
	storeOutput := []generated.Object{generated.Object{ID: "1", CreatedBy: "userID"}}
	suite.store.EXPECT().ListObjects(100, &store.Filter{Field: "key", Value: "value"}).Return(storeOutput, nil)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(suite.T(), err)
	q := req.URL.Query()
	q.Add("key", "value")
	q.Add("key", "anotherValue")
	req.URL.RawQuery = q.Encode()
	h.ListHandler(rr, req)

	var res []generated.Object
	err = json.NewDecoder(rr.Body).Decode(&res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), storeOutput, res)
}

func (suite *HandlersTestSuite) TestUpdate() {
	input := generated.Object{ID: "1"}
	getOutput := generated.Object{ID: "1", CreatedBy: "userID"}
	storeOutput := generated.Object{ID: "2"}

	suite.store.EXPECT().GetObject("1").Return(&getOutput, nil)
	suite.executor.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	suite.store.EXPECT().UpdateObject(&input, "action").Return(&storeOutput, nil)

	rr := httptest.NewRecorder()
	h.UpdateHandler(rr, mux.SetURLVars(suite.request(input), map[string]string{"id": "1", "action": "action"}))

	assert.Equal(suite.T(), storeOutput, suite.decode(rr.Body))
}

func (suite *HandlersTestSuite) TestUpdateUnauthorized() {
	input := generated.Object{ID: "1"}
	getOutput := generated.Object{ID: "1", CreatedBy: "anotherUserID"}

	suite.store.EXPECT().GetObject("1").Return(&getOutput, nil)
	suite.store.EXPECT().UpdateObject(gomock.Any(), gomock.Any()).Times(0)

	rr := httptest.NewRecorder()
	h.UpdateHandler(rr, mux.SetURLVars(suite.request(input), map[string]string{"id": "1", "action": "action"}))

	assert.Equal(suite.T(), http.StatusForbidden, rr.Result().StatusCode)
}

func (suite *HandlersTestSuite) TestUpdateCustomLogic() {
	customLogic := "something"
	h.CustomLogic = model.AllCustomLogic{
		Update: map[string]*model.CustomLogic{
			"action": &model.CustomLogic{Before: &customLogic, After: &customLogic},
		},
	}

	input := generated.Object{ID: "1"}
	getOutput := generated.Object{ID: "1", CreatedBy: "userID"}
	beforeCustomLogicOutput := generated.Object{ID: "1", Test: "test"}
	storeOutput := generated.Object{ID: "2"}
	afterCustomLogicOutput := generated.Object{ID: "3"}

	suite.store.EXPECT().GetObject("1").Return(&getOutput, nil)
	suite.executor.EXPECT().Execute(gomock.Any(), "before", "action").
		Times(1).
		Return(suite.response(beforeCustomLogicOutput), nil)
	suite.store.EXPECT().UpdateObject(&beforeCustomLogicOutput, "action").Return(&storeOutput, nil)
	suite.executor.EXPECT().Execute(gomock.Any(), "after", "action").
		Times(1).
		Return(suite.response(afterCustomLogicOutput), nil)

	rr := httptest.NewRecorder()
	h.UpdateHandler(rr, mux.SetURLVars(suite.request(input), map[string]string{"id": "1", "action": "action"}))

	assert.Equal(suite.T(), afterCustomLogicOutput, suite.decode(rr.Body))
}

func (suite *HandlersTestSuite) TestDelete() {
	getOutput := generated.Object{ID: "1", CreatedBy: "userID"}

	suite.executor.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	suite.store.EXPECT().GetObject("1").Return(&getOutput, nil)
	suite.store.EXPECT().DeleteObject("1").Return(nil)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "", nil)
	assert.NoError(suite.T(), err)
	h.DeleteHandler(rr, mux.SetURLVars(req, map[string]string{"id": "1"}))

	assert.Equal(suite.T(), http.StatusNoContent, rr.Result().StatusCode)
}

func (suite *HandlersTestSuite) TestDeleteUnauthorized() {
	storeOutput := generated.Object{ID: "1", CreatedBy: "anotherUserID"}
	suite.store.EXPECT().GetObject("1").Return(&storeOutput, nil)
	suite.store.EXPECT().DeleteObject(gomock.Any).Times(0)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "", nil)
	assert.NoError(suite.T(), err)
	h.DeleteHandler(rr, mux.SetURLVars(req, map[string]string{"id": "1"}))

	assert.Equal(suite.T(), http.StatusForbidden, rr.Result().StatusCode)
}

func (suite *HandlersTestSuite) TestDeleteCustomLogic() {
	customLogic := "something"
	h.CustomLogic = model.AllCustomLogic{Delete: &model.CustomLogic{Before: &customLogic, After: &customLogic}}

	getOutput := generated.Object{ID: "1", CreatedBy: "userID"}
	afterCustomLogicOutput := generated.Object{ID: "2"}

	suite.store.EXPECT().GetObject("1").Return(&getOutput, nil)
	suite.executor.EXPECT().Execute(gomock.Any(), "before", metrics.DELETE).
		Times(1).
		Return(suite.response(getOutput), nil)
	suite.store.EXPECT().DeleteObject("1").Return(nil)
	suite.executor.EXPECT().Execute(gomock.Any(), "after", metrics.DELETE).
		Times(1).
		Return(suite.response(afterCustomLogicOutput), nil)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "", nil)
	assert.NoError(suite.T(), err)
	h.DeleteHandler(rr, mux.SetURLVars(req, map[string]string{"id": "1"}))

	assert.Equal(suite.T(), http.StatusOK, rr.Result().StatusCode)
	assert.Equal(suite.T(), afterCustomLogicOutput, suite.decode(rr.Body))
}

func (suite *HandlersTestSuite) request(obj generated.Object) *http.Request {
	req, err := http.NewRequest("POST", "", suite.encode(obj))
	assert.NoError(suite.T(), err)
	return req
}

func (suite *HandlersTestSuite) response(obj generated.Object) *http.Response {
	return &http.Response{Body: ioutil.NopCloser(suite.encode(obj))}
}

func (suite *HandlersTestSuite) encode(obj generated.Object) *bytes.Reader {
	bs, err := json.Marshal(obj)
	assert.NoError(suite.T(), err)
	return bytes.NewReader(bs)
}

func (suite *HandlersTestSuite) decode(body io.Reader) generated.Object {
	var res generated.Object
	err := json.NewDecoder(body).Decode(&res)
	assert.NoError(suite.T(), err)
	return res
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}
