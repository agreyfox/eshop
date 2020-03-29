package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [flags] <project name>",
	Short: "creates a project directory of the name supplied as a parameter",
	Long: `Creates a project directory of the name supplied as a parameter
immediately following the 'new' option in the $GOPATH/src directory. Note:
'new' depends on the program 'git' and possibly a network connection. If
there is no local repository to clone from at the local machine's $GOPATH,
'new' will attempt to clone the 'github.com/agreyfox/dms' package from
over the network.`,
	Example: `$ dms new github.com/nilslice/proj
> New dms project created at $GOPATH/src/github.com/nilslice/proj`,
	RunE: func(cmd *cobra.Command, args []string) error {

		projectName := "dms"
		if len(args) > 0 {
			projectName = args[0]
		} else {
			msg := "Please provide a project name."
			msg += "\nThis will create a directory within your $GOPATH/src."
			return fmt.Errorf("%s", msg)
		}
		dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		fmt.Println("Current working directory is:", dir)
		return newProjectInDir(projectName)
	},
}

// name2path transforns a project name to an absolute path
func name2path(projectName string) (string, error) {
	/* 	gopath, err := getGOPATH()
	   	if err != nil {
	   		return "", err
	   	} */
	//gosrc := filepath.Join(gopath, "src")
	var err error
	path := projectName
	// support current directory
	if path == "." {
		path, err = os.Getwd()
		if err != nil {
			return "", err
		}
	} else {
		fmt.Printf("New project is %s\n", path)
	}

	// make sure path is inside $GOPATH/src
	/* srcrel, err := filepath.Rel(gosrc, path)
	if err != nil {
		return "", err
	}
	if len(srcrel) >= 2 && srcrel[:2] == ".." {
		return "", fmt.Errorf("path '%s' must be inside '%s'", projectName, gosrc)
	}
	if srcrel == "." {
		return "", fmt.Errorf("path '%s' must not be %s", path, filepath.Join("GOPATH", "src"))
	} */

	_, err = os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err == nil {
		err = os.ErrExist
	} else if os.IsNotExist(err) {
		err = nil
	}

	return path, err
}

func newProjectInDir(path string) error {
	path, err := name2path(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// path exists, ask if it should be overwritten
	if os.IsNotExist(err) {
		fmt.Printf("Using '%s' as project directory\n", path)
		fmt.Println("Path exists, overwrite contents? (y/N):")

		answer, err := getAnswer()
		if err != nil {
			return err
		}

		switch answer {
		case "n", "no", "\r\n", "\n", "":
			fmt.Println("")

		case "y", "yes":
			err := os.RemoveAll(path)
			if err != nil {
				return fmt.Errorf("Failed to overwrite %s. \n%s", path, err)
			}

			return createProjectInDir(path)

		default:
			fmt.Println("Input not recognized. No files overwritten. Answer as 'y' or 'n' only.")
		}

		return nil
	}

	return createProjectInDir(path)
}

func createProjectInDir(path string) error {
	gopath, err := getGOPATH()
	if err != nil {
		return err
	}
	repo := ponzuRepo
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	local := filepath.Join(dir, "/dms/") //, filepath.Join(repo...))
	network := "https://" + strings.Join(repo, "/") + ".git"

	fmt.Printf("network repo is %s\n", network)

	if !strings.HasPrefix(path, dir) {
		path = filepath.Join(dir, path)
	}

	// create the directory or overwrite it
	fmt.Printf("==>create dir %s\n", path)
	err = os.MkdirAll(path, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	if dev {
		fmt.Printf("==>create dev  dir %s\n", path)
		if fork != "" {
			local = filepath.Join(gopath, "src", fork)
		}
		fmt.Printf("==>git clone dms-dev %s\n", local)
		err = execAndWait("git", "clone", local, "--branch", "dms-dev", "--single-branch", path)
		if err != nil {
			return err
		}
		fmt.Printf("==>create vendor core package dir %s\n", path)
		err = vendorCorePackages(path)
		if err != nil {
			return err
		}

		fmt.Println("Dev build cloned from " + local + ":dms-dev")
		return nil
	}
	fmt.Printf("==>git clone %s, %s\n", dir, path)
	// try to git clone the repository from the local machine's $GOPATH
	err = execAndWait("git", "clone", dir+"/dms", path)
	if err != nil {
		fmt.Println("Couldn't clone from", dir+"/dms", "- trying again...")
	}
	/* if err != nil {
		fmt.Println("Couldn't clone from", local, "- trying network...")

		// try to git clone the repository over the network
		networkClone := exec.Command("git", "clone", network, path)
		networkClone.Stdout = os.Stdout
		networkClone.Stderr = os.Stderr

		err = networkClone.Start()
		if err != nil {
			fmt.Println("Network clone failed to start. Try again and make sure you have a network connection.")
			return err
		}
		err = networkClone.Wait()
		if err != nil {
			fmt.Println("Network clone failure.")
			// failed
			return fmt.Errorf("Failed to clone files from local machine [%s] and over the network [%s].\n%s", local, network, err)
		}
	} */
	fmt.Printf("==>vendor core package %s\n", path)
	err = vendorCorePackages(path)
	if err != nil {
		return err
	}

	keyPath := filepath.Join(path, "key")
	fmt.Printf("==>create the tls key %s\n", keyPath)
	err = os.MkdirAll(keyPath, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	staticPath := filepath.Join(path)
	err = os.MkdirAll(staticPath, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}
	staticSrc := filepath.Join(dir, "dms", "system", "admin", "static")

	fmt.Printf("==>Copy the static dir  %s to %s \n", staticSrc, staticPath)
	err = copyAll(staticSrc, staticPath)
	if err != nil {
		return err
	}
	fmt.Println("==>Remove unnecessary file ")
	// remove non-project files and directories
	rmPaths := []string{".git", ".circleci"}
	for _, rm := range rmPaths {
		dir := filepath.Join(path, rm)
		err = os.RemoveAll(dir)
		if err != nil {
			fmt.Println("Failed to remove directory from your project path. Consider removing it manually:", dir)
		}
	}

	fmt.Println("New dms project created at", path)
	return nil
}

func init() {
	newCmd.Flags().StringVar(&fork, "fork", "", "modify repo source for dms core development")
	newCmd.Flags().BoolVar(&dev, "dev", false, "modify environment for dms core development")

	RegisterCmdlineCommand(newCmd)
}
