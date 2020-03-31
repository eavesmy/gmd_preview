package main

import (
"github.com/fsnotify/fsnotify"
"io/ioutil"
"bytes"
"fmt"
"io"
"flag"
"text/template"
"github.com/gorilla/websocket"
"net/http"
)

var FILE string
var CSS string
var PORT string
var reloadChan = make(chan bool)
var Conn *websocket.Conn

type Md struct {
    Title string
    Css string
    Markdown string
    Port string
}

var md *Md

func init() {
	flag.StringVar(&PORT, "port", "8080", "Service port")
	flag.StringVar(&FILE, "file", "", "Markdown file path")
	flag.StringVar(&CSS, "css", "https://eva7base.com/css/sspai.css", "Markdown css file.Can use network location or local path.")
	flag.Parse()

	if FILE == "" {
		panic("Require markdown file path.")
	}

    md = &Md{}

    md.Title = FILE
    md.Css = CSS
    md.Markdown = "测试"
    md.Port = PORT
}

func main() {
    
	// temp := ""

    go watcher()

    http.HandleFunc("/",index)
    http.HandleFunc("/ws",ws)

    fmt.Println("Service start at 8080")

	if err := http.ListenAndServe(":" + PORT, nil); err != nil {
        panic(err)
    }
}

// Browser get html page.
func index(res http.ResponseWriter,req *http.Request){
    Html(res)
}

// Browser get websocket.
func ws(res http.ResponseWriter,req *http.Request){
        conn,err := (&websocket.Upgrader{CheckOrigin:func(r *http.Request) bool {return true}}).Upgrade(res,req,nil)

        if err != nil {
            fmt.Println(err)
        }
        Conn = conn
        
        go func(){
        
        for{
            <-reloadChan
            data := []byte{}
            buffer := bytes.NewBuffer(data)
            Html(buffer)
            conn.WriteMessage(websocket.TextMessage,data)
        }
        }()
}

func Html(w io.Writer){
    temp := template.New("html")
    temp.Parse(`
<html>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="{{.Css}}">
    <title>{{.Title}}</title>
    <body>
        <div id="content"></div>
        <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
        <script>
            window.onload = () => {
                let md = "{{.Markdown}}";
                let conn = new WebSocket("ws://localhost:{{.Port}}/ws")
                conn.onmessage = (ret) => {
                    md = ret.data;
                    document.getElementById('content').innerHTML = marked(md);
                };
            }
        </script>
    </body>
</html>
    `)
    temp.Execute(w,md)
}

func watcher(){
    watch,err:=fsnotify.NewWatcher()
    if err != nil {
        panic(err)
    }
    defer watch.Close()

    err = watch.Add(FILE)

    for {
        event := <- watch.Events

        if event.Op&fsnotify.Write == fsnotify.Write {
            buffer, err := ioutil.ReadFile(FILE)
            if err != nil {
                panic(err)
            }
            md.Markdown = string(buffer)
            reloadChan <- true
        }
    }
}
