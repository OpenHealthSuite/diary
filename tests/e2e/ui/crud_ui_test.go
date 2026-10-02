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

	target := "http://localhost:8936"

	if os.Getenv("OFD_E2E_TARGET") != "" {
		target = os.Getenv("OFD_E2E_TARGET")
	} else {
		config := config.ServerConfiguration{
			Port:              8936,
			SqliteFile:        ":memory:",
			UserId:            "e2e-ui-user",
			TemplateDirectory: "../../../web/template",
			StaticDirectory:   "../../../web/static",
		}

		srv, err := server.NewServer(&config)
		require.NoError(t, err)
		quit, err := srv.RunServer()
		defer func() {
			(*quit) <- os.Interrupt
		}()
	}

	pw, err := playwright.Run()
	require.NoError(t, err)

	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)

	t.Run("Test create log", func(tt *testing.T) {
		page, err := browser.NewPage()
		require.NoError(tt, err)
		if _, err = page.Goto(target); err != nil {
			tt.Fatalf("could not goto: %v", err)
		}
		alb := page.Locator(".add-log-button").First()
		err = alb.Click()
		require.NoError(tt, err)
		ln := page.Locator("[name='name']").First()
		err = ln.Fill("First Log")
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
	t.Run("Test edit log", func(tt *testing.T) {
		page, err := browser.NewPage()
		require.NoError(tt, err)
		if _, err = page.Goto(target); err != nil {
			tt.Fatalf("could not goto: %v", err)
		}
		alb := page.Locator(".add-log-button").First()
		err = alb.Click()
		require.NoError(tt, err)
		ln := page.Locator("[name='name']").First()
		err = ln.Fill("Second Log")
		require.NoError(tt, err)
		mc := page.Locator("[name='metric_calories']").First()
		err = mc.Fill("543")
		require.NoError(tt, err)
		sub := page.Locator("button[type='submit']").First()
		err = sub.Click()
		require.NoError(tt, err)
		_, err = page.Reload()
		require.NoError(tt, err)
		tm := page.Locator(".topmetric").First()
		tms, err := tm.TextContent()
		require.NoError(tt, err)
		assert.Equal(tt, "666 Calories Total", tms)

		secondLog := page.Locator(".food-log", playwright.PageLocatorOptions{
			HasText: "Second Log",
		}).First()
		err = secondLog.GetByRole("button", playwright.LocatorGetByRoleOptions{
			Name: "Edit",
		}).Click()
		require.NoError(tt, err)
		emc := page.Locator("[name='metric_calories']").First()
		err = emc.Clear()
		require.NoError(tt, err)
		err = emc.Fill("678")
		require.NoError(tt, err)
		esub := page.GetByRole("button", playwright.PageGetByRoleOptions{
			Name: "Submit",
		}).First()
		err = esub.Click()
		require.NoError(tt, err)

		_, err = page.Reload()
		require.NoError(tt, err)
		etm := page.Locator(".topmetric").First()
		etms, err := etm.TextContent()
		require.NoError(tt, err)
		assert.Equal(tt, "801 Calories Total", etms)
	})
}
