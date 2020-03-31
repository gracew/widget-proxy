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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HandlersTestSuite struct {
	suite.Suite
	store         *mocks.MockStore
	caller        *mocks.MockCustomLogicCaller
	authenticator *mocks.MockAuthenticator
}

var auth = model.Auth{
	Read: &model.AuthPolicy{Type: model.AuthPolicyTypeCreatedBy},
	Delete: &model.AuthPolicy{Type: model.AuthPolicyTypeCreatedBy},
}

func (suite *HandlersTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())
	suite.store = mocks.NewMockStore(mockCtrl)
	suite.caller = mocks.NewMockCustomLogicCaller(mockCtrl)
	suite.authenticator = mocks.NewMockAuthenticator(mockCtrl)
	suite.authenticator.EXPECT().GetUserId(gomock.Any()).Return("userID", nil)
}

func (suite *HandlersTestSuite) TestCreate() {
	h := Handlers{
		Store:             suite.store,
		CustomLogic:       model.AllCustomLogic{},
		CustomLogicCaller: suite.caller,
		Authenticator:     suite.authenticator,
	}

	input := generated.Object{ID: "1"}
	storeInput := generated.Object{ID: "1", CreatedBy: "userID"}
	storeOutput := generated.Object{ID: "2"}

	suite.caller.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	suite.store.EXPECT().CreateObject(&storeInput).Return(&storeOutput, nil)

	rr := httptest.NewRecorder()
	h.CreateHandler(rr, suite.request(input))

	assert.Equal(suite.T(), storeOutput, suite.decode(rr.Body))
}

func (suite *HandlersTestSuite) TestCreateCustomLogic() {
	customLogic := "something"
	h := Handlers{
		Store:             suite.store,
		CustomLogic:       model.AllCustomLogic{Create: &model.CustomLogic{Before: &customLogic, After: &customLogic}},
		CustomLogicCaller: suite.caller,
		Authenticator:     suite.authenticator,
	}

	input := generated.Object{ID: "1"}
	beforeCustomLogicOutput := generated.Object{ID: "2"}
	storeInput := generated.Object{ID: "2", CreatedBy: "userID"}
	storeOutput := generated.Object{ID: "3"}
	afterCustomLogicOutput := generated.Object{ID: "4"}

	suite.caller.EXPECT().Call(gomock.Any(), "before", metrics.CREATE).
		Times(1).
		Return(suite.response(beforeCustomLogicOutput), nil)
	suite.store.EXPECT().CreateObject(&storeInput).Return(&storeOutput, nil)
	suite.caller.EXPECT().Call(gomock.Any(), "after", metrics.CREATE).
		Times(1).
		Return(suite.response(afterCustomLogicOutput), nil)

	rr := httptest.NewRecorder()
	h.CreateHandler(rr, suite.request(input))

	assert.Equal(suite.T(), afterCustomLogicOutput, suite.decode(rr.Body))
}

func (suite *HandlersTestSuite) TestRead() {
	h := Handlers{Store: suite.store, CustomLogicCaller: suite.caller, Authenticator: suite.authenticator, Auth: auth}

	input := generated.Object{ID: "1"}
	storeOutput := generated.Object{ID: "1", CreatedBy: "userID"}

	suite.store.EXPECT().GetObject(input.ID).Return(&storeOutput, nil)

	rr := httptest.NewRecorder()
	h.ReadHandler(rr, mux.SetURLVars(suite.request(input), map[string]string{"id": "1"}))

	assert.Equal(suite.T(), storeOutput, suite.decode(rr.Body))
}

func (suite *HandlersTestSuite) TestListDefaultPageSize() {
	h := Handlers{Store: suite.store, CustomLogicCaller: suite.caller, Authenticator: suite.authenticator, Auth: auth}

	input := generated.Object{ID: "1"}
	storeOutput := []generated.Object{generated.Object{ID: "1", CreatedBy: "userID"}}

	suite.store.EXPECT().ListObjects(100).Return(storeOutput, nil)

	rr := httptest.NewRecorder()
	h.ListHandler(rr, suite.request(input))

	var res []generated.Object
	err := json.NewDecoder(rr.Body).Decode(&res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), storeOutput, res)
}

func (suite *HandlersTestSuite) TestListPageSizeQuery() {
	h := Handlers{Store: suite.store, CustomLogicCaller: suite.caller, Authenticator: suite.authenticator, Auth: auth}

	input := generated.Object{ID: "1"}
	storeOutput := []generated.Object{generated.Object{ID: "1", CreatedBy: "userID"}}

	suite.store.EXPECT().ListObjects(50).Return(storeOutput, nil)

	rr := httptest.NewRecorder()
	req := suite.request(input)
	q := req.URL.Query()
    q.Add("pageSize", "50")
	req.URL.RawQuery = q.Encode()
	h.ListHandler(rr, req)

	var res []generated.Object
	err := json.NewDecoder(rr.Body).Decode(&res)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), storeOutput, res)
}

