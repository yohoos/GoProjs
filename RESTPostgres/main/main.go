package main

import (
	"Tutorials/RESTPostgres/app"
)

func main() {
	a := app.App{}
	a.Initialize("yohoos", "magicdust50", "testdb")
	a.Run(":8080")
}
