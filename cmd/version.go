package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Prints the version of eshop your project is using.",
	Long: `Prints the version of eshop your project is using. Must be called from
within a eshop project directory.`,
	Example: `$ eshop version
> eshop v0.8.2
(or)
$ eshop version --cli
> eshop v0.9.2`,
	Run: func(cmd *cobra.Command, args []string) {
		p, err := version(cli)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "eshop v%s\n", p["version"])
	},
}

func version(isCLI bool) (map[string]interface{}, error) {
	kv := make(map[string]interface{})

	info := filepath.Join("cmd", "dms.json")
	/* if isCLI {
		gopath, err := getGOPATH()
		if err != nil {
			return nil, err
		}
		repo := filepath.Join(gopath, "eshop")
		info = filepath.Join(repo, "cmd", "dms.json")
	}
	*/
	b, err := ioutil.ReadFile(info)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &kv)
	if err != nil {
		return nil, err
	}

	return kv, nil
}

func init() {
	versionCmd.Flags().BoolVar(&cli, "cli", false, "specify that information should be returned about the CLI, not project")
	RegisterCmdlineCommand(versionCmd)
}
