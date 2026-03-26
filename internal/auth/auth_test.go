package auth_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetUserId_HappyPath(xt *testing.T) {
	gc, _ := gin.CreateTestContext(httptest.NewRecorder())
	testuserid := "some-user-id"
	gc.Set("userId", testuserid)
	uid, err := auth.GetUserId(gc)
	require.NoError(xt, err)
	assert.Equal(xt, testuserid, *uid)
}

func Test_GetUserId_Missing(xt *testing.T) {
	gc, _ := gin.CreateTestContext(httptest.NewRecorder())
	uid, err := auth.GetUserId(gc)
	require.Error(xt, err)
	assert.Nil(xt, uid)
	assert.Equal(xt, auth.ErrNoUserIdentification, err)
}

func Test_GetUserId_NotString(xt *testing.T) {
	gc, _ := gin.CreateTestContext(httptest.NewRecorder())
	gc.Set("userId", 123)
	uid, err := auth.GetUserId(gc)
	require.Error(xt, err)
	assert.Nil(xt, uid)
	assert.Equal(xt, auth.ErrNoUserIdentification, err)
}
