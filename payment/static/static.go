package static

//go:generate go-bindata-assetfs -o web_static.go web/...

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/agreyfox/eshop/system/logs"

	"github.com/go-zoo/bone"
	"go.uber.org/zap"
)

const version = "v0.5"

var (
	logger   *zap.SugaredLogger = logs.Log.Sugar()
	showHelp bool

	port       string
	staticPath string
)

func usage(appName, version string) {
	fmt.Printf("Usage: %s [OPTIONS] [DB_NAME]", appName)
	fmt.Printf("\nOPTIONS:\n\n")
	flag.VisitAll(func(f *flag.Flag) {
		if len(f.Name) > 1 {
			fmt.Printf("    -%s, -%s\t%s\n", f.Name[0:1], f.Name, f.Usage)
		}
	})
	fmt.Printf("\n\nVersion %s\n", version)
}

// Start to mdb, rdb *bolt.DB, mainMux *bone.Mux
func Start(mainMux *bone.Mux) {

	logger.Info("starting static page payment  service...")
	initStatic()
	pwd, erro := os.Getwd()
	if erro != nil {
		logger.Error("Couldn't find current directory for file server.")
	}
	//CreateTest()

	logger.Info("Initial static payment environment")

	boltMux := bone.New() //.Prefix("admin")
	boltMux.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome static page service"))
	}))

	boltMux.HandleFunc("/", Index)
	boltMux.PostFunc("/dopay", userSubmit)

	//boltMux.HandleFunc("/notify", http.HandlerFunc(Notify))

	boltMux.HandleFunc("/return", http.HandlerFunc(Succeed))
	boltMux.HandleFunc("/cancel", Failed)
	boltMux.HandleFunc("/pay", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pwd, _ := os.Getwd()
		logger.Debug("Welcome static page service:" + pwd)
		payfile := pwd + "/pages/pay.html"
		http.ServeFile(w, r, payfile)

	}))
	//boltMux.Handle("//", http.StripPrefix("/static/", db.CacheControl(http.FileServer(restrict(http.Dir(staticDir))))))
	pageDir := filepath.Join(pwd, "pages")

	mainMux.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(pageDir))))
	mainMux.Get("/thanks", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Thanks! WelCome static still"))
	}))
	mainMux.SubRoute("/payment/static", boltMux)
	//mainMux.SubRoute("/payment/static,static", boltMux)
}
