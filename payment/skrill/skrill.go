package skrill

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

const version = "v0.1.0"

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

	logger.Info("starting skrill  service...")
	initSkrill()
	pwd, erro := os.Getwd()
	if erro != nil {
		logger.Error("Couldn't find current directory for file server.")
	}
	//CreateTest()

	logger.Info("Initial skrill payment environment")

	boltMux := bone.New() //.Prefix("admin")
	boltMux.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome skrill"))
	}))

	boltMux.HandleFunc("/", Index)
	boltMux.PostFunc("/dopay", userSubmit)
	boltMux.PostFunc("/pay", createPayment)
	boltMux.HandleFunc("/notify", http.HandlerFunc(Notify))

	boltMux.HandleFunc("/return", http.HandlerFunc(Succeed))
	boltMux.HandleFunc("/cancel", Failed)

	//boltMux.Handle("/web/", http.StripPrefix("/static/", db.CacheControl(http.FileServer(restrict(http.Dir(staticDir))))))
	pageDir := filepath.Join(pwd, "payment/skrill", "web")

	mainMux.Get("/skrill/*", http.StripPrefix("/skrill/", http.FileServer(http.Dir(pageDir))))
	mainMux.Get("/thanks", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Thanks! WelCome skrill"))
	}))
	mainMux.SubRoute("/payment/skrill", boltMux)
	//mainMux.SubRoute("/payment/skrill,skrill", boltMux)
}
