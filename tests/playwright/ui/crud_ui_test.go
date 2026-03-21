package ui_test

import (
	"os"
	"testing"

	"github.com/openhealthsuite/diary/internal/config"
	"github.com/openhealthsuite/diary/internal/server"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CRUD_Logs(t *testing.T) {

	config := config.ServerConfiguration{
		Port:                     8936,
		PostgresConnectionString: "",
		SqliteFile:               ":memory:",
		UserId:                   "single-test-user-id",
		SignoutEndpoint:          "/logout",
		UserIdHeader:             "",

		TemplateDirectory: "../../../web/template",
		StaticDirectory:   "../../../web/static",
	}

	srv, err := server.NewServer(&config)
	require.NoError(t, err)

	quit, err := srv.RunServer()

	pw, err := playwright.Run()
	require.NoError(t, err)

	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)

	t.Run("Test create log", func(tt *testing.T) {
		page, err := browser.NewPage()
		require.NoError(tt, err)
		if _, err = page.Goto("http://localhost:8936"); err != nil {
			tt.Fatalf("could not goto: %v", err)
		}
		alb := page.Locator(".add-log-button").First()
		err = alb.Click()
		require.NoError(tt, err)
		ln := page.Locator("[name='name']").First()
		err = ln.Fill("My Log Name")
		require.NoError(tt, err)
		mc := page.Locator("[name='metric_calories']").First()
		err = mc.Fill("123")
		require.NoError(tt, err)
		sub := page.Locator("button[type='submit']").First()
		err = sub.Click()
		require.NoError(tt, err)
		tm := page.Locator(".topmetric").First()
		tms, err := tm.TextContent()
		require.NoError(tt, err)
		assert.Equal(tt, "123 Calories Total", tms)
	})

	(*quit) <- os.Interrupt
}
