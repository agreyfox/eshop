//
// boltdbweb is a webserver base GUI for interacting with BoltDB databases.
//
// For authorship see https://github.com/evnix/boltdbweb
// MIT license is included in repository
//
package payssion

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
	//dbHandler     *bolt.DB
	//tempDBHandler *bolt.DB
	//tempDBName    string = "payments"
	//dbName               = "Order"
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
	//logger.Info("init Payssion beckend  web interface")
}

/*
	key, err := db.GetParameterFromConfig("PaymentSetting", "name", "payssion_returnURL", "valueString")
	if err == nil {
		returnURL = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "payssion_cancelURL", "valueString")
	if err == nil {
		cancelURL = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "notify_email", "valueString")
	if err == nil {
		emailURL = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "payssion_apikey", "valueString")
	if err == nil {
		APIKey = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "payssion_SecretKey", "valueString")
	if err == nil {
		SecretKey = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "payssion_notifyURL", "valueString")
	if err == nil {
		notifyURL = key
	}
	key, err = db.GetParameterFromConfig("PaymentSetting", "name", "paysession_apibase", "valueString")
	if err == nil {
		apibase = key
	}

*/

func Start(mainMux *bone.Mux) {

	logger.Info("Initial payssion payment environment")

	//dbHandler = mdb // keep main db

	//tempDBHandler = rdb
	initPayssion()
	boltMux := bone.New() //.Prefix("admin")
	boltMux.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome payssion"))
	}))

	boltMux.HandleFunc("/", Index)

	boltMux.PostFunc("/pay", createPayment)
	boltMux.PostFunc("/dopay", userSubmit)
	boltMux.HandleFunc("/notify", http.HandlerFunc(Notify))

	boltMux.HandleFunc("/return", http.HandlerFunc(Succeed))
	boltMux.HandleFunc("/cancel", Failed)

	pwd, erro := os.Getwd()
	if erro != nil {
		logger.Error("Couldn't find current directory for file server.")
	}

	//boltMux.Handle("/web/", http.StripPrefix("/static/", db.CacheControl(http.FileServer(restrict(http.Dir(staticDir))))))
	pageDir := filepath.Join(pwd, "payment/payssion", "web")

	mainMux.Get("/payssion/*", http.StripPrefix("/payssion/", http.FileServer(http.Dir(pageDir))))
	mainMux.Get("/thanks", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Thanks! Welcome"))
	}))
	mainMux.SubRoute("/payment/payssion", boltMux)
	//mainMux.SubRoute("/payment/payssion,payssion", boltMux)
}
