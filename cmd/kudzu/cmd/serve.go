package cmd

import (
	"strings"

	"github.com/kudzu-cms/kudzu/system/system"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:     "serve [flags] <service,service>",
	Aliases: []string{"s"},
	Short:   "run the server (serve is wrapped by the run command)",
	Hidden:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		services := strings.Split(args[0], ",")
		if bind == "" {
			bind = "localhost"
		}
		system.Run(bind, port, httpsport, services, dev, devhttps, docs, docsport)
		return nil
	},
}

func init() {

	serveCmd.Flags().StringVar(&bind, "bind", "localhost", "address for kudzu to bind the HTTP(S) server")
	serveCmd.Flags().IntVar(&httpsport, "https-port", 443, "port for kudzu to bind its HTTPS listener")
	serveCmd.Flags().IntVar(&port, "port", 8080, "port for kudzu to bind its HTTP listener")
	serveCmd.Flags().IntVar(&docsport, "docs-port", 1234, "[dev environment] override the documentation server port")
	serveCmd.Flags().BoolVar(&docs, "docs", false, "[dev environment] run HTTP server to view local HTML documentation")
	serveCmd.Flags().BoolVar(&https, "https", false, "enable automatic TLS/SSL certificate management")
	serveCmd.Flags().BoolVar(&devhttps, "dev-https", false, "[dev environment] enable automatic TLS/SSL certificate management")

	rootCmd.AddCommand(serveCmd)

}
