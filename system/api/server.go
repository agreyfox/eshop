// Package api sets the various API handlers which provide an HTTP interface to
// Ponzu content, and include the types and interfaces to enable client-side
// interactivity with the system.
package api

import (
	"github.com/agreyfox/eshop/system/api/user"
	"github.com/agreyfox/eshop/system/logs"
	"github.com/go-zoo/bone"
	"go.uber.org/zap"
)

var (
	err    error
	logger *zap.SugaredLogger = logs.Log.Sugar()
)

// Run adds Handlers to default http listener for API
func Run(mainMux *bone.Mux) {
	logger.Debug("Start api interface")
	apiv1Mux := bone.New().Prefix("/v1")
	apiv1Mux.HandleFunc("/contents", Record(CORS(Gzip(contentsHandler))))

	apiv1Mux.HandleFunc("/content", Record(CORS(Gzip(contentHandler))))

	apiv1Mux.HandleFunc("/content/create", Record(CORS(createContentHandler)))

	apiv1Mux.HandleFunc("/content/update", Record(CORS(updateContentHandler)))

	apiv1Mux.HandleFunc("/content/delete", Record(CORS(deleteContentHandler)))

	apiv1Mux.HandleFunc("/search", Record(user.CustomerAuth(CORS(Gzip(searchContentHandler)))))

	apiv1Mux.HandleFunc("/uploads", Record(user.CustomerAuth(CORS(Gzip(uploadsHandler)))))

	apiv1Mux.HandleFunc("/user/register", CORS(user.RegisterUsersHandler))
	apiv1Mux.HandleFunc("/user/login", CORS(user.LoginHandler))
	apiv1Mux.HandleFunc("/user/renew", user.RenewHandler)

	apiv1Mux.HandleFunc("/user/logout", user.CustomerAuth(user.LogoutHandler))
	apiv1Mux.HandleFunc("/user/forgot", user.CustomerAuth(user.ForgotPasswordHandler))
	apiv1Mux.HandleFunc("/user/recovery", user.RecoveryKeyHandler)
	mainMux.SubRoute("/api", apiv1Mux)
}
