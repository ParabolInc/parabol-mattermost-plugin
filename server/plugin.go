package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	botUserID      = "botUserID"
	requestTimeout = 30 * time.Second
	// well below the 4kb limit of nginx
	maxHeaderLength = 1024
)

type SlashCommand struct {
	Trigger     string `json:"trigger"`
	Description string `json:"description"`
}

type ClientConfig struct {
	Commands []SlashCommand `json:"commands"`
}

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	commands []SlashCommand

	// router is the HTTP router for handling API requests.
	router *mux.Router
}

type Context struct {
	Ctx    context.Context
	UserID string
	User   *model.User
}

type HTTPHandlerFuncWithContext func(c *Context, w http.ResponseWriter, r *http.Request)

func safeCopyHeader(from http.Header, header string, to http.Header) error {
	for _, value := range from.Values(header) {
		if len(header)+len(value) > maxHeaderLength {
			return errors.New("header too long")
		}
		// Prevent header injection by disallowing newline characters
		if strings.ContainsAny(value, "\r\n") {
			return errors.New("invalid characters in header")
		}
		to.Add(header, value)
	}
	return nil
}

func (p *Plugin) createContext(userID string) (*Context, context.CancelFunc) {
	user, _ := p.API.GetUser(userID)
	// TODO check email and email verified

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

	context := &Context{
		Ctx:    ctx,
		UserID: userID,
		User:   user,
	}

	return context, cancel
}

func (p *Plugin) authenticated(handler HTTPHandlerFuncWithContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "Not authorized"}`))
			return
		}

		context, cancel := p.createContext(userID)
		defer cancel()
		handler(context, w, r)
	}
}

/*
Mattermost strips the plugin path prefix from the request before forwarding it to the plugin.
If we want to verify the path of the request, we need to add it back.
https://github.com/mattermost/mattermost/blob/751d84bf13aa63f4706843318e45e8ca8401eba5/server/channels/app/plugin_requests.go#L226
*/
func (p *Plugin) fixedPath(handler http.HandlerFunc) http.HandlerFunc {
	pluginID := p.API.GetPluginID()
	path := "/plugins/" + pluginID
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = path + r.URL.Path
		handler(w, r)
	}
}

// signChannelToken creates an HMAC-SHA256 token over channelID using ParabolToken as the key.
// The resulting hex string is an opaque credential that Parabol stores per channel and sends
// as a Bearer token when calling /notify. The Go plugin re-derives and compares to verify.
func signChannelToken(secret []byte, channelID string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(channelID))
	return hex.EncodeToString(mac.Sum(nil))
}

// verifyChannelToken checks a Bearer token against the expected HMAC for channelID.
func verifyChannelToken(secret []byte, channelID, token string) bool {
	expected := signChannelToken(secret, channelID)
	// Use hmac.Equal for constant-time comparison
	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return false
	}
	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		return false
	}
	return hmac.Equal(expectedBytes, tokenBytes)
}

// notify handles POST /notify/{channelID}.
// Auth: Bearer token in the Authorization header, verified as HMAC-SHA256(ParabolToken, channelID).
func (p *Plugin) notify(w http.ResponseWriter, r *http.Request) {
	config := p.getConfiguration()

	vars := mux.Vars(r)
	channelID := vars["channelID"]

	// Verify Bearer token
	authHeader := r.Header.Get("Authorization")
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Missing Bearer token"}`))
		return
	}
	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if !verifyChannelToken([]byte(config.ParabolToken), channelID, token) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "Invalid token"}`))
		return
	}

	userID, err1 := p.API.KVGet(botUserID)
	if err1 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Bot User not found", "originalError": "%v"}`, err1)
		_, _ = w.Write([]byte(msg))
		return
	}

	var props map[string]any
	if err := getJSON(r.Body, &props); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Error parsing body", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}
	if _, err := p.API.CreatePost(&model.Post{
		ChannelId: channelID,
		Props:     props,
		UserId:    string(userID),
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Error posting notification", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}
}

// auth handles GET /auth?state=<base64-encoded-state>.
// It redirects the user to Parabol's OAuth2 authorize endpoint so they can log in.
// After login, Parabol redirects to /mattermost/callback which postMessages the auth token
// back to this opener window via the JS plugin.
func (p *Plugin) auth(c *Context, w http.ResponseWriter, r *http.Request) {
	config := p.getConfiguration()
	if config.OAuthClientID == "" {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error": "OAuth not configured"}`))
		return
	}

	stateParam := r.URL.Query().Get("state")
	if stateParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Missing state parameter"}`))
		return
	}

	redirectURI := config.ParabolURL + "/mattermost/callback"
	q := url.Values{}
	q.Set("client_id", config.OAuthClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "graphql:persisted")
	q.Set("state", stateParam)
	authorizeURL := config.ParabolURL + "/oauth/authorize?" + q.Encode()

	http.Redirect(w, r, authorizeURL, http.StatusFound)
}

