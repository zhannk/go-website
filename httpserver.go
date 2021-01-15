package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

const (
	HTMLDIR   = "html"
	UPLOADDIR = "upload"
)

var templates = make(map[string]*template.Template)
var basedir string

func init() {
	//get the correct working dir
	err := os.Chdir(path.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	////pwd with final pathseparator
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	basedir = pwd + string(os.PathSeparator)
	pwd = pwd + string(os.PathSeparator) + HTMLDIR + string(os.PathSeparator)

	fmt.Println(pwd)
	fis, err := ioutil.ReadDir(HTMLDIR)
	if err != nil {
		panic(err)
	}

	for _, fi := range fis {
		name := fi.Name()
		fp := pwd + name
		templates[name] = template.Must(template.ParseFiles(fp))

		fmt.Println("parsed template :", fp)
	}
}

// renderHTML render page specified by rc
func renderHTML(rc string, params interface{}, w http.ResponseWriter, r *http.Request) {

	if temp, ok := templates[rc]; ok {
		temp.Execute(w, params)
	} else {
		http.Error(w, "404 No resource found", http.StatusNotFound)
	}
}

// uploadHandler handles /upload requests
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	uri := r.RequestURI
	path := r.URL.Path
	method := r.Method
	body := r.Body

	fmt.Println("uri: " + uri + "   path: " + path + "  method: " + method)
	defer body.Close()

	switch method {
	case "GET":
		renderHTML("upload.html", nil, w, r)

	case "POST":
		fi, fh, err := r.FormFile("image")

		if err != nil {
			panic(err)
		}

		defer fi.Close()
		fp := basedir + UPLOADDIR + string(os.PathSeparator) + fh.Filename
		size := fh.Size

		fo, err := os.Create(fp)
		if err != nil {
			panic(err)
		}
		defer fo.Close()

		n, err := io.Copy(fo, fi)
		if err != nil {
			panic(err)
		}

		if n != size {
			http.Error(w, " File size not match when persisting ", http.StatusInternalServerError)
		}
		http.Redirect(w, r, "view?id="+fh.Filename, http.StatusFound)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	uri := r.RequestURI
	path := r.URL.Path
	method := r.Method
	body := r.Body

	fmt.Println("uri: " + uri + "   path: " + path + "  method: " + method)
	defer body.Close()
	fp := basedir + UPLOADDIR + string(os.PathSeparator) + r.FormValue("id")

	http.ServeFile(w, r, fp)

}
func listHandler(w http.ResponseWriter, r *http.Request) {
	uri := r.RequestURI
	path := r.URL.Path
	method := r.Method
	body := r.Body

	fmt.Println("uri: " + uri + "   path: " + path + "  method: " + method)
	defer body.Close()
	ud := basedir + UPLOADDIR + string(os.PathSeparator)

	fis, err := ioutil.ReadDir(ud)
	if err != nil {
		panic(err)
	}
	param := make(map[string]interface{})
	fns := []string{}

	for _, fi := range fis {
		fns = append(fns, fi.Name())
	}

	param["images"] = fns
	renderHTML("list.html", param, w, r)
}

func safeHandler(fn func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			e := recover()
			if e, ok := e.(error); ok {
				fmt.Println(e)
				http.Error(w, e.Error(), http.StatusInternalServerError)
			}
		}()
		fn(w, r)
	}
}

func main() {

	mux := http.DefaultServeMux

	mux.HandleFunc("/upload", safeHandler(uploadHandler))
	mux.HandleFunc("/view", safeHandler(viewHandler))
	mux.HandleFunc("/", safeHandler(listHandler))

	fmt.Println(UPLOADDIR, HTMLDIR)

	http.ListenAndServe(":8080", mux)
}
