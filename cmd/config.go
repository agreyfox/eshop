package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/agreyfox/eshop/system/db"
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(initCmd)
	configCmd.AddCommand(allCmd)
	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(getCmd)
	RegisterCmdlineCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management utility",
	Long:  `Configuration management utility.`,
	Args:  cobra.NoArgs,
}

var initCmd = &cobra.Command{
	Use: "init",
	//Aliases: []string{"c"},
	Short: "make default system configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return InitilizedDb()
	},
}
var allCmd = &cobra.Command{
	Use: "all ",
	//Aliases: []string{"c"},
	Short: "List all the configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		DistplayAllConfig()
		return nil
	},
}
var setCmd = &cobra.Command{
	Use: "set <key> <value> ...",
	//Aliases: []string{"c"},
	Short: "save key-value pair to configuration db",
	RunE: func(cmd *cobra.Command, args []string) error {
		SetConfig(args[0], args[1])
		return nil
	},
}
var getCmd = &cobra.Command{
	Use: "get <key> ...",
	//Aliases: []string{"c"},
	Short: "get value from configuration db",
	RunE: func(cmd *cobra.Command, args []string) error {
		GetConfig(args[0])
		return nil
	},
}

// manully initial db
func InitilizedDb() error {

	db.Init(systemdb)
	defer db.Close()

	fmt.Println("System DB initialized!")
	return nil
}

// DistplayAllConfig to Display all the system configuration
func DistplayAllConfig() {
	logger.Info("Trying to get the all configuration in system db")
	db.Init(systemdb)
	defer db.Close()

	config, err := db.ConfigAll()
	if err != nil {
		logger.Fatal("DB open error,", err)
	}
	//s, e := db.GetParameterFromConfig("PaymentSetting", "name", "company_name", "valueString")
	//fmt.Println(s)
	//fmt.Println(e)
	var ma map[string]interface{}
	err = json.Unmarshal(config, &ma)
	PrettyPrint(ma)
}

// SetConfig set config one pair
func SetConfig(key string, value interface{}) {
	fmt.Printf("Try to save the config %s--%v\n ", key, value)
	db.Init(systemdb)
	defer db.Close()

	err := db.PutConfig(key, value)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}
	fmt.Printf("Config %s with value %s saved! \n", key, value)
}

// GetConfig to get one key's value
func GetConfig(key string) {
	fmt.Printf("Try to get the config %s\n", key)
	db.Init(systemdb)
	defer db.Close()

	result, err := db.Config(key)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}
	//value := string(result[:])
	fmt.Println("===================================================================")
	fmt.Printf("\t\tConfig [%s] value is [%v ]\n", key, result)
	fmt.Println("===================================================================")
}

// PrettyPrint map[string]internface output
func PrettyPrint(obj map[string]interface{}) {
	prettyJSON, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		log.Fatal("Failed to generate json", err)
	}
	fmt.Println("===================================================================")
	fmt.Printf("\t\t%s\n", string(prettyJSON))
	fmt.Println("===================================================================")
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
