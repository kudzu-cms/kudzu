package cmd

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/spf13/cobra"
)

func buildkudzuServer() error {
	// execute go build -o kudzu-cms cmd/kudzu/*.go
	cmdPackageName := strings.Join([]string{".", "cmd", "kudzu"}, "/")
	buildOptions := []string{"build", "-o", buildOutputName(), cmdPackageName}
	return execAndWait("go", buildOptions...)
}

// BuildCmd builds the project.
var buildCmd = &cobra.Command{
	Use:   "build [flags]",
	Short: "build will build/compile the project to then be run.",
	Long: `From within your kudzu project directory, running build will copy and move
the necessary files from your workspace into the vendored directory, and
will build/compile the project to then be run.

By providing the 'gocmd' flag, you can specify which Go command to build the
project, if testing a different release of Go.

Errors will be reported, but successful build commands return nothing.`,
	Example: `$ kudzu build
(or)
$ kudzu build --gocmd=go1.8rc1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return buildkudzuServer()
	},
}

func buildPlugins() {
	log.Println("[Plugins] Build")
	err := filepath.Walk("./plugins", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			soBuildCmd := exec.Command("go", "build", "-buildmode=plugin", "-o", path+".so", path)
			log.Println("Plugin: " + info.Name())
			log.Println("\tBuilding: " + soBuildCmd.String())
			err := soBuildCmd.Run()
			if err != nil {
				return err
			}
			log.Println("\tLoading: " + path + ".so")
			p, err := plugin.Open(path + ".so")
			if err != nil {
				return err
			}
			log.Println("\tAttaching: " + path + ".so")
			// Call the Attach method. All content types must implement Attachable.
			a, err := p.Lookup("Attach")
			if err != nil {
				return err
			}
			a.(func())()
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
	log.Println("[Plugins] Done")

}

func init() {
	rootCmd.AddCommand(buildCmd)
}
