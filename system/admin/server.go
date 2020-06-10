package admin

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/agreyfox/eshop/system"
	"github.com/agreyfox/eshop/system/admin/user"
	"github.com/agreyfox/eshop/system/api"
	"github.com/agreyfox/eshop/system/db"
	"github.com/agreyfox/eshop/system/logs"
	"github.com/go-zoo/bone"
	"go.uber.org/zap"
)

var (
	err    error
	logger *zap.SugaredLogger = logs.Log.Sugar()
)

// Run adds Handlers to default http listener for Admin
func Run(mainMux *bone.Mux) {
	//mainMux := bone.New()
	logger.Debug("Start admin interface")
	adminMux := bone.New() //.Prefix("admin")
	adminMux.HandleFunc("/", adminHandler)
	adminMux.HandleFunc("/init", initHandler)

	adminMux.HandleFunc("/login", loginHandler)
	adminMux.HandleFunc("/logout", logoutHandler)

	adminMux.HandleFunc("/recover", forgotPasswordHandler)
	adminMux.HandleFunc("/recover/key", recoveryKeyHandler)

	adminMux.HandleFunc("/addons", user.Auth(addonsHandler))
	adminMux.HandleFunc("/addon", user.Auth(addonHandler))

	adminMux.HandleFunc("/configure", user.Auth(configHandler))
	adminMux.HandleFunc("/configure/users", user.Auth(configUsersHandler))
	adminMux.HandleFunc("/configure/users/edit", user.Auth(configUsersEditHandler))
	adminMux.HandleFunc("/configure/users/delete", user.Auth(configUsersDeleteHandler))

	adminMux.HandleFunc("/uploads", user.Auth(uploadContentsHandler))
	adminMux.HandleFunc("/uploads/search", user.Auth(uploadSearchHandler))

	adminMux.HandleFunc("/contents", user.Auth(contentsHandler))
	adminMux.HandleFunc("/contents/search", user.Auth(searchHandler))
	adminMux.HandleFunc("/contents/export", user.Auth(exportHandler))

	adminMux.HandleFunc("/edit", user.Auth(editHandler))
	adminMux.HandleFunc("/edit/delete", user.Auth(deleteHandler))
	adminMux.HandleFunc("/edit/approve", user.Auth(approveContentHandler))
	adminMux.HandleFunc("/edit/upload", user.Auth(editUploadHandler))
	adminMux.HandleFunc("/edit/upload/delete", user.Auth(deleteUploadHandler))
	// Database & uploads backup via HTTP route registered with Basic Auth middleware.
	adminMux.HandleFunc("/backup", system.BasicAuth(backupHandler))

	/* http.HandleFunc("/admin", user.Auth(adminHandler))

	http.HandleFunc("/admin/init", initHandler)

	http.HandleFunc("/admin/login", loginHandler)
	http.HandleFunc("/admin/logout", logoutHandler)

	http.HandleFunc("/admin/recover", forgotPasswordHandler)
	http.HandleFunc("/admin/recover/key", recoveryKeyHandler)

	http.HandleFunc("/admin/addons", user.Auth(addonsHandler))
	http.HandleFunc("/admin/addon", user.Auth(addonHandler))

	http.HandleFunc("/admin/configure", user.Auth(configHandler))
	http.HandleFunc("/admin/configure/users", user.Auth(configUsersHandler))
	http.HandleFunc("/admin/configure/users/edit", user.Auth(configUsersEditHandler))
	http.HandleFunc("/admin/configure/users/delete", user.Auth(configUsersDeleteHandler))

	http.HandleFunc("/admin/uploads", user.Auth(uploadContentsHandler))
	http.HandleFunc("/admin/uploads/search", user.Auth(uploadSearchHandler))

	http.HandleFunc("/admin/contents", user.Auth(contentsHandler))
	http.HandleFunc("/admin/contents/search", user.Auth(searchHandler))
	http.HandleFunc("/admin/contents/export", user.Auth(exportHandler))

	http.HandleFunc("/admin/edit", user.Auth(editHandler))
	http.HandleFunc("/admin/edit/delete", user.Auth(deleteHandler))
	http.HandleFunc("/admin/edit/approve", user.Auth(approveContentHandler))
	http.HandleFunc("/admin/edit/upload", user.Auth(editUploadHandler))
	http.HandleFunc("/admin/edit/upload/delete", user.Auth(deleteUploadHandler)) */

	pwd, err := os.Getwd()
	if err != nil {
		logger.Error("Couldn't find current directory for file server.")
	}

	staticDir := filepath.Join(pwd, "static")

	logger.Infof("Server static  root is %s\n", staticDir)

	adminMux.Handle("/static/", http.StripPrefix("/static/", db.CacheControl(http.FileServer(restrict(http.Dir(staticDir))))))
	pageDir := filepath.Join(pwd, "pages")

	v1Mux := bone.New()
	//v1Mux.HandleFunc("/login", loginRestHandler)
	v1Mux.Post("/login", http.HandlerFunc(login))
	//v1Mux.HandleFunc("/logout", http.HandlerFunc(logout))
	v1Mux.Post("/logout", http.HandlerFunc(logout))

	v1Mux.Post("/recover", http.HandlerFunc(recoverRequest))
	v1Mux.Post("/recover/key", http.HandlerFunc(recoverPassword))
	v1Mux.Post("/backup", http.HandlerFunc(backup))

	//v1Mux.HandleFunc("/recover", forgotPasswordRestHandler)
	//v1Mux.HandleFunc("/recover/key", recoveryKeyRestHandler)

	v1Mux.HandleFunc("/addons", user.Auth(addonsRestHandler))
	v1Mux.HandleFunc("/addon", user.Auth(addonRestHandler))

	v1Mux.Get("/config", http.HandlerFunc(getConfig))
	v1Mux.Post("/config", http.HandlerFunc(saveConfig))

	//v1Mux.HandleFunc("/configure", user.Auth(configRestHandler))
	//v1Mux.HandleFunc("/configure/users", user.Auth(configUsersRestHandler))
	//v1Mux.HandleFunc("/configure/users/edit", user.Auth(configUsersEditRestHandler))
	//v1Mux.HandleFunc("/configure/users/delete", user.Auth(configUsersDeleteRestHandler))
	v1Mux.Get("/files", http.HandlerFunc(getMediaContents))
	//v1Mux.HandleFunc("/uploads", user.Auth(uploadContentsRestHandler))
	//v1Mux.HandleFunc("/uploads/search", user.Auth(uploadSearchRestHandler))
	v1Mux.Get("/file", http.HandlerFunc(getMedia))
	v1Mux.Delete("/file", http.HandlerFunc(deleteMediaContent))
	v1Mux.Post("/file", http.HandlerFunc(uploadMediaContent))
	v1Mux.Get("/files/search", http.HandlerFunc(searchMediaContent))

	//v1Mux.HandleFunc("/contents", user.Auth(contentsRestHandler))
	v1Mux.Get("/contents", http.HandlerFunc(getContents))
	v1Mux.Get("/contents/search", http.HandlerFunc(searchContent))
	v1Mux.Get("/contents/export", http.HandlerFunc(export))

	//v1Mux.HandleFunc("/contents/search", user.Auth(searchRestHandler))

	//v1Mux.HandleFunc("/contents/export", user.Auth(exportRestHandler))
	v1Mux.Post("/content", user.Auth(createContent))
	v1Mux.Post("/content/update", http.HandlerFunc(updateContent))
	//v1Mux.HandleFunc("/edit", user.Auth(editRestHandler))
	//v1Mux.HandleFunc("/edit/delete", user.Auth(deleteRestHandler))
	v1Mux.Get("/content", http.HandlerFunc(getContent))
	//v1Mux.HandleFunc("/content/delete", http.HandlerFunc(deleteContent))
	v1Mux.Post("/content/approve", http.HandlerFunc(approveContent))
	v1Mux.Post("/content/reject", http.HandlerFunc(rejectContent))
	v1Mux.Delete("/content", http.HandlerFunc(deleteContent))

	//v1Mux.HandleFunc("/edit/approve", user.Auth(approveContentRestHandler))
	//v1Mux.HandleFunc("/edit/upload", user.Auth(editUploadRestHandler))
	//v1Mux.HandleFunc("/edit/upload/delete", user.Auth(deleteUploadRestHandler))

	logger.Debugf("Magement server   root is %s\n", pageDir)

	mainMux.Handle("/mgt/", http.StripPrefix("/mgt", http.FileServer(restrict(http.Dir(pageDir)))))

	//http.Handle("/admin/static/", http.StripPrefix("/admin/static/", http.FileServer(http.Dir(staticDir))))
	// API path needs to be registered within server package so that it is handled
	// even if the API server is not running. Otherwise, images/files uploaded
	// through the editor will not load within the admin system.
	uploadsDir := filepath.Join(pwd, "uploads")
	mainMux.Handle("/api/uploads/", api.Record(api.CORS(db.CacheControl(http.StripPrefix("/api/uploads", http.FileServer(restrict(http.Dir(uploadsDir))))))))

	adminMux.SubRoute("/v1", v1Mux)
	mainMux.SubRoute("/admin", adminMux)

	logger.Debug("Start admin rest interface")

}

// Docs adds the documentation file server to the server, accessible at
// http://localhost:1234 by default
func Docs(port int) {
	pwd, err := os.Getwd()
	if err != nil {
		logger.Fatal("Couldn't find current directory for file server.")
	}

	docsDir := filepath.Join(pwd, "docs", "build")

	addr := fmt.Sprintf(":%d", port)
	url := fmt.Sprintf("http://localhost%s", addr)

	fmt.Println("")
	fmt.Println("View documentation offline at:", url)
	fmt.Println("")

	go http.ListenAndServe(addr, http.FileServer(http.Dir(docsDir)))
}
