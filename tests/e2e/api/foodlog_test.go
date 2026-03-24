package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CRUD_Logs(xt *testing.T) {
	target_host := "http://localhost:8937"
	useridheader := "x-openfooddiary-userid"

	if os.Getenv("OFD_E2E_TARGET") != "" {
		target_host = os.Getenv("OFD_E2E_TARGET")
	} else {

		config := config.ServerConfiguration{
			Port:                     8937,
			PostgresConnectionString: "",
			SqliteFile:               ":memory:",

			SignoutEndpoint: "/logout",
			UserIdHeader:    useridheader,

			TemplateDirectory: "../../../web/template",
			StaticDirectory:   "../../../web/static",
		}

		srv, err := server.NewServer(&config)
		require.NoError(xt, err)
		quit, err := srv.RunServer()
		defer func() {
			(*quit) <- os.Interrupt
		}()
	}

	xt.Run("Happy Path :: Bad Retreives, Creates, Retreives, Edits, Reretrieves, Deletes, Fails Retreive, Redelete succeeds false", func(t *testing.T) {

		test_user_id := uuid.NewString()

		random_id := uuid.NewString()

		hdrs := map[string]string{
			useridheader: test_user_id,
		}

		hc := http.Client{}

		resp := doReq(&hc, t, "GET", target_host+"/api/logs"+random_id, nil, hdrs)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)

		date := time.Now()
		endDate := date.Add(5 * time.Minute)

		created_item := map[string]any{
			"name":   "My Food Log",
			"labels": []string{"some label", "other label"},
			"time": map[string]string{
				"start": date.Format(time.RFC3339),
				"end":   endDate.Format(time.RFC3339),
			},
			"metrics": map[string]int64{
				"calories": 500,
			},
		}
		resp = doReq(&hc, t, "POST", target_host+"/api/logs", toJsonBody(t, created_item), hdrs)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		testItemId := toString(t, resp.Body)
		fmt.Println(testItemId)
		assert.Greater(t, len(testItemId), 0)

		resp = doReq(&hc, t, "GET", target_host+"/api/logs/"+testItemId, nil, hdrs)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		storedItem := decode[map[string]any](t, resp.Body)
		assertSubset(t, created_item, storedItem)

		modified_item := map[string]any{
			"name":   "My Updated Food Log",
			"labels": []string{"some other label", "other some label"},
			"time": map[string]string{
				"start": date.Add(-2 * time.Hour).Format(time.RFC3339),
				"end":   endDate.Add(-1 * time.Hour).Format(time.RFC3339),
			},
			"metrics": map[string]int64{
				"calories": 765,
			},
		}

		resp = doReq(&hc, t, "PUT", target_host+"/api/logs/"+testItemId, toJsonBody(t, modified_item), hdrs)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		resp = doReq(&hc, t, "GET", target_host+"/api/logs/"+testItemId, nil, hdrs)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		storedItem = decode[map[string]any](t, resp.Body)
		assertSubset(t, modified_item, storedItem)

		resp = doReq(&hc, t, "DELETE", target_host+"/api/logs/"+testItemId, nil, hdrs)
		defer resp.Body.Close()

		assert.Equal(t, 204, resp.StatusCode)

		resp = doReq(&hc, t, "GET", target_host+"/api/logs/"+testItemId, nil, hdrs)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)

		resp = doReq(&hc, t, "DELETE", target_host+"/api/logs/"+testItemId, nil, hdrs)
		defer resp.Body.Close()

		assert.Equal(t, 204, resp.StatusCode)

	})

	xt.Run("Queries :: can add some logs, and get expected query results", func(t *testing.T) {

		test_user_id := uuid.NewString()

		hdrs := map[string]string{
			useridheader: test_user_id,
		}

		hc := http.Client{}

		past_log := map[string]any{
			"name":   "Past Log",
			"labels": []string{},
			"time": map[string]any{
				"start": time.Date(1999, time.October, 10, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
				"end":   time.Date(1999, time.October, 11, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
			"metrics": map[string]any{
				"calories": int64(500),
			},
		}

		center_log := map[string]any{
			"name":   "Center Log",
			"labels": []string{},
			"time": map[string]any{
				"start": time.Date(1999, time.October, 15, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
				"end":   time.Date(1999, time.October, 16, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
			"metrics": map[string]any{
				"calories": int64(500),
			},
		}

		future_log := map[string]any{
			"name":   "Future Log",
			"labels": []string{},
			"time": map[string]any{
				"start": time.Date(1999, time.October, 20, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
				"end":   time.Date(1999, time.October, 21, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
			"metrics": map[string]any{
				"calories": int64(500),
			},
		}

		resp := doReq(&hc, t, "POST", target_host+"/api/logs", toJsonBody(t, past_log), hdrs)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		past_log_id := toString(t, resp.Body)

		resp = doReq(&hc, t, "POST", target_host+"/api/logs", toJsonBody(t, center_log), hdrs)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		center_log_id := toString(t, resp.Body)

		resp = doReq(&hc, t, "POST", target_host+"/api/logs", toJsonBody(t, future_log), hdrs)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		future_log_id := toString(t, resp.Body)

		queryLogs := func(start_date, end_date time.Time) []map[string]any {
			resp := doReq(&hc, t, "GET", target_host+"/api/logs?startDate="+start_date.Format(time.RFC3339)+"&endDate="+end_date.Format(time.RFC3339), nil, hdrs)
			defer resp.Body.Close()

			assert.Equal(t, 200, resp.StatusCode)
			return decode[[]map[string]any](t, resp.Body)
		}

		lgs := queryLogs(time.Date(1999, time.October, 14, 0, 0, 0, 0, time.UTC), time.Date(1999, time.October, 16, 0, 0, 0, 0, time.UTC))

		assert.Len(t, lgs, 1)
		assert.Equal(t, center_log_id, lgs[0]["id"].(string))

		lgs = queryLogs(time.Date(1999, time.October, 9, 0, 0, 0, 0, time.UTC), time.Date(1999, time.October, 16, 0, 0, 0, 0, time.UTC))

		assert.Len(t, lgs, 2)
		assert.Equal(t, past_log_id, lgs[0]["id"].(string))
		assert.Equal(t, center_log_id, lgs[1]["id"].(string))

		lgs = queryLogs(time.Date(1999, time.October, 15, 0, 0, 0, 0, time.UTC), time.Date(1999, time.October, 30, 0, 0, 0, 0, time.UTC))

		assert.Len(t, lgs, 2)
		assert.Equal(t, center_log_id, lgs[0]["id"].(string))
		assert.Equal(t, future_log_id, lgs[1]["id"].(string))

		lgs = queryLogs(time.Date(2011, time.October, 15, 0, 0, 0, 0, time.UTC), time.Date(2012, time.October, 30, 0, 0, 0, 0, time.UTC))

		assert.Len(t, lgs, 0)
	})

}

func assertSubset(t *testing.T, sourceItem map[string]any, comparedItem map[string]any) {
	for k, v := range sourceItem {
		storedVal := comparedItem[k]
		var expectedNorm, actualNorm interface{}
		expectedBytes, err := json.Marshal(v)
		require.NoError(t, err)
		actualBytes, err := json.Marshal(storedVal)
		require.NoError(t, err)
		json.Unmarshal(expectedBytes, &expectedNorm)
		json.Unmarshal(actualBytes, &actualNorm)
		assert.Equal(t, expectedNorm, actualNorm)
	}
}

func doReq(hc *http.Client, t *testing.T, method string, url string, body io.Reader, headers map[string]string) *http.Response {
	req, err := http.NewRequest(method, url, body)
	for k, v := range headers {

		req.Header.Set(k, v)
	}
	require.NoError(t, err)
	resp, err := hc.Do(req)
	require.NoError(t, err)
	return resp
}

func toJsonBody(t *testing.T, obj interface{}) io.Reader {
	rs, err := json.Marshal(obj)
	require.NoError(t, err)
	return bytes.NewReader(rs)
}

func toString(t *testing.T, bdy io.Reader) string {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(bdy)
	require.NoError(t, err)
	return buf.String()

}

func decode[E any](t *testing.T, bdy io.Reader) E {
	rs := new(E)
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(bdy)
	require.NoError(t, err)
	err = json.Unmarshal(buf.Bytes(), &rs)
	require.NoError(t, err)
	return *rs
}
