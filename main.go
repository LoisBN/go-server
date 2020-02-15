package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	uuid "github.com/satori/go.uuid"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseFiles("./index.html"))
}

func main() {
	err := os.Chdir("server")
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", fileServer)
	http.HandleFunc("/send", upload)
	http.HandleFunc("/download/", download)
	http.HandleFunc("/favicon.ico", home)
	http.ListenAndServe(":7150", nil)
}

func home(w http.ResponseWriter, req *http.Request) {
	f, err := os.Open("./image.jpg")
	if err != nil {
		http.Error(w, "file not found", 404)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}

func serveHome(w http.ResponseWriter, req *http.Request) {
	f, err := os.Open("./image.jpg")
	if err != nil {
		http.Error(w, "file not found", 404)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		http.Error(w, "file not found", 404)
		return
	}
	http.ServeContent(w, req, f.Name(), fi.ModTime(), f)
}

func serveFileHome(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "./image.jpg")
}

func fileServer(w http.ResponseWriter, req *http.Request) {
	c, err := req.Cookie("connexion")
	if err != nil {
		c = &http.Cookie{
			Name:  "connexion",
			Value: "1",
		}
		fmt.Println(err.Error())
	}
	cv, err := strconv.Atoi(c.Value)
	cv++
	http.SetCookie(w, &http.Cookie{
		Name:  "connexion",
		Value: strconv.Itoa(cv),
	})

	c, err = req.Cookie("session")
	if err != nil {
		id, err := uuid.NewV4()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cookie := &http.Cookie{
			Name:     "session",
			Value:    id.String(),
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
	}
	s, err := filepath.Glob("./*")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
	}
	tpl.ExecuteTemplate(w, "index.html", s)
	//id,_ := uuid.NewV4()
	//fmt.Println(id)

}

func upload(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		f, h, err := req.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("code goes here")
		defer f.Close()
		n, err := os.Create(h.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer n.Close()
		bs, err := ioutil.ReadAll(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s := string(bs)
		n.WriteString(s)
	}
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusSeeOther)
}

func download(w http.ResponseWriter, req *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:  "download",
		Value: http.LocalAddrContextKey.String(),
	})
	fmt.Println(string(req.URL.Path[10:]))
	f, err := os.Open(string(req.URL.Path[10:]))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s := string(bs)
	fmt.Println("works")
	io.WriteString(w, s)
}
