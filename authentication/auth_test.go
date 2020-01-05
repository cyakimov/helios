package authentication

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cyakimov/helios/authentication/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockProvider struct {
	mock.Mock
}

func (m *mockProvider) FetchUser(r *http.Request) (providers.UserInfo, error) {
	args := m.Called(r)
	profile := args.Get(0)

	return profile.(providers.UserInfo), nil
}

func (m *mockProvider) GetLoginURL(callbackURL, state string) string {
	args := m.Called(callbackURL, state)
	return args.String(0)
}

func TestHelios_Middleware(t *testing.T) {
	// returns a http.HandlerFunc for testing http middleware
	testHandler := func() http.HandlerFunc {
		fn := func(rw http.ResponseWriter, req *http.Request) {}
		return http.HandlerFunc(fn)
	}

	oauth2 := new(mockProvider)
	auth := NewHeliosAuthentication(oauth2, "test", 5*time.Minute)

	mdw := auth.Middleware(testHandler())

	token, _ := IssueJWTWithSecret("test", "t@t", time.Now().Add(5*time.Minute))
	tests := []struct {
		HeaderName, HeaderValue string
		StatusCode              int
	}{
		{
			HeaderName:  "",
			HeaderValue: "",
			StatusCode:  http.StatusTemporaryRedirect,
		},
		{
			HeaderName:  HeaderName,
			HeaderValue: token,
			StatusCode:  http.StatusOK,
		},
		{
			HeaderName:  HeaderName,
			HeaderValue: "jiberish",
			StatusCode:  http.StatusTemporaryRedirect,
		},
		{
			HeaderName:  "Cookie",
			HeaderValue: CookieName + "=abc",
			StatusCode:  http.StatusTemporaryRedirect,
		},
		{
			HeaderName:  "Cookie",
			HeaderValue: CookieName + "=" + token,
			StatusCode:  http.StatusOK,
		},
	}

	// setup expectations
	loginURL := "http://login"
	oauth2.On("GetLoginURL", "http://testing/.well-known/callback", "http://testing").Return(loginURL).Times(3)
	for _, test := range tests {
		req := httptest.NewRequest("GET", "http://testing", nil)
		res := httptest.NewRecorder()
		req.Header.Set(test.HeaderName, test.HeaderValue)
		mdw.ServeHTTP(res, req)
		assert.Equal(t, test.StatusCode, res.Code)
	}
	oauth2.AssertExpectations(t)
}

func TestHelios_CallbackHandler(t *testing.T) {
	oauth2 := new(mockProvider)
	auth := NewHeliosAuthentication(oauth2, "test", 5*time.Minute)

	state := base64.StdEncoding.EncodeToString([]byte("http://secret-stuff"))
	tests := []struct {
		URL        string
		StatusCode int
	}{
		{"http://testing?state=123", http.StatusInternalServerError},
		{"http://testing?state=" + state, http.StatusFound},
	}

	profile := providers.UserInfo{Email: "t@test"}
	oauth2.On("FetchUser", mock.Anything).Return(profile)

	for _, test := range tests {
		req := httptest.NewRequest("GET", test.URL, nil)
		res := httptest.NewRecorder()

		handler := http.HandlerFunc(auth.CallbackHandler)
		handler.ServeHTTP(res, req)

		assert.Equal(t, test.StatusCode, res.Code)
	}
	oauth2.AssertExpectations(t)
}
