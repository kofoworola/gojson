package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

//go:embed templates/index.html
var indexHTML string

//go:embed assets/*
var fs embed.FS

type Handler struct {
	template *template.Template
	mux      *mux.Router
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

	h.mux = r
	return h, nil
}

func (s *Handler) HomePage(w http.ResponseWriter, req *http.Request) {
	s.template.Execute(w, nil)
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
