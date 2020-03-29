package main

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func buildDmsServer() error {
	// copy all ./content files to internal vendor directory
	src := "content"
	dst := filepath.Join("cmd", "dmss", "vendor", "github.com", "agreyfox", "dmss", "content")
	err := emptyDir(dst)
	if err != nil {
		return err
	}
	err = copyFilesWarnConflicts(src, dst, []string{"doc.go"})
	if err != nil {
		return err
	}

	// copy all ./addons files & dirs to internal vendor directory
	src = "addons"
	dst = filepath.Join("cmd", "dms", "vendor")
	err = copyFilesWarnConflicts(src, dst, nil)
	if err != nil {
		return err
	}

	// execute go build -o ponzu-cms cmd/ponzu/*.go
	cmdPackageName := strings.Join([]string{".", "cmd", "dms"}, "/")
	buildOptions := []string{"build", "-o", buildOutputName(), cmdPackageName}
	return execAndWait(gocmd, buildOptions...)
}

var buildCmd = &cobra.Command{
	Use:   "build [flags]",
	Short: "build will build/compile the project to then be run.",
	Long: `From within your dms project directory, running build will copy and move
the necessary files from your workspace into the vendored directory, and
will build/compile the project to then be run.

By providing the 'gocmd' flag, you can specify which Go command to build the
project, if testing a different release of Go.

Errors will be reported, but successful build commands return nothing.`,
	Example: `$ dms build
(or)
$ dms build --gocmd=go1.8rc1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return buildDmsServer()
	},
}

func init() {
	RegisterCmdlineCommand(buildCmd)
}
