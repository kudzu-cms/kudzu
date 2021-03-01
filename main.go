package main

import "github.com/kudzu-cms/kudzu/app"

func main() {
	services := [2]string{"admin", "api"}
	app.Run("localhost", 8080, false, 8043, services[0:2], false, false, false, 8081)
}
