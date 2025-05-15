package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"path"
	"testing"

	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/random"
)

const (
	host = "localhost:8080"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	resp := e.POST("/url").
		WithJSON(save.Request{
			URLToSave: gofakeit.URL(),
			Alias:     random.NewRandomString(10),
		}).
		WithBasicAuth("myuser", "mypass").
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	resp.Value("status").String().IsEqual("ok")
	resp.ContainsKey("alias")
}

func TestURLShortener_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			error: "field URL is not a valid URL",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			resp := e.POST("/url").
				WithJSON(save.Request{
					URLToSave: tc.url,
					Alias:     tc.alias,
				}).
				WithBasicAuth("myuser", "mypass").
				Expect().
				Status(http.StatusOK).
				JSON().Object()

			if tc.error != "" {
				resp.Value("error").String().IsEqual(tc.error)
				resp.NotContainsKey("alias")
				return
			}

			resp.Value("status").String().IsEqual("ok")

			alias := tc.alias
			if alias == "" {
				resp.Value("alias").String().NotEmpty()
				alias = resp.Value("alias").String().Raw()
			} else {
				resp.Value("alias").String().IsEqual(alias)
			}

			// Проверка редиректа (обязательно должен быть GET хэндлер)
			testRedirect(t, alias, tc.url)

			// Удаление
			respDel := e.DELETE("/"+path.Join("url", alias)).
				WithBasicAuth("myuser", "mypass").
				Expect().
				Status(http.StatusOK).
				JSON().Object()

			respDel.Value("status").String().IsEqual("ok")

			testRedirectNotFound(t, alias)
		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   "/" + alias,
	}

	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)
	require.Equal(t, urlToRedirect, redirectedToURL)
}

func testRedirectNotFound(t *testing.T, alias string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   "/" + alias, // ✅ prepend "/" here too
	}

	_, err := api.GetRedirect(u.String())
	require.Equal(t, err, api.ErrInvalidStatusCode)
}
