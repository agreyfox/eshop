// Package api sets the various API handlers which provide an HTTP interface to
// Ponzu content, and include the types and interfaces to enable client-side
// interactivity with the system.
package api

import (
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
	logger.Debug("Start content api interface")
	apiv1Mux := bone.New().Prefix("/v1")
	//apiv1Mux.HandleFunc("/contents", Record(CORS(Gzip(contentsHandler))))
	apiv1Mux.Get("/contents", Record(CustomerAuth(CORS(Gzip(contents)))))

	//apiv1Mux.HandleFunc("/content", Record(CORS(Gzip(contentHandler))))
	apiv1Mux.Get("/content", Record(CustomerAuth(CORS(Gzip(content)))))
	apiv1Mux.Post("/content/create", Record(CustomerAuth(CORS(createContent))))

	//apiv1Mux.HandleFunc("/content/create", Record(CORS(createContentHandler)))

	apiv1Mux.HandleFunc("/content/update", Record(CORS(updateContentHandler)))

	apiv1Mux.HandleFunc("/content/delete", Record(CORS(deleteContentHandler)))

	apiv1Mux.Get("/search", Record(CustomerAuth(CORS(Gzip(searchContent)))))

	//apiv1Mux.HandleFunc("/search", Record(CustomerAuth(CORS(Gzip(searchContentHandler)))))

	//apiv1Mux.HandleFunc("/uploads", Record(CustomerAuth(CORS(Gzip(uploadsHandler)))))
	apiv1Mux.Get("/files", Record(CustomerAuth(CORS(Gzip(uploads)))))
	apiv1Mux.Get("/pics", Record(CORS(Gzip(getMedia))))

	//apiv1Mux.HandleFunc("/user/register", CORS(user.RegisterUsersHandler))
	apiv1Mux.Post("/register", CORS(RegisterUser))
	apiv1Mux.Get("/renew", CORS(Renew))
	apiv1Mux.Post("/logout", CustomerAuth(Logout))
	apiv1Mux.Post("/forgot", CORS(Forgot))
	apiv1Mux.Post("/login", CORS(Login))
	apiv1Mux.Post("/recovery", CORS(Recovery))

	apiv1Mux.HandleFunc("/user/login", CORS(LoginHandler))
	apiv1Mux.HandleFunc("/user/renew", RenewHandler)
	apiv1Mux.HandleFunc("/user/logout", CustomerAuth(LogoutHandler))
	apiv1Mux.HandleFunc("/user/forgot", CustomerAuth(ForgotPasswordHandler))
	apiv1Mux.HandleFunc("/user/recovery", RecoveryKeyHandler)
	mainMux.SubRoute("/api", apiv1Mux)
}
