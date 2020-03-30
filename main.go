package main

import "fmt"
import "flag"
import "net/http"

var FILE string
var CSS string

func init() {
	flag.StringVar(&FILE, "file", "", "Markdown file path")
	flag.StringVar(&CSS, "css", "https://eva7base.com/css/sspai.css", "Markdown css file.Can use network location or local path.")
	flag.Parse()

	if FILE == "" {
		panic("Require markdown file path.")
	}
	fmt.Println(FILE)
}

func main() {

	// temp := ""

	serve()
}

func serve() {
	var handler http.HandlerFunc = func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("Hello"))
	}
	http.ListenAndServe(":8080", handler)
}
