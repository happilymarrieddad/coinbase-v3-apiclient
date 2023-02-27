// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/QuantFu-Inc/coinbase-adv/client (interfaces: CoinbaseClient)

// Package mocks is a generated GoMock package.
package mocks

import (
	http "net/http"
	reflect "reflect"

	client "github.com/QuantFu-Inc/coinbase-adv/client"
	model "github.com/QuantFu-Inc/coinbase-adv/model"
	gomock "github.com/golang/mock/gomock"
)

// MockCoinbaseClient is a mock of CoinbaseClient interface.
type MockCoinbaseClient struct {
	ctrl     *gomock.Controller
	recorder *MockCoinbaseClientMockRecorder
}

// MockCoinbaseClientMockRecorder is the mock recorder for MockCoinbaseClient.
type MockCoinbaseClientMockRecorder struct {
	mock *MockCoinbaseClient
}

// NewMockCoinbaseClient creates a new mock instance.
func NewMockCoinbaseClient(ctrl *gomock.Controller) *MockCoinbaseClient {
	mock := &MockCoinbaseClient{ctrl: ctrl}
	mock.recorder = &MockCoinbaseClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCoinbaseClient) EXPECT() *MockCoinbaseClientMockRecorder {
	return m.recorder
}

// CancelOrders mocks base method.
func (m *MockCoinbaseClient) CancelOrders(arg0 []string) (*model.CancelOrderResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CancelOrders", arg0)
	ret0, _ := ret[0].(*model.CancelOrderResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CancelOrders indicates an expected call of CancelOrders.
func (mr *MockCoinbaseClientMockRecorder) CancelOrders(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CancelOrders", reflect.TypeOf((*MockCoinbaseClient)(nil).CancelOrders), arg0)
}

// CheckAuthentication mocks base method.
func (m *MockCoinbaseClient) CheckAuthentication(arg0 *http.Request, arg1 []byte) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CheckAuthentication", arg0, arg1)
}

// CheckAuthentication indicates an expected call of CheckAuthentication.
func (mr *MockCoinbaseClientMockRecorder) CheckAuthentication(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckAuthentication", reflect.TypeOf((*MockCoinbaseClient)(nil).CheckAuthentication), arg0, arg1)
}

// CreateOrder mocks base method.
func (m *MockCoinbaseClient) CreateOrder(arg0 *model.CreateOrderRequest) (*model.CreateOrderResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOrder", arg0)
	ret0, _ := ret[0].(*model.CreateOrderResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateOrder indicates an expected call of CreateOrder.
func (mr *MockCoinbaseClientMockRecorder) CreateOrder(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOrder", reflect.TypeOf((*MockCoinbaseClient)(nil).CreateOrder), arg0)
}

// GetAccount mocks base method.
func (m *MockCoinbaseClient) GetAccount(arg0 string) (*model.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccount", arg0)
	ret0, _ := ret[0].(*model.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccount indicates an expected call of GetAccount.
func (mr *MockCoinbaseClientMockRecorder) GetAccount(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccount", reflect.TypeOf((*MockCoinbaseClient)(nil).GetAccount), arg0)
}

// GetOrder mocks base method.
func (m *MockCoinbaseClient) GetOrder(arg0 string) (*model.GetOrderResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrder", arg0)
	ret0, _ := ret[0].(*model.GetOrderResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrder indicates an expected call of GetOrder.
func (mr *MockCoinbaseClientMockRecorder) GetOrder(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrder", reflect.TypeOf((*MockCoinbaseClient)(nil).GetOrder), arg0)
}

// GetPrice mocks base method.
func (m *MockCoinbaseClient) GetPrice(arg0, arg1 string) (*float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPrice", arg0, arg1)
	ret0, _ := ret[0].(*float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPrice indicates an expected call of GetPrice.
func (mr *MockCoinbaseClientMockRecorder) GetPrice(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPrice", reflect.TypeOf((*MockCoinbaseClient)(nil).GetPrice), arg0, arg1)
}

// GetProduct mocks base method.
func (m *MockCoinbaseClient) GetProduct(arg0 string) (*model.GetProductResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProduct", arg0)
	ret0, _ := ret[0].(*model.GetProductResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProduct indicates an expected call of GetProduct.
func (mr *MockCoinbaseClientMockRecorder) GetProduct(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProduct", reflect.TypeOf((*MockCoinbaseClient)(nil).GetProduct), arg0)
}

// GetQuote mocks base method.
func (m *MockCoinbaseClient) GetQuote(arg0 string) (*client.Quote, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetQuote", arg0)
	ret0, _ := ret[0].(*client.Quote)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetQuote indicates an expected call of GetQuote.
func (mr *MockCoinbaseClientMockRecorder) GetQuote(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetQuote", reflect.TypeOf((*MockCoinbaseClient)(nil).GetQuote), arg0)
}

// HttpClient mocks base method.
func (m *MockCoinbaseClient) HttpClient() *http.Client {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HttpClient")
	ret0, _ := ret[0].(*http.Client)
	return ret0
}

// HttpClient indicates an expected call of HttpClient.
func (mr *MockCoinbaseClientMockRecorder) HttpClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HttpClient", reflect.TypeOf((*MockCoinbaseClient)(nil).HttpClient))
}

// IsTokenValid mocks base method.
func (m *MockCoinbaseClient) IsTokenValid(arg0 int64) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsTokenValid", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsTokenValid indicates an expected call of IsTokenValid.
func (mr *MockCoinbaseClientMockRecorder) IsTokenValid(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsTokenValid", reflect.TypeOf((*MockCoinbaseClient)(nil).IsTokenValid), arg0)
}

// ListAccounts mocks base method.
func (m *MockCoinbaseClient) ListAccounts(arg0 *client.ListAccountsParams) (*model.ListAccountsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAccounts", arg0)
	ret0, _ := ret[0].(*model.ListAccountsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAccounts indicates an expected call of ListAccounts.
func (mr *MockCoinbaseClientMockRecorder) ListAccounts(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAccounts", reflect.TypeOf((*MockCoinbaseClient)(nil).ListAccounts), arg0)
}

// ListFills mocks base method.
func (m *MockCoinbaseClient) ListFills(arg0 *client.ListFillsParams) (*model.ListFillsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListFills", arg0)
	ret0, _ := ret[0].(*model.ListFillsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListFills indicates an expected call of ListFills.
func (mr *MockCoinbaseClientMockRecorder) ListFills(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListFills", reflect.TypeOf((*MockCoinbaseClient)(nil).ListFills), arg0)
}

// ListOrders mocks base method.
func (m *MockCoinbaseClient) ListOrders(arg0 *client.ListOrdersParams) (*model.ListOrdersResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListOrders", arg0)
	ret0, _ := ret[0].(*model.ListOrdersResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListOrders indicates an expected call of ListOrders.
func (mr *MockCoinbaseClientMockRecorder) ListOrders(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOrders", reflect.TypeOf((*MockCoinbaseClient)(nil).ListOrders), arg0)
}

// SetRateLimit mocks base method.
func (m *MockCoinbaseClient) SetRateLimit(arg0 int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetRateLimit", arg0)
}

// SetRateLimit indicates an expected call of SetRateLimit.
func (mr *MockCoinbaseClientMockRecorder) SetRateLimit(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetRateLimit", reflect.TypeOf((*MockCoinbaseClient)(nil).SetRateLimit), arg0)
}