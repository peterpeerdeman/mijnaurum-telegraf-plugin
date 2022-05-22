package mijnaurum

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

var testdataDir = getTestdataDir()

func TestInitDefault(t *testing.T) {
	// This test should succeed with the default initialization.
	plugin := &MijnAurum{
		Username: "testuser",
		Password: "testpass",
		Log:      testutil.Logger{},
	}

	// Test the initialization succeeds
	require.NoError(t, plugin.Init())

	// Also test that default values are set correctly
	require.Equal(t, "testuser", plugin.Username)
	require.Equal(t, "testpass", plugin.Password)
	require.Equal(t, "https://mijnaurum.nl", plugin.url)
}

func TestInitFailAllEmpty(t *testing.T) {
	plugin := &MijnAurum{
		Log: testutil.Logger{},
	}
	require.Error(t, plugin.Init())
}

func TestInitFailUserEmpty(t *testing.T) {
	plugin := &MijnAurum{
		Log:      testutil.Logger{},
		Password: "testpass",
	}
	require.Error(t, plugin.Init())
}

func TestInitFailPassEmpty(t *testing.T) {
	plugin := &MijnAurum{
		Log:      testutil.Logger{},
		Username: "testuser",
	}
	require.Error(t, plugin.Init())
}

func TestGather(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/user/v2/authentication" {
					sampleAuthenticationResponse, err := ioutil.ReadFile(filepath.Join(testdataDir, "sample_authentication_response.json"))
					cookie := &http.Cookie{Name: "Auth-Token", Value: "80d2255d-d4fb-4d50-a3d5-b86ae7e6aae8"}
					http.SetCookie(w, cookie)
					w.WriteHeader(http.StatusOK)
					_, err = fmt.Fprintln(w, string(sampleAuthenticationResponse))
					require.NoError(t, err)
				} else if r.URL.Path == "/user/v2/users/abcuserid/sources" {
					sampleSourcesResponse, err := ioutil.ReadFile(filepath.Join(testdataDir, "sample_sources_response.json"))
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
					_, err = fmt.Fprintln(w, string(sampleSourcesResponse))
					require.NoError(t, err)
				} else if r.URL.Path == "/user/v2/users/abcuserid/actuals" {
					sampleActualsResponse, err := ioutil.ReadFile(filepath.Join(testdataDir, "sample_actuals_response.json"))
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
					_, err = fmt.Fprintln(w, string(sampleActualsResponse))
					require.NoError(t, err)
				} else if r.URL.Path == "/user/v2/users/abcuserid/history" {
					sampleGetHistoryResponse, err := ioutil.ReadFile(filepath.Join(testdataDir, "sample_history_response.json"))
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
					_, err = fmt.Fprintln(w, string(sampleGetHistoryResponse))
					require.NoError(t, err)
				}
			},
		),
	)
	defer ts.Close()

	tests := []struct {
		name     string
		plugin   *MijnAurum
		expected []telegraf.Metric
	}{
		{
			name: "gather heat",
			plugin: &MijnAurum{
				Username:   "testuser",
				Password:   "testpass",
				Collectors: []string{"heat"},
				url:        ts.URL,
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"mijnaurum",
					map[string]string{
						"source":      "dZGWYt_pp20TnzlFzHxBKzsOR6X-cXA-xLZLSTNMuJaVzodeWGHa1SJS03mDjykT",
						"source_type": "heat",
						"rate_unit":   "J/h",
						"unit":        "GJ",
						"meter_id":    "D0zkFgSTtzvSSmqvAmshBMJp_qVkgpjqEuibvND3l9n-VEhaVEyFF97vQNfzVY5j",
						"location_id": "z4O1ho3dGhmH-w2zu2d4YOgUsP77jfQadSw0lCL3SnqONvtExoxS-tgjiAmyxdmK",
					},
					map[string]interface{}{
						"day_value":   0.014,
						"day_cost":    0.30198,
						"week_value":  0.088,
						"week_cost":   1.8981599999999998,
						"month_value": 0.395,
						"month_cost":  8.520150000000001,
						"year_value":  1.6360000000000001,
						"year_cost":   35.288520000000005,
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator
			tt.plugin.Log = testutil.Logger{}
			require.NoError(t, tt.plugin.Init())
			require.NoError(t, tt.plugin.Gather(&acc))
			require.Len(t, acc.Errors, 0, "found errors accumulated by acc.AddError()")
			acc.Wait(len(tt.expected))
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestAuthenticationFailed(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				sampleAuthenticationFailedResponse, err := ioutil.ReadFile(filepath.Join(testdataDir, "sample_authentication_failed_response.json"))
				_, err = fmt.Fprintln(w, string(sampleAuthenticationFailedResponse))
				require.NoError(t, err)
			},
		),
	)
	defer ts.Close()

	tests := []struct {
		name     string
		plugin   *MijnAurum
		expected string
	}{
		{
			name: "authentication failed",
			plugin: &MijnAurum{
				Username: "usertest",
				Password: "userpass",
				url:      ts.URL,
			},
			expected: "statuscode from mijnaurum authentication was not 200 but 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator

			tt.plugin.Log = testutil.Logger{}
			require.NoError(t, tt.plugin.Init())

			err := tt.plugin.Gather(&acc)
			require.Error(t, err)
			require.EqualError(t, err, tt.expected)
		})
	}
}

func TestGetSourceString(t *testing.T) {
	sampleSourcesResponse, err := ioutil.ReadFile(filepath.Join(testdataDir, "sample_sources_response.json"))
	if err != nil {
		log.Fatalf("Error occured during unmarshaling. Error: %s", err.Error())
	}
	sourcesResponse := SourcesResponse{}
	json.Unmarshal([]byte(sampleSourcesResponse), &sourcesResponse)
	plugin := &MijnAurum{
		Username:   "testuser",
		Password:   "testpass",
		Collectors: []string{"heat"},
		sources:    sourcesResponse.Sources,
	}

	sourcesString := plugin.getSourceString()
	require.Equal(t, "aBebAKYnZnd5p6wFRWoT8iUlamI5DJkoEFWwBBQOqb519dNRfqFwVWiHrGQAR0pV,dZGWYt_pp20TnzlFzHxBKzsOR6X-cXA-xLZLSTNMuJaVzodeWGHa1SJS03mDjykT", sourcesString)

}

func getTestdataDir() string {
	dir, err := os.Getwd()
	if err != nil {
		// if we cannot even establish the test directory, further progress is meaningless
		panic(err)
	}

	return filepath.Join(dir, "testdata")
}
