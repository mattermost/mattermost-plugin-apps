package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

const (
	AppSecret = "1234"

	CommandTrigger = "test"
)

var IncludeInvalid bool
var Log = utils.MustMakeCommandLogger(zapcore.DebugLevel)

var ErrUnexpectedSignMethod = errors.New("unexpected signing method")
var ErrMissingHeader = errors.Errorf("missing %s: Bearer header", apps.OutgoingAuthHeader)
var ErrActingUserMismatch = errors.New("JWT claim doesn't match actingUser.Id in context")

type callHandler func(*apps.CallRequest) apps.CallResponse

func main() {
	rootURL := os.Getenv("ROOT_URL")
	if rootURL == "" {
		rootURL = AppManifest.Deploy.HTTP.RootURL
	}

	if os.Getenv("INCLUDE_INVALID") != "" {
		IncludeInvalid = true
	}

	portStr := os.Getenv("PORT")
	if portStr == "" && rootURL != "" {
		// Get the port from the original manifest's root_url.
		u, err := url.Parse(rootURL)
		if err != nil {
			panic(err)
		}
		portStr = u.Port()
		if portStr == "" {
			portStr = "8080"
		}
	}

	listen := ":" + portStr
	if rootURL == "" {
		rootURL = "http://localhost" + listen
	}
	AppManifest.Deploy.HTTP.RootURL = rootURL

	r := mux.NewRouter()
	r.HandleFunc(ManifestPath, httputils.DoHandleJSON(AppManifest))
	r.PathPrefix(StaticPath).Handler(http.StripPrefix("/", http.FileServer(http.FS(StaticFS))))

	handleCall(r, InstallPath, handleInstall)
	handleCall(r, BindingsPath, handleBindings)
	initHTTPEmbedded(r)
	initHTTPError(r)
	initHTTPForms(r)
	initHTTPLookup(r)
	initHTTPNavigate(r)
	initHTTPOK(r)
	initNumBindingsCommand(r)
	initHTTPSubscriptions(r)

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := errors.Errorf("path not found: %s", r.URL.Path)
		_ = httputils.WriteJSON(w, apps.NewErrorResponse(err))
	})

	server := &http.Server{
		Addr:              listen,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      5 * time.Second,
		Handler:           r,
	}

	Log.Infof("test app started, listening on port %s, manifest at %s/manifest.json. Use %s as the JWT secret.", portStr, AppManifest.Deploy.HTTP.RootURL, AppSecret)
	panic(server.ListenAndServe())
}

func handle(f callHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		creq, err := apps.CallRequestFromJSONReader(r.Body)
		if err != nil {
			httputils.WriteErrorIfNeeded(w, utils.NewInvalidError(err))
			return
		}

		err = checkJWT(r, creq)
		if err != nil {
			httputils.WriteErrorIfNeeded(w, utils.NewInvalidError(err))
			return
		}

		cresp := f(creq)

		log := Log.With("req", creq.Path, "resp", cresp.Type)
		if creq.Context.Subject != "" {
			log = log.With("subject", creq.Context.Subject)
		}
		if len(creq.Values) > 0 {
			log = log.With("values", creq.Values)
		}
		if cresp.Text != "" {
			log = log.With("text", cresp.Text)
		}
		if cresp.Data != nil {
			log = log.With("data", cresp.Data)
		}
		if cresp.Form != nil {
			log = log.With("form", cresp.Form)
		}
		log.Debugw("processed call")

		_ = httputils.WriteJSON(w, cresp)
	}
}

func handleCall(router *mux.Router, path string, f callHandler) {
	router.HandleFunc(path, handle(f))
}

func handleError(text string) callHandler {
	return func(_ *apps.CallRequest) apps.CallResponse {
		return apps.CallResponse{
			Type: apps.CallResponseTypeError,
			Text: text,
		}
	}
}

func handleForm(f apps.Form) callHandler {
	return func(_ *apps.CallRequest) apps.CallResponse {
		return apps.NewFormResponse(f)
	}
}

type lookupResponse struct {
	Items []apps.SelectOption `json:"items"`
}

func handleLookup(items []apps.SelectOption) callHandler {
	return func(creq *apps.CallRequest) apps.CallResponse {
		query := strings.ToLower(creq.Query)
		finalItems := []apps.SelectOption{}

		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Label), query) {
				finalItems = append(finalItems, item)
			}
		}

		return apps.NewDataResponse(lookupResponse{finalItems})
	}
}

func checkJWT(req *http.Request, creq *apps.CallRequest) error {
	authValue := req.Header.Get(apps.OutgoingAuthHeader)
	if !strings.HasPrefix(authValue, "Bearer ") {
		return ErrMissingHeader
	}

	jwtoken := strings.TrimPrefix(authValue, "Bearer ")
	claims := apps.JWTClaims{}

	_, err := jwt.ParseWithClaims(jwtoken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSignMethod, token.Header["alg"])
		}
		return []byte(AppSecret), nil
	})
	if err != nil {
		return err
	}

	if creq.Context.ActingUser != nil && creq.Context.ActingUser.Id != claims.ActingUserID {
		return utils.NewInvalidError(ErrActingUserMismatch)
	}

	return nil
}
