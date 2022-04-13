package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/k911mipt/geolocation/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testBatch map[string]struct {
	jsonResp string
	err      error
	respCode int
}

type storeMock struct {
	tests testBatch
	t     *testing.T
}

func (s *storeMock) FetchGeoInfo(ctx context.Context, ip string) (store.IpGeoInfo, error) {
	var ipGeoInfo store.IpGeoInfo
	require.NoError(s.t, json.Unmarshal([]byte(s.tests[ip].jsonResp), &ipGeoInfo))
	err := s.tests[ip].err
	return ipGeoInfo, err
}

func buildRequest(ip string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/geolocation/"+ip, nil)
	vars := map[string]string{
		"ip": ip,
	}
	req = mux.SetURLVars(req, vars)
	return req
}

func TestGeoInfoHandler(t *testing.T) {
	storeMock := storeMock{
		t: t,
		tests: testBatch{
			"160.103.7.140": {
				jsonResp: `{
					"ip_address": "160.103.7.140",
					"country_code": "CZ",
					"country": "Nicaragua",
					"city": "New Neva",
					"latitude": -68.31023296602508,
					"longitude": -37.62435199624531,
					"mystery_value": "7301823115"
				}`,
				err:      nil,
				respCode: 200,
			},
			"0.0": {
				jsonResp: `{
					"code": 400,
					"message": "Invalid IP"
				}`,
				err:      nil,
				respCode: 400,
			},
			"8.8.8.8": {
				jsonResp: `{
					"code": 404,
					"message": "No geoinfo found for the given IP"
				}`,
				err:      store.ErrNotFound,
				respCode: 404,
			},
		},
	}
	geoInfoHandler := getGeoInfoHandler(&storeMock)

	for ip, test := range storeMock.tests {
		req := buildRequest(ip)
		resp := httptest.NewRecorder()
		geoInfoHandler(resp, req)

		assert.Equal(t, test.respCode, resp.Code)
		assert.JSONEq(t, test.jsonResp, resp.Body.String())
	}
}
