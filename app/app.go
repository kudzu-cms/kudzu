package app

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/kudzu-cms/kudzu/system/admin"
	"github.com/kudzu-cms/kudzu/system/api"
	"github.com/kudzu-cms/kudzu/system/api/analytics"
	"github.com/kudzu-cms/kudzu/system/db"
	"github.com/kudzu-cms/kudzu/system/tls"
)

// ErrWrongOrMissingService informs a user that the services to run must be
// explicitly specified when serve is called
var ErrWrongOrMissingService = errors.New("To execute 'kudzu serve', " +
	"you must specify which service to run.")

// Run starts the project.
func Run(bind string, port int, https bool, httpsport int, services []string, dev bool, devhttps bool, docs bool, docsport int) error {

	pluginsPath := filepath.Join(".", "plugins")
	info, err := os.Stat(pluginsPath)
	if !os.IsNotExist(err) && info.IsDir() {
		buildPlugins()
	}

	db.Init()
	defer db.Close()

	analytics.Init()
	defer analytics.Close()

	if len(services) == 0 {
		return ErrWrongOrMissingService
	}

	for _, service := range services {
		if service == "api" {
			api.Run()
		} else if service == "admin" {
			admin.Run()
		} else {
			return ErrWrongOrMissingService
		}
	}

	// run docs server if --docs is true
	if docs {
		admin.Docs(docsport)
	}

	// init search index
	go db.InitSearchIndex()

	// save the https port the system is listening on
	err = db.PutConfig("https_port", fmt.Sprintf("%d", httpsport))
	if err != nil {
		log.Fatalln("System failed to save config. Please try to run again.", err)
	}

	// cannot run production HTTPS and development HTTPS together
	if devhttps {
		fmt.Println("Enabling self-signed HTTPS... [DEV]")

		go tls.EnableDev()
		fmt.Println("Server listening on https://localhost:10443 for requests... [DEV]")
		fmt.Println("----")
		fmt.Println("If your browser rejects HTTPS requests, try allowing insecure connections on localhost.")
		fmt.Println("on Chrome, visit chrome://flags/#allow-insecure-localhost")

	} else if https {
		fmt.Println("Enabling HTTPS...")

		go tls.Enable()
		fmt.Printf("Server listening on :%s for HTTPS requests...\n", db.ConfigCache("https_port").(string))
	}

	// save the https port the system is listening on so internal system can make
	// HTTP api calls while in dev or production w/o adding more cli flags
	err = db.PutConfig("http_port", fmt.Sprintf("%d", port))
	if err != nil {
		log.Fatalln("System failed to save config. Please try to run again.", err)
	}

	// save the bound address the system is listening on so internal system can make
	// HTTP api calls while in dev or production w/o adding more cli flags
	err = db.PutConfig("bind_addr", bind)
	if err != nil {
		log.Fatalln("System failed to save config. Please try to run again.", err)
	}

	fmt.Printf("Server listening at http://%s:%d for HTTP requests...\n", bind, port)
	fmt.Printf("\nVisit http://%s:%d/admin to get started.\n", bind, port)
	return http.ListenAndServe(fmt.Sprintf("%s:%d", bind, port), nil)
}

func buildPlugins() {
	log.Println("[Plugins] Build")
	err := filepath.Walk(filepath.Join(".", ".plugins"), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".so") {
			log.Println("\tLoading: " + path)
			p, err := plugin.Open(path)
			if err != nil {
				return err
			}
			log.Println("\tAttaching: " + path)
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
