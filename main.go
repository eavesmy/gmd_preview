package main

import (
	"flag"
	"io"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"net/http"
	"os"
	"time"
)

var reloadChan = make(chan bool)
var Conn *websocket.Conn
var File *os.File

type Md struct {
	Title    string
	Css      string
	Markdown []byte
	Port     string
}

var md *Md

// init params.
func init() {

	md = &Md{}

	flag.StringVar(&md.Port, "port", "8080", "Service port")
	flag.StringVar(&md.Title, "file", "", "Markdown file path")
	flag.StringVar(&md.Css, "css", "https://eva7base.com/css/sspai.css", "Markdown css file.Can use network location or local path.")
	flag.Parse()

	if md.Title == "" {
		panic("Require markdown file path.")
	}
	File,_ = os.Open(md.Title)
}

func main() {

	go watcher()
	go write()

	http.HandleFunc("/", index)
	http.HandleFunc("/ws", ws)

	fmt.Println("Service start at 8080")

	if err := http.ListenAndServe(":"+md.Port, nil); err != nil {
		panic(err)
	}
}

// Browser get html page.
func index(res http.ResponseWriter, req *http.Request) {
	Html(res)
}

// Browser get websocket.
func ws(res http.ResponseWriter, req *http.Request) {
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)

	if err != nil {
		fmt.Println(err)
	}
	Conn = conn
}

// Gen html data.
func Html(w io.Writer) {
	temp := template.New("html")
	temp.Parse(`
<html>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="{{.Css}}">
    <title>{{.Title}}</title>
    <style>
        #content {
            width: 80vw;
            margin: 0 auto;
        }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
    <body>
        <div id="content"></div>
        <script>
            let md = {{.Markdown}};
            window.onload = () => {
                render();
                let conn = new WebSocket("ws://localhost:8080/ws")
                conn.onmessage = (ret) => {
                    console.log(ret)
                    md = ret.data;
                    render();
                };
            }
            function render(){
                document.getElementById('content').innerHTML = marked(md);
            }
        </script>
    </body>
</html>
    `)
	temp.Execute(w, md)
}

// use long poll for platform compatibility and stay code sample
func watcher() {
	readFile()
}

func readFile(){
	b := []byte{}
	File.Read(b)
	if len(md.Markdown) != len(b) {
		md.Markdown = b
		reloadChan <- true
	}
	time.AfterFunc(5, readFile)
}

// Send md data to client.
func write() {
	for {
		<-reloadChan

		if Conn == nil {
			continue
		}

		Conn.WriteMessage(websocket.TextMessage, []byte(md.Markdown))
	}
}
