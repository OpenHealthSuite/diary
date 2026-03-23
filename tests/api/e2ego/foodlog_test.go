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

func Test_CRUD_Logs(t *testing.T) {

	useridheader := "x-user-id"

	config := config.ServerConfiguration{
		Port:                     8937,
		PostgresConnectionString: "",
		SqliteFile:               ":memory:",

		SignoutEndpoint: "/logout",
		UserIdHeader:    useridheader,

		TemplateDirectory: "../../../web/template",
		StaticDirectory:   "../../../web/static",
	}

	target_host := "http://localhost:8937"

	srv, err := server.NewServer(&config)
	require.NoError(t, err)

	quit, err := srv.RunServer()

	t.Run("Happy Path :: Bad Retreives, Creates, Retreives, Edits, Reretrieves, Deletes, Fails Retreive, Redelete succeeds false", func(tt *testing.T) {

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
		// storedItem := toMap(t, resp.Body)
		// for k, v := range created_item {
		// 	assert.Equal(t, v, storedItem[k])
		// }

	})

	(*quit) <- os.Interrupt
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

// func toMap(t *testing.T, bdy io.Reader) map[any]any {
// 	rs := map[any]any{}
// 	buf := new(bytes.Buffer)
// 	_, err := buf.ReadFrom(bdy)
// 	require.NoError(t, err)
// 	err = json.Unmarshal(buf.Bytes(), &rs)
// 	require.NoError(t, err)
// 	return rs
// }
