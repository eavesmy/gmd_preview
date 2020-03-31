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

type Md struct {
    Title string
    Css string
    Markdown string
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
    md.Markdown = ""
}

func main() {
    
	// temp := ""

    go watcher()

	var handler http.HandlerFunc = func(res http.ResponseWriter, req *http.Request) {

        ws := websocket.Upgrader{}
        conn,err := ws.Upgrade(res,req,nil)

        if err != nil {
            fmt.Println(err)
        }

        
        for{
            <-reloadChan
            data := []byte{}
            buffer := bytes.NewBuffer(data)
            Html(buffer)
            conn.WriteMessage(0,data)
        }
	}
    fmt.Println("Service start at 8080")
	if err := http.ListenAndServe(":" + PORT, handler); err != nil {
        panic(err)
    }
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
            document.getElementById('content').innerHTML =
            marked("{{.Markdown}}");
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
