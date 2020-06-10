//
// boltdbweb is a webserver base GUI for interacting with BoltDB databases.
//
// For authorship see https://github.com/evnix/boltdbweb
// MIT license is included in repository
//
package payment

//go:generate go-bindata-assetfs -o web_static.go web/...

import (
	"flag"
	"fmt"
	"github.com/agreyfox/eshop/payment/data"
	"github.com/agreyfox/eshop/payment/paypal"
	"github.com/agreyfox/eshop/payment/payssion"
	"github.com/agreyfox/eshop/payment/skrill"
	"github.com/agreyfox/eshop/system/logs"
	"github.com/go-zoo/bone"
	"go.uber.org/zap"
	"os"

	"github.com/boltdb/bolt"
)

const version = "v0.1.0"

var (
	logger   *zap.SugaredLogger = logs.Log.Sugar()
	showHelp bool

	port       string
	staticPath string
	mainMux    *bone.Mux
	initial    bool
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
	initial = false
}

func InitialPayment(db *bolt.DB, mux *bone.Mux) {
	logger.Info("Prepare Eshop Payment service environnement...")

	pwd, erro := os.Getwd()
	if erro != nil {
		logger.Error("Couldn't find current directory for file server.")
	}

	logger.Info("Initial eshop payment record data file ")

	data.SystemDBHandler = db // keep main db
	mainMux = mux

	var err error

	data.PaymentDBHandler, err = bolt.Open(pwd+"/"+data.DbFile, 0666, nil)
	if err != nil {
		logger.Fatal(err)
	}

	err = data.PaymentDBHandler.Update(func(tx *bolt.Tx) error {
		// initialize db with all content type buckets & sorted bucket for type
		_, err := tx.CreateBucketIfNotExists([]byte(data.DBName))
		if err != nil {
			logger.Debug("Error in check Record db")
			return err
		}
		return nil
	})

	if err != nil {
		logger.Fatal("initialize payment record db with buckets.Please check!", err)
	}
	initial = true

}

// 枢纽，跑不同的内容
func Run(serviceName string) {

	//	initpaypal() // repalce to not use default initial
	switch serviceName {
	case "paypal":
		paypal.Start(mainMux)
	case "payssion":
		payssion.Start(mainMux)
	case "skrill":
		skrill.Start(mainMux)
	default:
		logger.Fatal(" Wrong payment service name!")
	}

}

//for other module to access db handler.
func GetDBHandler() *bolt.DB {
	if initial {
		return data.PaymentDBHandler
	} else {
		logger.Debug("You need should initialize first ")
		return nil
	}
}
