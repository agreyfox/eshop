// Package api sets the various API handlers which provide an HTTP interface to
// Ponzu content, and include the types and interfaces to enable client-side
// interactivity with the system.
package api

import (
	"net/http"

	"github.com/agreyfox/eshop/system/api/user"
	"github.com/agreyfox/eshop/system/logs"
	"go.uber.org/zap"
)

var (
	err    error
	logger *zap.SugaredLogger = logs.Log.Sugar()
)

// Run adds Handlers to default http listener for API
func Run() {
	logger.Debug("Start api interface")
	http.HandleFunc("/api/contents", Record(CORS(Gzip(contentsHandler))))

	http.HandleFunc("/api/content", Record(CORS(Gzip(contentHandler))))

	http.HandleFunc("/api/content/create", Record(CORS(createContentHandler)))

	http.HandleFunc("/api/content/update", Record(CORS(updateContentHandler)))

	http.HandleFunc("/api/content/delete", Record(CORS(deleteContentHandler)))

	http.HandleFunc("/api/search", Record(user.CustomerAuth(CORS(Gzip(searchContentHandler)))))

	http.HandleFunc("/api/uploads", Record(user.CustomerAuth(CORS(Gzip(uploadsHandler)))))

	http.HandleFunc("/api/user/register", user.RegisterUsersHandler)
	http.HandleFunc("/api/user/login", user.LoginHandler)

	http.HandleFunc("/api/user/logout", user.CustomerAuth(user.LogoutHandler))
	http.HandleFunc("/api/user/forgot", user.CustomerAuth(user.ForgotPasswordHandler))
	http.HandleFunc("/api/user/recovery", user.RecoveryKeyHandler)
}
