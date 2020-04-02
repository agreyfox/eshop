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
	"go.uber.org/zap"
)

var (
	err    error
	logger *zap.SugaredLogger = logs.Log.Sugar()
)

// Run adds Handlers to default http listener for Admin
func Run() {
	logger.Debug("Start admin interface")
	http.HandleFunc("/admin", user.Auth(adminHandler))

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
	http.HandleFunc("/admin/edit/upload/delete", user.Auth(deleteUploadHandler))

	pwd, err := os.Getwd()
	if err != nil {
		logger.Error("Couldn't find current directory for file server.")
	}

	staticDir := filepath.Join(pwd, "static")

	logger.Infof("Server static  root is %s\n", staticDir)

	http.Handle("/admin/static/", http.StripPrefix("/admin/static/", db.CacheControl(http.FileServer(restrict(http.Dir(staticDir))))))
	pageDir := filepath.Join(pwd, "pages")

	logger.Debugf("Magement server   root is %s\n", pageDir)

	http.Handle("/admin/pages/", http.StripPrefix("/admin/pages/", http.FileServer(restrict(http.Dir(pageDir)))))

	//http.Handle("/admin/static/", http.StripPrefix("/admin/static/", http.FileServer(http.Dir(staticDir))))
	// API path needs to be registered within server package so that it is handled
	// even if the API server is not running. Otherwise, images/files uploaded
	// through the editor will not load within the admin system.
	uploadsDir := filepath.Join(pwd, "uploads")
	http.Handle("/api/uploads/", api.Record(api.CORS(db.CacheControl(http.StripPrefix("/api/uploads/", http.FileServer(restrict(http.Dir(uploadsDir))))))))

	// Database & uploads backup via HTTP route registered with Basic Auth middleware.
	http.HandleFunc("/admin/backup", system.BasicAuth(backupHandler))
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