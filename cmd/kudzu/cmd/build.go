package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
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

func init() {

	files, err := ioutil.ReadDir("./plugins")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Println(file.Name())
		p, err := plugin.Open("./plugins/" + file.Name())
		if err != nil {
			fmt.Println("error")
			log.Fatal(err)
		}
		// _, err = p.Lookup("Page")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		_, err = p.Lookup("MarshalEditor")
		if err != nil {
			log.Fatal(err)
		}
		_, err = p.Lookup("Hello")
		if err != nil {
			fmt.Println("error")
			log.Fatal(err)
		}

	}

	rootCmd.AddCommand(buildCmd)
}
