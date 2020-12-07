//
// boltdbweb is a webserver base GUI for interacting with BoltDB databases.
//
// For authorship see https://github.com/evnix/boltdbweb
// MIT license is included in repository
//
package paypal

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

const version = "v0.2.0"

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

func init() {
	//logger.Info("init Paypal beckend  web interface")
}

func Start(mainMux *bone.Mux) {

	logger.Info("starting PayPal beckend service...")
	initpaypal() // repalce to not use default initial

	boltMux := bone.New() //.Prefix("admin")
	boltMux.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	}))

	boltMux.HandleFunc("/", Index)
	boltMux.PostFunc("/dopay", userSubmit)
	boltMux.PostFunc("/pay", createPayment)
	boltMux.HandleFunc("/return", Succeed)
	boltMux.PostFunc("/Failed", Failed)
	boltMux.HandleFunc("/notify", Notify)
	boltMux.HandleFunc("/cancel", Failed)
	boltMux.HandleFunc("/test", Test)
	boltMux.Get("/info/:id", http.HandlerFunc(TransactionInfo))

	pwd, erro := os.Getwd()
	if erro != nil {
		logger.Error("Couldn't find current directory for file server.")
	}

	//boltMux.Handle("/web/", http.StripPrefix("/static/", db.CacheControl(http.FileServer(restrict(http.Dir(staticDir))))))
	pageDir := filepath.Join(pwd, "payment/paypal", "web")

	mainMux.Get("/paypal/*", http.StripPrefix("/paypal/", http.FileServer(http.Dir(pageDir))))
	mainMux.Get("/thanks", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Thanks! WelCome "))
	}))
	mainMux.SubRoute("/payment/paypal", boltMux)
	//mainMux.SubRoute("/payment/paypal,paypal", boltMux)

}
