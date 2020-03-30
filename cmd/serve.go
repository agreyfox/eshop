package cmd

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/agreyfox/eshop/content"
	"github.com/agreyfox/eshop/system/admin"
	"github.com/agreyfox/eshop/system/api"
	"github.com/agreyfox/eshop/system/api/analytics"
	"github.com/agreyfox/eshop/system/db"
	"github.com/agreyfox/eshop/system/tls"
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

		analytics.Init()
		defer analytics.Close()

		services := strings.Split(args[0], ",")
		fmt.Println("Service is ", services)
		for _, service := range services {
			if service == "api" {
				api.Run()
			} else if service == "admin" {
				admin.Run()
			} else {
				return ErrWrongOrMissingService
			}
		}

		// run docs server if --docs is true
		if docs {
			admin.Docs(docsport)
		}

		// init search index
		go db.InitSearchIndex()

		// save the https port the system is listening on
		err := db.PutConfig("https_port", fmt.Sprintf("%d", httpsport))
		if err != nil {
			log.Fatalln("System failed to save config. Please try to run again.", err)
		}

		// cannot run production HTTPS and development HTTPS together
		if devhttps {
			fmt.Println("Enabling self-signed HTTPS... [DEV]")

			go tls.EnableDev()
			fmt.Println("Server listening on https://localhost:10443 for requests... [DEV]")
			fmt.Println("----")
			fmt.Println("If your browser rejects HTTPS requests, try allowing insecure connections on localhost.")
			fmt.Println("on Chrome, visit chrome://flags/#allow-insecure-localhost")

		} else if https {
			fmt.Println("Enabling HTTPS...")

			go tls.Enable()
			fmt.Printf("Server listening on :%s for HTTPS requests...\n", db.ConfigCache("https_port").(string))
		}

		// save the https port the system is listening on so internal system can make
		// HTTP api calls while in dev or production w/o adding more cli flags
		err = db.PutConfig("http_port", fmt.Sprintf("%d", port))
		if err != nil {
			log.Fatalln("System failed to save config. Please try to run again.", err)
		}

		// save the bound address the system is listening on so internal system can make
		// HTTP api calls while in dev or production w/o adding more cli flags
		if bind == "" {
			bind = "localhost"
		}
		err = db.PutConfig("bind_addr", bind)
		if err != nil {
			log.Fatalln("System failed to save config. Please try to run again.", err)
		}

		fmt.Printf("Server listening at %s:%d for HTTP requests...\n", bind, port)
		fmt.Println("\nVisit '/admin' to get started.")
		log.Fatalln(http.ListenAndServe(fmt.Sprintf("%s:%d", bind, port), nil))
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
