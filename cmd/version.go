package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/agreyfox/eshop/system/admin/user"
	"github.com/agreyfox/eshop/system/api/analytics"
	"github.com/agreyfox/eshop/system/db"

	smtp2go "github.com/agreyfox/eshop/system/email"
	"github.com/agreyfox/eshop/system/ip"
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

var emailCmd = &cobra.Command{
	Use:     "email",
	Aliases: []string{"email"},
	Short:   "Try to connect admin email to send a test email .",
	Long:    `Testing email send via admin email configuration.`,
	Example: `$ eshop email`,
	Run: func(cmd *cobra.Command, args []string) {
		db.Init(systemdb)
		defer db.Close()
		//db.PutConfig("Key", config.GenerateKey())

		//analytics.Init()
		//defer analytics.Close()

		tomail := []string{"18901882538@189.cn"}
		if len(args) == 1 {
			tomail = append(tomail, args[0])
		}
		fmt.Printf("Try to send email to %v\n", tomail)
		emailcontent := smtp2go.Email{
			//From: admin.MailUser,
			To:       tomail,
			Subject:  "Trying out EShop email service",
			TextBody: "Eshop Test Message",
			HtmlBody: "<h1>Eshop Test Message</h1>",
		}
		res, err := smtp2go.Send(&emailcontent)
		if err != nil {
			fmt.Printf("An Error Occurred: %s\n", err)
		}
		if res.Data.Succeeded == 1 {
			fmt.Printf("Sent Successfully: %v\n", res)
		} else {
			fmt.Printf("Sent with error: %v\n", res)
		}

		fmt.Fprintf(os.Stdout, "Email testing is done, please check email\n")
	},
}

var ipCmd = &cobra.Command{
	Use:     "ip",
	Aliases: []string{"ip"},
	Short:   "Try to search ip for country .",
	Long:    `Found the right country for specified IP .`,
	Example: `$ eshop ip 193.168.3.3`,
	Run: func(cmd *cobra.Command, args []string) {
		db.Init(systemdb)
		defer db.Close()
		//db.PutConfig("Key", config.GenerateKey())

		analytics.Init()
		defer analytics.Close()
		ip.Init()

		searchip := ""
		if len(args) == 1 {
			searchip = args[0]
		}
		fmt.Printf("Try to search ip :%s\n", searchip)
		email := ip.NewClient("", true)

		res, err := email.LookupStandard(searchip)
		if err != nil {
			fmt.Printf("An Error Occurred: %s\n", err)
		} else {
			fmt.Printf("The ip belone to  %s \n", res)
		}
		mm, err := email.QueryIPByDB(searchip)
		fmt.Printf("Query local db , return %s\n", mm)

	},
}

var createUserCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{"u"},
	Short:   "Create system admin usr from cli.",
	Long:    `Assign new addmin to system.`,
	Example: `$ eshop user  add jihua.gao@gmail.com axxdsdawe`,
	Run: func(cmd *cobra.Command, args []string) {
		db.Init(systemdb)
		defer db.Close()
		//db.PutConfig("Key", config.GenerateKey())

		if len(args) < 3 {
			fmt.Println("use user add/rm email@address.com xxx")
			return
		}
		cmddd := args[0]
		fmt.Println(args)
		switch strings.ToLower(cmddd) {
		case "add":
			ur, err := user.New(args[1], args[2])
			if err != nil {
				fmt.Println("User operation error")
				return
			}
			ur.Perm.Admin = true
			ur.IsAdmin = true // to keep it is admin user
			_, err = db.SetUser(ur)
			if err == nil {
				fmt.Println("user created !")
			} else {
				fmt.Println("User created error !", err)
			}
		case "rm":
		case "chpassed":
		default:
			fmt.Println("no cmd present")

		}

	},
}

func init() {
	versionCmd.Flags().BoolVar(&cli, "cli", false, "specify that information should be returned about the CLI, not project")
	emailCmd.Flags().BoolVar(&email, "email", false, "start to test email connection")
	ipCmd.Flags().StringVar(&ipip, "ip", "127.0.0.1", "start to test ip lookup service connection")
	createUserCmd.Flags().StringVar(&usercmd, "add", "admin@127.0.0.1", "create/remove system admin user")
	RegisterCmdlineCommand(versionCmd)
	RegisterCmdlineCommand(emailCmd)
	RegisterCmdlineCommand(ipCmd)
	RegisterCmdlineCommand(createUserCmd)
}
