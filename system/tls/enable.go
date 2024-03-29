// Package tls provides the functionality to Ponzu systems to encrypt HTTP traffic
// through the ability to generate self-signed certificates for local development
// and fetch/update production certificates from Let's Encrypt.
package tls

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/agreyfox/eshop/system/db"

	"golang.org/x/crypto/acme/autocert"
)

// newManager attempts to locate or create the cert cache directory and the
// certs for TLS encryption and returns an autocert.Manager
func newManager() autocert.Manager {
	pwd, err := os.Getwd()
	if err != nil {
		logger.Fatal("Couldn't find working directory to locate or save certificates.")
	}

	cache := autocert.DirCache(filepath.Join(pwd, "system", "tls", "certs"))
	if _, err := os.Stat(string(cache)); os.IsNotExist(err) {
		err := os.MkdirAll(string(cache), os.ModePerm|os.ModeDir)
		if err != nil {
			logger.Fatal("Couldn't create cert directory at", cache)
		}
	}

	// get host/domain and email from Config to use for TLS request to Let's encryption.
	// we will fail fatally if either are not found since Let's Encrypt will rate-limit
	// and sending incomplete requests is wasteful and guaranteed to fail its check
	host, err := db.Config("domain")
	if err != nil {
		logger.Fatal("Error identifying host/domain during TLS set-up.", err)
	}

	if host == nil {
		logger.Fatal("No 'domain' field set in Configuration. Please add a domain before attempting to make certificates.")
	}
	fmt.Println("Using", string(host), "as host/domain for certificate...")
	fmt.Println("NOTE: if the host/domain is not configured properly or is unreachable, HTTPS set-up will fail.")

	email, err := db.Config("admin_email")
	if err != nil {
		logger.Fatal("Error identifying admin email during TLS set-up.", err)
	}

	if email == nil {
		logger.Fatal("No 'admin_email' field set in Configuration. Please add an admin email before attempting to make certificates.")
	}
	fmt.Println("Using", string(email), "as contact email for certificate...")

	return autocert.Manager{
		Prompt:      autocert.AcceptTOS,
		Cache:       cache,
		HostPolicy:  autocert.HostWhitelist(string(host)),
		RenewBefore: time.Hour * 24 * 30,
		Email:       string(email),
	}
}

// Enable runs the setup for creating or locating production certificates and
// starts the TLS server
func Enable(mux http.Handler) {
	//m := newManager()
	/*
		server := &http.Server{
			Addr:      fmt.Sprintf(":%s", db.ConfigCache("https_port").(string)),
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		} */

	/* 	cert, err := tls.LoadX509KeyPair("key/prod/server.pem", "key/prod/server.key")
	   	if err != nil {
	   		log.Fatalln(err)
	   		return
	   	} */
	/* 	server := &http.Server{
	   		Addr:      fmt.Sprintf(":%s", db.ConfigCache("https_port").(string)),
	   		TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
	   	}
	*/
	// launch http listener for "http-01" ACME challenge
	//go http.ListenAndServe(":http", m.HTTPHandler(nil))

	//logger.Fatal(server.ListenAndServeTLS("", ""))
	po := fmt.Sprintf(":%s", db.ConfigCache("https_port").(string))
	logger.Warnf("Start https server at port:%s", po)
	logger.Fatal(http.ListenAndServeTLS(po, "key/prod/server.pem", "key/prod/server.key", mux))
}
