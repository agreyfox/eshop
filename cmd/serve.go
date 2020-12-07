package cmd

import (
	"errors"
	"fmt"

	"net/http"
	"strings"

	"github.com/agreyfox/eshop/boltdbweb"
	_ "github.com/agreyfox/eshop/content"
	"github.com/agreyfox/eshop/payment"
	"github.com/agreyfox/eshop/prometheus"
	"github.com/agreyfox/eshop/system/admin"
	"github.com/agreyfox/eshop/system/api"
	"github.com/agreyfox/eshop/system/api/analytics"
	"github.com/agreyfox/eshop/system/db"
	"github.com/agreyfox/eshop/system/tls"
	"github.com/go-zoo/bone"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
)

var ErrWrongOrMissingService = errors.New("To execute 'eshop serve', " +
	"you must specify which service to run.")

var serveCmd = &cobra.Command{
	Use:     "serve [flags] <service,service>",
	Aliases: []string{"s"},
	Short:   "run the server (serve is wrapped by the run command)",
	Hidden:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return ErrWrongOrMissingService
		}

		db.Init()
		defer db.Close()
		//db.PutConfig("Key", config.GenerateKey())

		analytics.Init()
		defer analytics.Close()

		services := strings.Split(args[0], ",")
		logger.Info("Start Service : ", services)

		mainMux := bone.New()
		for _, s := range services {
			if s == "paypal" || s == "payssion" || s == "skrill" {
				payment.InitialPayment(db.Store(), mainMux)
				break
			}
		}
		for _, service := range services {
			//fmt.Println(service)
			if service == "api" {
				api.Run(mainMux)
			} else if service == "admin" {
				admin.Run(mainMux)
			} else if service == "db" {
				boltdbweb.Run(db.Store(), mainMux) //run bolt db instance
			} else if service == "static" || service == "paypal" || service == "payssion" || service == "skrill" {
				payment.Run(service)
			} else if service == "monitor" {
				go prometheus.Run(":9001", mainMux)
			} else {
				return ErrWrongOrMissingService
			}
		}

		/* c := cors.New(cors.Options{
			AllowedOrigins:   []string{"https://support.bk.cloudns.cc", "http://127.0.0.1:8080"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "CONNECT", "HEAD"},
			AllowCredentials: true,
			AllowedHeaders:   []string{"Accept", "Content-Type", "Lqcms_token", "Content-Length", "Accept-Encoding", "Authorization", "X-CSRF-Token"},
			// Enable Debugging for testing, consider disabling in production
			Debug: true,
		}) */
		cmainMux := cors.AllowAll().Handler(mainMux)
		//cmainMux := c.Handler(mainMux)
		// run docs server if --docs is true
		if docs {
			admin.Docs(docsport)
		}

		// init search index
		go db.InitSearchIndex()

		// save the https port the system is listening on
		err := db.PutConfig("https_port", fmt.Sprintf("%d", httpsport))
		if err != nil {
			logger.Fatal("System failed to save config. Please try to run again.", err)
		}

		// cannot run production HTTPS and development HTTPS together
		if devhttps {
			logger.Info("Enabling self-signed HTTPS... [DEV]")

			go tls.EnableDev(cmainMux)
			logger.Info("Server listening on https://localhost:10443 for requests... [DEV]")

			logger.Info("If your browser rejects HTTPS requests, try allowing insecure connections on localhost.")
			logger.Info("on Chrome, visit chrome://flags/#allow-insecure-localhost")

		} else if https {
			logger.Info("Enabling HTTPS...")

			go tls.Enable(cmainMux)
			logger.Warnf("Server listening on :%s for HTTPS requests...\n", db.ConfigCache("https_port").(string))
		}

		// save the https port the system is listening on so internal system can make
		// HTTP api calls while in dev or production w/o adding more cli flags
		err = db.PutConfig("http_port", fmt.Sprintf("%d", port))
		if err != nil {
			logger.Fatalf("System failed to save config. Please try to run again.", err)
		}

		// save the bound address the system is listening on so internal system can make
		// HTTP api calls while in dev or production w/o adding more cli flags
		if bind == "" {
			bind = "localhost"
		}
		err = db.PutConfig("bind_addr", bind)
		if err != nil {
			logger.Fatalf("System failed to save config. Please try to run again.", err)
		}

		logger.Infof("Server listening at %s:%d for HTTP requests...\n", bind, port)
		logger.Info("\nVisit '/admin' to get started.")

		fmt.Println(http.ListenAndServe(fmt.Sprintf("%s:%d", bind, port), cmainMux))

		return nil
	},
}

func init() {
	serveCmd.Flags().StringVar(&bind, "bind", "localhost", "address for eshop to bind the HTTP(S) server")
	serveCmd.Flags().IntVar(&httpsport, "https-port", 443, "port for eshop to bind its HTTPS listener")
	serveCmd.Flags().IntVar(&port, "port", 8080, "port for shop to bind its HTTP listener")
	serveCmd.Flags().IntVar(&docsport, "docs-port", 1234, "[dev environment] override the documentation server port")
	serveCmd.Flags().BoolVar(&docs, "docs", false, "[dev environment] run HTTP server to view local HTML documentation")
	serveCmd.Flags().BoolVar(&https, "https", false, "enable automatic TLS/SSL certificate management")
	serveCmd.Flags().BoolVar(&devhttps, "dev-https", false, "[dev environment] enable automatic TLS/SSL certificate management")

	RegisterCmdlineCommand(serveCmd)
}
