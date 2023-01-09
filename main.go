package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hkujath/goblog/page"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

var (
	flagSrcFolder   = flag.String("src", "./pages/", "blog folder")
	flagTmplFolder  = flag.String("tmpl", "./templates/", "template folder")
	flagFilesFolder = flag.String("files", "./files/", "path to the file server")
	flagPort        = flag.String("port", ":8001", "port of the web server")
)

func main() {
	flag.Parse()
	// Handler function
	http.HandleFunc("/page/", makePageHandlerFunc())
	http.HandleFunc("/api/", makeAPIHandlerFunc())
	http.HandleFunc("/comments/", makeCommentHandlerFunc())
	http.Handle("/files/",
		http.StripPrefix("/files/", http.FileServer(http.Dir(*flagFilesFolder))))
	http.HandleFunc("/", makeIndexHandlerFunc())

	// start server
	fmt.Printf("Starting server at port %s\n", *flagPort)
	err := http.ListenAndServe(*flagPort, nil)
	if err != nil {
		fmt.Println("ListenAndServe:", err)
	}
}

func makeIndexHandlerFunc() http.HandlerFunc {
	tmpl, err := page.ParseFiles("index.tmpl.html", *flagTmplFolder)
	if err != nil {
		fmt.Printf("makeIndexHandlerFunc: %s\n", err)
		panic("Can`t parse template files.")
	}

	var ps page.Pages
	go func() {
		for {
			ps, err = page.LoadPages(*flagSrcFolder)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("index loaded")
			time.Sleep(30 * time.Second)
		}
	}()

	return func(writer http.ResponseWriter, request *http.Request) {
		err = tmpl.ExecuteTemplate(writer, "base", ps)
		if err != nil {
			fmt.Println("Error during execution of template: ", err)
		}
	}
}

func makePageHandlerFunc() http.HandlerFunc {
	tmpl, err := page.ParseFiles("page.tmpl.html", *flagTmplFolder)
	if err != nil {
		fmt.Printf("makePageHandlerFunc: %s\n", err)
		panic("Can`t parse template files.")
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		//fmt.Fprintf(writer, "url: %s: Index", request.URL.Path)
		f := request.URL.Path[len("/page/"):]
		fPath := filepath.Join(*flagSrcFolder, f)

		p, err := page.LoadPage(fPath)
		if err != nil {
			fmt.Println(err)
		}

		err = tmpl.ExecuteTemplate(writer, "base", p)
		if err != nil {
			fmt.Println("Error during execution of template: ", err)
		}
	}
}

func makeCommentHandlerFunc() http.HandlerFunc {
	var mutex = &sync.Mutex{}
	return func(writer http.ResponseWriter, request *http.Request) {
		title := request.URL.Path[len("/comments/"):]

		name := request.FormValue("name")
		comment := request.FormValue("comment")

		c := page.Comment{Name: name, Content: comment}
		mutex.Lock()
		cs, err := page.LoadComments(title)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		cs = append(cs, c)
		err = page.SaveComments(title, cs)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		mutex.Unlock()
		http.Redirect(writer, request, "/page/"+title, http.StatusFound)
	}
}

func makeAPIHandlerFunc() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ps, err := page.LoadPages(*flagSrcFolder)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		enc := json.NewEncoder(writer)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", " ")
		err = enc.Encode(ps)
		if err != nil {
			fmt.Println("Cannot encode pages to json")
		}
	}
}
