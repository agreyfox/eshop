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

	"github.com/agreyfox/eshop/boltdbweb/web"
	"github.com/agreyfox/eshop/system/logs"

	"github.com/go-zoo/bone"
	"go.uber.org/zap"

	"github.com/boltdb/bolt"
)

const version = "v0.1.0"

var (
	logger     *zap.SugaredLogger = logs.Log.Sugar()
	showHelp   bool
	dbHandler  *bolt.DB
	dbName    ="paypal"
	dbFilename = "record.db"
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
	logger.Info("init Paypal beckend  web interface")
}

func Run(mainMux *bone.Mux) {

	logger.Info("starting PayPal beckend service...")

	pwd, erro := os.Getwd()
	if erro != nil {
		logger.Error("Couldn't find current directory for file server.")
	}

	logger.Info("Initial paypal payment record data file ")
	if store != nil {
		return
	}

	var err error
	store, err = bolt.Open("system.db", 0666, nil)
	if dbHandler != nil {
		logger.Fatal(err)
		
	}

	err = store.Update(func(tx *bolt.Tx) error {
		// initialize db with all content type buckets & sorted bucket for type
		_,err:=tx.CreateBucketIfNotExists([]byte(dbName))
		if err!=nil{
			Logger.Debugf("Error in check Record db")
			return 
		}
		
	})

	if err != nil {
		logger.Fatal("initialize db with buckets.Please check ", err)
	}


	boltMux := bone.New() //.Prefix("admin")
	mainMux.Get("/paypalping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("paypalpong"))
	}))

	boltMux.GetFunc("/", web.Index)

	mainMux.PostFunc("/pay", createPayment)
	mainMux.PostFunc("/Succeed", Succeed)
	mainMux.PostFunc("/Failed", Failed)
	mainMux.PostFunc("/return", Succeed)
	mainMux.PostFunc("/cancle", Failed)

	//boltMux.Handle("/web/", http.StripPrefix("/static/", db.CacheControl(http.FileServer(restrict(http.Dir(staticDir))))))
	pageDir := filepath.Join(pwd, "payment/papyal", "web")

	mainMux.Get("/paypal/*", http.StripPrefix("/paypal/", http.FileServer(http.Dir(pageDir))))

	mainMux.SubRoute("/payment/paypal", boltMux)

}

/*
func main() {
	appName := path.Base(os.Args[0])
	flag.Parse()
	args := flag.Args()

	if showHelp == true {
		usage(appName, version)
		os.Exit(0)
	}

	// If non-flag options are included assume bolt db is specified.
	if len(args) > 0 {
		dbName = args[0]
	}

	if dbName == "" {
		usage(appName, version)
		log.Printf("\nERROR: Missing boltdb name\n")
		os.Exit(1)
	}

	fmt.Print(" ")
	log.Info("starting boltdb-browser..")

	var err error
	db, err = bolt.Open(dbName, 0600, &bolt.Options{Timeout: 2 * time.Second})
	boltbrowserweb.Db = db

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// OK, we should be ready to define/run web server safely.
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/", boltbrowserweb.Index)

	r.GET("/buckets", boltbrowserweb.Buckets)
	r.POST("/createBucket", boltbrowserweb.CreateBucket)
	r.POST("/put", boltbrowserweb.Put)
	r.POST("/get", boltbrowserweb.Get)
	r.POST("/deleteKey", boltbrowserweb.DeleteKey)
	r.POST("/deleteBucket", boltbrowserweb.DeleteBucket)
	r.POST("/prefixScan", boltbrowserweb.PrefixScan)

	r.StaticFS("/web", assetFS())

	r.Run(":" + port)
}
*/