func (suite *HandlersTestSuite) TestUpdate() {
	h := Handlers{
		Store:             suite.store,
		CustomLogic:       model.AllCustomLogic{},
		CustomLogicCaller: suite.caller,
		Authenticator:     suite.authenticator,
	}

	input := generated.Object{ID: "1"}
	storeOutput := generated.Object{ID: "2"}

	suite.caller.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	suite.store.EXPECT().UpdateObject(&input, "action").Return(&storeOutput, nil)

	rr := httptest.NewRecorder()
	h.UpdateHandler(rr, mux.SetURLVars(suite.request(input), map[string]string{"id": "1", "action": "action"}))

	assert.Equal(suite.T(), storeOutput, suite.decode(rr.Body))
}

func (suite *HandlersTestSuite) TestUpdateCustomLogic() {
	customLogic := "something"
	h := Handlers{
		Store: suite.store,
		CustomLogic: model.AllCustomLogic{
			Update: map[string]*model.CustomLogic{
				"action": &model.CustomLogic{Before: &customLogic, After: &customLogic},
			},
		},
		CustomLogicCaller: suite.caller,
		Authenticator:     suite.authenticator,
	}

	input := generated.Object{ID: "1"}
	beforeCustomLogicOutput := generated.Object{ID: "1", Test: "test"}
	storeOutput := generated.Object{ID: "2"}
	afterCustomLogicOutput := generated.Object{ID: "3"}

	suite.caller.EXPECT().Call(gomock.Any(), "before", "action").
		Times(1).
		Return(suite.response(beforeCustomLogicOutput), nil)
	suite.store.EXPECT().UpdateObject(&beforeCustomLogicOutput, "action").Return(&storeOutput, nil)
	suite.caller.EXPECT().Call(gomock.Any(), "after", "action").
		Times(1).
		Return(suite.response(afterCustomLogicOutput), nil)

	rr := httptest.NewRecorder()
	h.UpdateHandler(rr, mux.SetURLVars(suite.request(input), map[string]string{"id": "1", "action": "action"}))

	assert.Equal(suite.T(), afterCustomLogicOutput, suite.decode(rr.Body))
}

func (suite *HandlersTestSuite) TestDelete() {
	h := Handlers{
		Store:             suite.store,
		CustomLogic:       model.AllCustomLogic{},
		CustomLogicCaller: suite.caller,
		Authenticator:     suite.authenticator,
		Auth: auth,
	}

	getOutput := generated.Object{ID: "1", CreatedBy: "userID"}

	suite.caller.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	suite.store.EXPECT().GetObject("1").Return(&getOutput, nil)
	suite.store.EXPECT().DeleteObject("1").Return(nil)

	rr := httptest.NewRecorder()
	h.DeleteHandler(rr, mux.SetURLVars(suite.request(getOutput), map[string]string{"id": "1"}))
}

func (suite *HandlersTestSuite) TestDeleteCustomLogic() {
	customLogic := "something"
	h := Handlers{
		Store:             suite.store,
		CustomLogic:       model.AllCustomLogic{Delete: &model.CustomLogic{Before: &customLogic, After: &customLogic}},
		CustomLogicCaller: suite.caller,
		Authenticator:     suite.authenticator,
		Auth: auth,
	}

	getOutput := generated.Object{ID: "1", CreatedBy: "userID"}
	afterCustomLogicOutput := generated.Object{ID: "2"}

	suite.store.EXPECT().GetObject("1").Return(&getOutput, nil)
	suite.caller.EXPECT().Call(gomock.Any(), "before", metrics.DELETE).
		Times(1).
		Return(suite.response(getOutput), nil)
	suite.store.EXPECT().DeleteObject("1").Return(nil)
	suite.caller.EXPECT().Call(gomock.Any(), "after", metrics.DELETE).
		Times(1).
		Return(suite.response(afterCustomLogicOutput), nil)

	rr := httptest.NewRecorder()
	h.DeleteHandler(rr, mux.SetURLVars(suite.request(getOutput), map[string]string{"id": "1"}))
}

func (suite *HandlersTestSuite) request(obj generated.Object) *http.Request {
	req, err := http.NewRequest("POST", "url", suite.encode(obj))
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

/*func TestBeforeCustomLogicNoop2(t *testing.T) {
	obj := &generated.Object{ID: "id", Test: "test"}
	bs, err := json.Marshal(obj)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "url", bytes.NewReader((bs)))
	assert.NoError(t, err)

	res, err := applyBeforeCustomLogic(req, &model.CustomLogic{}, metrics.CREATE)
	assert.NoError(t, err)
	assert.Equal(t, obj, res)
}

func TestBeforeCustomLogic(t *testing.T) {
	before := "foo"
	customLogic := &model.CustomLogic{Before: &before}
	r := &http.Request{}
	applyBeforeCustomLogic(r, customLogic, metrics.CREATE)
}
*/

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}
