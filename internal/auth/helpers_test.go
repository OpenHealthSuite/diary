package auth

import (
	"testing"

	"github.com/openhealthsuite/diary/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig() *config.ServerConfiguration {
	return &config.ServerConfiguration{
		SessionCookieName: "openfooddiary-session",
		SessionMaxAge:     86400,
		SessionSecure:     true,
	}
}

func Test_IsSafeRedirect(t *testing.T) {
	cases := map[string]bool{
		"/logs":            true,
		"/config?tab=1":    true,
		"/":                true,
		"":                 false,
		"logs":             false,
		"//evil.com":       false,
		"https://evil.com": false,
		"/logs\r\nX: y":    false,
	}
	for target, expected := range cases {
		assert.Equal(t, expected, isSafeRedirect(target), "target %q", target)
	}
}

func Test_NewRedisStore_BadUrlFailsAtBoot(t *testing.T) {
	for _, raw := range []string{
		"",
		"not-a-url",
		"http://localhost:6379",
		"redis://localhost:6379/notanumber",
		"redis://127.0.0.1:1/",
	} {
		t.Run(raw, func(t *testing.T) {
			store, err := newRedisStore(&config.ServerConfiguration{
				SessionRedisUrl: raw,
				SessionSecret:   "test-secret",
			})
			require.Error(t, err)
			assert.Nil(t, store)
		})
	}
}

func Test_SessionKeyPairs_Deterministic(t *testing.T) {
	auth1, encrypt1, err := sessionKeyPairs("a-secret")
	require.NoError(t, err)
	auth2, encrypt2, err := sessionKeyPairs("a-secret")
	require.NoError(t, err)
	assert.Equal(t, auth1, auth2)
	assert.Equal(t, encrypt1, encrypt2)
	assert.Len(t, auth1, 32)
	assert.Len(t, encrypt1, 32)

	auth3, _, err := sessionKeyPairs("another-secret")
	require.NoError(t, err)
	assert.NotEqual(t, auth1, auth3)
}

func Test_SessionKeyPairs_EphemeralWhenUnset(t *testing.T) {
	auth1, _, err := sessionKeyPairs("")
	require.NoError(t, err)
	auth2, _, err := sessionKeyPairs("")
	require.NoError(t, err)
	assert.NotEqual(t, auth1, auth2)
}
