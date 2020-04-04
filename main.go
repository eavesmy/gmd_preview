package main

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
)

var reloadChan = make(chan bool)
var Conn *websocket.Conn

type Md struct {
	Title    string
	Css      string
	Markdown []byte
	Port     string
	Text     string
}

func (m *Md) toString() string {
	return string(m.Markdown)
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
	var err error
	md.Markdown, err = ioutil.ReadFile(md.Title)
	if err != nil {
		panic(err)
	}
}

func main() {

	go watcher()
	go write()

	http.HandleFunc("/", index)
	http.HandleFunc("/ws", ws)

	fmt.Println("Service start at",md.Port)

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
    md.Text = md.toString()
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
            let md = {{.Text}};
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

	watch, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watch.Events:
				fmt.Println("event", event)
				reloadChan <- true
			case err := <-watch.Errors:
				fmt.Println("error", err)
			}
		}
	}()

	if err = watch.Add(md.Title); err != nil {
		fmt.Println("watch file err:", err)
	}
	<-done
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
