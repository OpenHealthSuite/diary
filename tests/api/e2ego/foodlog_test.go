package e2e_test

import (
	"net/http"
	"os"
	"testing"

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

	srv, err := server.NewServer(&config)
	require.NoError(t, err)

	quit, err := srv.RunServer()

	t.Run("Happy Path :: Bad Retreives, Creates, Retreives, Edits, Reretrieves, Deletes, Fails Retreive, Redelete succeeds false", func(tt *testing.T) {

		test_user_id := uuid.NewString()

		random_id := uuid.NewString()

		req, err := http.NewRequest("GET", "http://localhost:8937/api/logs"+random_id, nil)
		req.Header.Set(useridheader, test_user_id)
		require.NoError(t, err)
		hc := http.Client{}
		resp, err := hc.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode)

	})

	(*quit) <- os.Interrupt
}
