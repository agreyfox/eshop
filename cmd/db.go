package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	_ "github.com/agreyfox/eshop/content"
	"github.com/agreyfox/eshop/system/admin"
	"github.com/agreyfox/eshop/system/db"
	"github.com/agreyfox/eshop/system/item"
	"github.com/spf13/cobra"
)

var (
	inputdb    string
	inputfile  string
	bucketname string
	outputfile string
	outputdb   string
)

func init() {
	dumpCmd.Flags().StringVar(&outputfile, "out", "output.json", "the dump output file, default is output.json")
	dumpCmd.Flags().StringVar(&inputdb, "in", "", "the dump input bolt db file")
	dumpCmd.Flags().StringVar(&bucketname, "name", "", "the name of bucket in input db")
	restoreCmd.Flags().StringVar(&outputdb, "out", "output.db", "the restore output db file")
	restoreCmd.Flags().StringVar(&inputfile, "in", "", "the restore input json file")
	restoreCmd.Flags().StringVar(&bucketname, "name", "", "the name of bucket in output db")
	deleteCmd.Flags().StringVar(&outputdb, "out", "output.db", "the output db file after delete bucketname")
	deleteCmd.Flags().StringVar(&inputdb, "in", "", "the restore input db file")
	deleteCmd.Flags().StringVar(&bucketname, "name", "", "the name of bucket in output db need to be deleted")
	listCmd.Flags().StringVar(&bucketname, "name", "", "the name of bucket in output db")
	listCmd.Flags().StringVar(&inputdb, "in", "", "the name of bucket ")

	dbCmd.AddCommand(dumpCmd)
	dbCmd.AddCommand(restoreCmd)
	dbCmd.AddCommand(deleteCmd)
	dbCmd.AddCommand(listCmd)
	dbCmd.AddCommand(listContentCmd)

	RegisterCmdlineCommand(dbCmd)
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "db utility",
	Long:  `db dump/restore/delete utility.`,
	Args:  cobra.NoArgs,
}

var dumpCmd = &cobra.Command{
	Use:     "dump [flags]",
	Aliases: []string{"du"},
	Short:   "dump specified bucket content",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("dump bucketname %s from db %s,to file %s\n", bucketname, inputdb, outputfile)
		if len(inputdb) == 0 || bucketname == "" || len(outputfile) == 0 {
			return fmt.Errorf("Must provide enough parameter")
		}
		db.Init(inputdb)
		contents := db.ContentAll(bucketname)
		fmt.Println("Total entires is ", len(contents))
		f, err := os.Create(outputfile)
		if err != nil {
			logger.Panic(err)
		}
		w := bufio.NewWriter(f)
		defer f.Close()
		for _, item := range contents {
			w.Write(item)
			w.WriteByte('\n')
		}
		w.Flush()
		fmt.Println("done!")
		return nil
	},
}
var restoreCmd = &cobra.Command{
	Use:     "restore [flags]",
	Aliases: []string{"re"},
	Short:   "restore specified bucket of content",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("restore bucketname %s from file  %s,to db %s\n", bucketname, inputfile, outputdb)
		if len(outputdb) == 0 || bucketname == "" || len(inputfile) == 0 {
			return fmt.Errorf("Must provide enough parameter")
		}
		db.Init(outputdb)
		//contents := db.ContentAll(bucketname)
		contenttype := item.Types[bucketname]()

		//	fmt.Println("Total entires is ", len(contents))
		f, err := os.Open(inputfile)
		if err != nil {
			logger.Panic(err)
		}
		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)

		defer f.Close()
		for scanner.Scan() {
			fmt.Print("A Line:")
			buf := scanner.Bytes()
			//fmt.Println(string(buf))
			_ = json.Unmarshal(buf, contenttype)

			item, ok := contenttype.(item.Identifiable) //.(item.Item)
			if !ok {
				fmt.Println("Error in data ")
				continue
			}
			fmt.Printf("id:%d==>", item.ItemID())
			DisplayExtractObj(contenttype)
			fmt.Println("========>try to restore to db file ")
			//fmt.Println(err, contenttype)
			admin.RestoreContent(bucketname, item.ItemID(), buf)

		}

		fmt.Printf("\nRestore done!\n")

		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:     "remove [flags] <number>",
	Aliases: []string{"rm"},
	Short:   "delete a specified bucket ",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Remove specified bucket:%s from db :%s\n", bucketname, inputdb)
		var command string
		if len(args) > 0 {
			command = args[0]
		} else {
			command = "test"

		}
		if command == "yes" {
			fmt.Printf("To real delete content in dbb file ,we will save the back file  %s\n", ".bk-"+inputdb)

			if len(inputdb) == 0 || bucketname == "" {

				return fmt.Errorf("Please specified input db file name or bucketname")
			}
			fmt.Printf("To save the db file for backup %s\n", ".bk-"+inputdb)
			FileCopy(inputdb, ".bk-inputdb")

		}

		db.Init(inputdb)
		list, err := admin.GetAllKeyOfContent(bucketname)
		if err != nil {
			fmt.Println("no data and key in the bucket ", bucketname)
		}
		for i, _ := range list {
			if command == "test" {
				fmt.Printf("key %s will be delete\n", list[i])
			} else if command == "yes" {
				err := db.DeleteContent(bucketname + ":" + fmt.Sprint(list[i]))
				if err != nil {
					fmt.Printf("Delete error\n ")
				} else {
					fmt.Printf("\tcontent with key %s deleted!\n", list[i])
				}
			}
		}
		return nil
	},
}

var listContentCmd = &cobra.Command{
	Use: "all",
	//Aliases: []string{"rm"},
	Short: "show the system content/bucket",
	RunE: func(cmd *cobra.Command, args []string) error {
		DistplayAllBuckets()
		return nil
	},
}

var listCmd = &cobra.Command{
	Use: "list [flags] <number>",
	//Aliases: []string{"rm"},
	Short: "show the content/bucket record  number ",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Try to list the content %s with db bucket :%s\n", inputdb, bucketname)
		fmt.Printf("argument id %v\n", args)
		nn := 10
		var err error
		if len(args) > 0 {
			nn, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Printf("please give total list number %d \n", bucketname)
				nn = 10
			}
		}

		db.Init(inputdb)
		keylist, err := admin.GetAllKeyOfContent(bucketname)
		if err != nil {
			fmt.Println("not found the data ")
		}
		var j = 0
		for i := 0; i < len(keylist); i++ {
			ccc, err := db.Content(bucketname + ":" + keylist[i])
			if err == nil {
				fmt.Println(keylist[i], ":", string(ccc))
				j++
				if j > nn {
					break
				}
			}

		}

		return nil
	},
}

// DistplayAllConfig to Display all the system configuration
func DistplayAllBuckets() {
	logger.Info("all content bucket:")
	for key, _ := range item.Types {
		fmt.Println(key)

	}
}

// PrettyPrint map[string]internface output
func DisplayExtractObj(obj interface{}) {
	prettyJSON, err := json.MarshalIndent(obj, "", "")
	if err != nil {
		log.Fatal("Failed to generate json", err)
	}
	fmt.Println("===================================================================")
	fmt.Printf("\t\t%s\n", string(prettyJSON))
	fmt.Println("===================================================================")
}

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func FileCopy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