// link handles GET /link?channelId=<id>.
// It generates an HMAC-SHA256 channel token and returns it so the JS plugin can pass it
// to Parabol's linkMattermostChannel mutation as the channelToken argument.
func (p *Plugin) link(c *Context, w http.ResponseWriter, r *http.Request) {
	config := p.getConfiguration()
	channelID := r.URL.Query().Get("channelId")
	if channelID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Missing channelId"}`))
		return
	}

	token := signChannelToken([]byte(config.ParabolToken), channelID)
	body, err := json.Marshal(struct {
		ChannelToken string `json:"channelToken"`
	}{
		ChannelToken: token,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Marshal error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}

func (p *Plugin) graphql(w http.ResponseWriter, r *http.Request) {
	config := p.getConfiguration()
	graphqlURL := config.ParabolURL + "/graphql"

	defer func() { _ = r.Body.Close() }()
	req, err1 := http.NewRequest("POST", graphqlURL, r.Body)
	if err1 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Request error", "originalError": "%v"}`, err1)
		_, _ = w.Write([]byte(msg))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if errCopy := safeCopyHeader(r.Header, "x-application-authorization", req.Header); errCopy != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Header error", "originalError": "%v"}`, errCopy)
		_, _ = w.Write([]byte(msg))
		return
	}

	if errCopy := safeCopyHeader(r.Header, "authorization", req.Header); errCopy != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Header error", "originalError": "%v"}`, errCopy)
		_, _ = w.Write([]byte(msg))
		return
	}

	client := &http.Client{}
	res, err2 := client.Do(req)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Request error", "originalError": "%v"}`, err2)
		_, _ = w.Write([]byte(msg))
		return
	}
	defer func() { _ = res.Body.Close() }()

	w.WriteHeader(res.StatusCode)
	_, _ = io.Copy(w, res.Body)
}

func (p *Plugin) getConfig(c *Context, w http.ResponseWriter, r *http.Request) {
	config := p.getConfiguration()
	body, err := json.Marshal(struct {
		ParabolURL string `json:"parabolUrl"`
	}{
		ParabolURL: config.ParabolURL,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Marshal error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}

/*
Endpoint for module federation, serves components from parabol.
We cannot contact the Parabol instance directly from the webapp because of security settings on it.
*/
func (p *Plugin) components(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["file"]
	config := p.getConfiguration()
	componentURL := config.ParabolURL + "/components/" + file

	client := &http.Client{}
	res, err := client.Get(componentURL)
	if err != nil {
		http.Error(w, "Server Error", http.StatusBadGateway)
		msg := fmt.Sprintf(`{"error": "Request error", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}
	defer func() { _ = res.Body.Close() }()

	for header := range res.Header {
		if err := safeCopyHeader(res.Header, header, w.Header()); err != nil {
			http.Error(w, "Server Error", http.StatusInternalServerError)
			msg := fmt.Sprintf(`{"error": "Header error", "originalError": "%v"}`, err)
			_, _ = w.Write([]byte(msg))
			return
		}
	}

	w.WriteHeader(res.StatusCode)
	_, _ = io.Copy(w, res.Body)
}

func (p *Plugin) parabolRedirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	path := vars["path"]
	config := p.getConfiguration()
	redirectURL := config.ParabolURL + "/" + path
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func commandsEqual(a, b []SlashCommand) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Trigger != b[i].Trigger || a[i].Description != b[i].Description {
			return false
		}
	}
	return true
}

func (p *Plugin) connect(c *Context, w http.ResponseWriter, r *http.Request) {
	var config ClientConfig
	if err := getJSON(r.Body, &config); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg := fmt.Sprintf(`{"error": "Error parsing commands", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}
	if !commandsEqual(p.commands, config.Commands) {
		p.commands = config.Commands
		if err := p.registerCommands(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			msg := fmt.Sprintf(`{"error": "Error registering commands", "originalError": "%v"}`, err)
			_, _ = w.Write([]byte(msg))
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

// initRouter initializes the HTTP router for the plugin.
func (p *Plugin) initRouter() *mux.Router {
	router := mux.NewRouter()

	// Notifications: Parabol posts here with Bearer token auth (verified via ParabolToken HMAC)
	router.HandleFunc("/notify/{channelID}", p.fixedPath(p.notify)).Methods("POST")
	// OAuth: redirects to Parabol's /oauth/authorize so users can log in
	router.HandleFunc("/auth", p.authenticated(p.auth)).Methods("GET")
	// Channel linking: generates a signed channel token for the JS plugin to pass to Parabol
	router.HandleFunc("/link", p.authenticated(p.link)).Methods("GET")
	router.HandleFunc("/graphql", p.graphql).Methods("POST")
	router.HandleFunc("/connect", p.authenticated(p.connect)).Methods("POST")
	router.HandleFunc("/config", p.authenticated(p.getConfig)).Methods("GET")
	router.HandleFunc("/components/{file}", p.components).Methods("GET")
	router.HandleFunc("/parabol/{path...}", p.parabolRedirect).Methods("GET")

	return router
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
// The root URL is currently <siteUrl>/plugins/com.mattermost.plugin-starter-template/api/v1/. Replace com.mattermost.plugin-starter-template with the plugin ID.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}
