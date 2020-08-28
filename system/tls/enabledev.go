package tls

import (
	"net/http"
	"os"
	"path/filepath"
)

// EnableDev generates self-signed SSL certificates to use HTTPS & HTTP/2 while
// working in a development environment. The certs are saved in a different
// directory than the production certs (from Let's Encrypt), so that the
// acme/autocert package doesn't mistake them for it's own.
// Additionally, a TLS server is started using the default http mux.
func EnableDev(mux http.Handler) {
	setupDev()

	pwd, err := os.Getwd()
	if err != nil {
		logger.Fatal("Couldn't find working directory to activate dev certificates:", err)
	}

	vendorPath := filepath.Join(pwd, "key") //the key store and key下面

	cert := filepath.Join(vendorPath, "devcerts", "cert.pem")
	key := filepath.Join(vendorPath, "devcerts", "key.pem")

	logger.Fatal(http.ListenAndServeTLS(":10443", cert, key, mux))
}
