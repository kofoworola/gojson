package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kofoworola/gojson/gen"
	"github.com/sirupsen/logrus"
)

//go:embed templates/index.html
var indexHTML string

//go:embed assets/*
var fs embed.FS

//go:embed sample.go
var sampleData string

type Handler struct {
	template *template.Template
	mux      *mux.Router
}

type homePageData struct {
	GOData   string
	JSONData string
}

func NewHandler() (*Handler, error) {
	t, err := template.New("index").Parse(indexHTML)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	h := &Handler{
		template: t,
	}

	r := mux.NewRouter()
	r.HandleFunc("/", h.HomePage).Methods("GET")
	r.HandleFunc("/", h.PostHomePage).Methods("POST")

	h.mux = r
	return h, nil
}

func (s *Handler) HomePage(w http.ResponseWriter, req *http.Request) {

	s.template.Execute(w, homePageData{
		GOData: sampleData,
	})
}

func (s *Handler) PostHomePage(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		logrus.Errorf("error parsing input: %v", err)
		http.Error(w, "error reading input", http.StatusBadRequest)
		return
	}

	godata := req.PostForm.Get("godata")
	if godata == "" {
		http.Error(w, "go input can not be null", http.StatusBadRequest)
		return
	}

	wrapper, err := gen.NewFromString(godata)
	if err != nil {
		logrus.Errorf("error creating go wrapper: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := wrapper.GenerateJSONAst()
	if err != nil {
		logrus.Errorf("error generating json: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return

	}

	respStrings := make([]string, len(resp))
	for i, res := range resp {
		respStrings[i] = res.String(0)
	}

	dat := homePageData{
		GOData:   godata,
		JSONData: strings.Join(respStrings, "\n"),
	}

	s.template.Execute(w, dat)

}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if strings.HasPrefix(path, "assets") {
		file, err := fs.ReadFile(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, string(file))
		return
	} else {
		h.mux.ServeHTTP(w, r)
	}
}
