package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/agreyfox/eshop/system/logs"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var (
	bind      string
	httpsport int
	port      int
	docsport  int
	https     bool
	devhttps  bool
	docs      bool
	cli       bool

	gocmd string
	fork  string
	dev   bool

	year = fmt.Sprintf("%d", time.Now().Year())

	logger *zap.SugaredLogger = logs.Log.Sugar()
)

func init() {
	//cobra.OnInitialize(initConfig)

	rootCmd.SetVersionTemplate("lqcms version {{printf \"%s\" .Version}}\n")
	fmt.Println("\t\t========================================")
	fmt.Printf("\t\t\tlqcms engine starting.......\n")
	fmt.Println("\t\t========================================")
	pflags := rootCmd.PersistentFlags()
	pflags.StringVar(&gocmd, "gocmd", "go", "custom go command if using beta or new release of Go")

}

func addServerFlags(flags *pflag.FlagSet) {
	flags.StringP("address", "a", "127.0.0.1", "address to listen on")
	flags.StringP("log", "l", "stdout", "log output")
	flags.StringP("port", "p", "8080", "port to listen on")
	flags.StringP("cert", "t", "", "tls certificate")
	flags.StringP("key", "k", "", "tls key")
	flags.StringP("root", "r", ".", "root to prepend to relative paths")
	flags.String("socket", "", "socket to listen to (cannot be used with address, port, cert nor key flags)")
	flags.StringP("baseurl", "b", "", "base url")
}

// Execute executes the commands.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
