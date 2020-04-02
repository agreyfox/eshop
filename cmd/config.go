package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/agreyfox/management/db"

	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(initCmd)
	configCmd.AddCommand(allCmd)
	RegisterCmdlineCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management utility",
	Long:  `Configuration management utility.`,
	Args:  cobra.NoArgs,
}

var initCmd = &cobra.Command{
	Use: "config <init>...",
	//Aliases: []string{"c"},
	Short: "make default system configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return InitilizedDb()
	},
}
var allCmd = &cobra.Command{
	Use: "config <all>...",
	//Aliases: []string{"c"},
	Short: "List all the configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		DistplayAllConfig()
		return
	},
}

// manully initial db
func InitilizedDb() error {

	return nil
}

// to Display all the system configuration
func DistplayAllConfig() {
	logger.Info("Trying to get the all configuration in system db")
	config, err := db.ConfigAll()
	ma := map[string]interface{}
	err = json.Unmarshal(config,&ma)
	logger.Info("%v", ma,err)
}
/* 
func printSettings(ser *settings.Server, set *settings.Settings, auther auth.Auther) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Sign up:\t%t\n", set.Signup)
	fmt.Fprintf(w, "Create User Dir:\t%t\n", set.CreateUserDir)
	fmt.Fprintf(w, "Auth method:\t%s\n", set.AuthMethod)
	fmt.Fprintf(w, "Shell:\t%s\t\n", strings.Join(set.Shell, " "))
	fmt.Fprintln(w, "\nBranding:")
	fmt.Fprintf(w, "\tName:\t%s\n", set.Branding.Name)
	fmt.Fprintf(w, "\tFiles override:\t%s\n", set.Branding.Files)
	fmt.Fprintf(w, "\tDisable external links:\t%t\n", set.Branding.DisableExternal)
	fmt.Fprintln(w, "\nServer:")
	fmt.Fprintf(w, "\tLog:\t%s\n", ser.Log)
	fmt.Fprintf(w, "\tPort:\t%s\n", ser.Port)
	fmt.Fprintf(w, "\tBase URL:\t%s\n", ser.BaseURL)
	fmt.Fprintf(w, "\tRoot:\t%s\n", ser.Root)
	fmt.Fprintf(w, "\tSocket:\t%s\n", ser.Socket)
	fmt.Fprintf(w, "\tAddress:\t%s\n", ser.Address)
	fmt.Fprintf(w, "\tTLS Cert:\t%s\n", ser.TLSCert)
	fmt.Fprintf(w, "\tTLS Key:\t%s\n", ser.TLSKey)
	fmt.Fprintln(w, "\nDefaults:")
	fmt.Fprintf(w, "\tScope:\t%s\n", set.Defaults.Scope)
	fmt.Fprintf(w, "\tLocale:\t%s\n", set.Defaults.Locale)
	fmt.Fprintf(w, "\tView mode:\t%s\n", set.Defaults.ViewMode)
	fmt.Fprintf(w, "\tCommands:\t%s\n", strings.Join(set.Defaults.Commands, " "))
	fmt.Fprintf(w, "\tSorting:\n")
	fmt.Fprintf(w, "\t\tBy:\t%s\n", set.Defaults.Sorting.By)
	fmt.Fprintf(w, "\t\tAsc:\t%t\n", set.Defaults.Sorting.Asc)
	fmt.Fprintf(w, "\tPermissions:\n")
	fmt.Fprintf(w, "\t\tAdmin:\t%t\n", set.Defaults.Perm.Admin)
	fmt.Fprintf(w, "\t\tExecute:\t%t\n", set.Defaults.Perm.Execute)
	fmt.Fprintf(w, "\t\tCreate:\t%t\n", set.Defaults.Perm.Create)
	fmt.Fprintf(w, "\t\tRename:\t%t\n", set.Defaults.Perm.Rename)
	fmt.Fprintf(w, "\t\tModify:\t%t\n", set.Defaults.Perm.Modify)
	fmt.Fprintf(w, "\t\tDelete:\t%t\n", set.Defaults.Perm.Delete)
	fmt.Fprintf(w, "\t\tShare:\t%t\n", set.Defaults.Perm.Share)
	fmt.Fprintf(w, "\t\tDownload:\t%t\n", set.Defaults.Perm.Download)
	w.Flush()

	b, err := json.MarshalIndent(auther, "", "  ")
	checkErr(err)
	fmt.Printf("\nAuther configuration (raw):\n\n%s\n\n", string(b))
}
 */