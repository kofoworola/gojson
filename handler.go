package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kofoworola/gojson/gen"
	"github.com/kofoworola/gojson/logging"
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
	logger   *logging.Logger
}

type homePageData struct {
	GOData       string
	JSONData     string
	ErrorMessage string
}

func NewHandler(logger *logging.Logger) (*Handler, error) {
	t, err := template.New("index").Parse(indexHTML)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	h := &Handler{
		template: t,
		logger:   logger,
	}

	r := mux.NewRouter()
	r.HandleFunc("/", h.HomePage).Methods("GET")
	r.HandleFunc("/", h.PostHomePage).Methods("POST")

	h.mux = r
	return h, nil
}

func (h *Handler) HomePage(w http.ResponseWriter, req *http.Request) {

	h.template.Execute(w, homePageData{
		GOData: sampleData,
	})
}

func (h *Handler) PostHomePage(w http.ResponseWriter, req *http.Request) {
	logger := logging.FromContext(req.Context())

	if err := req.ParseForm(); err != nil {
		logger.Errorf("error parsing input: %v", err)
		h.respondError(w, "error reading input", sampleData, http.StatusBadRequest)
		return
	}

	godata := req.PostForm.Get("godata")
	if godata == "" {
		h.respondError(w, "go input can not be null", sampleData, http.StatusBadRequest)
		return
	}
	logger = logger.WithField("godata", godata)

	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("panic: %v", err)
			h.respondError(w, "internal server error", godata, http.StatusInternalServerError)
		}
	}()

	wrapper, err := gen.NewFromString(godata)
	if err != nil {
		logger.Errorf("error creating go wrapper: %v", err)
		h.respondError(w, fmt.Sprintf("syntax error: %v", err), godata, http.StatusBadRequest)
		return
	}

	resp, err := wrapper.GenerateJSONAst()
	if err != nil {
		logger.Errorf("error generating json: %v", err)
		h.respondError(w, fmt.Sprintf("error parsing fields: %v", err), godata, http.StatusBadRequest)
		return

	}

	respStrings := make([]string, len(resp))
	for i, res := range resp {
		respStrings[i] = fmt.Sprintf("//%s\n%s", res.Key, res.String(0))
	}

	dat := homePageData{
		GOData:   godata,
		JSONData: strings.Join(respStrings, "\n"),
	}

	h.template.Execute(w, dat)

}

func (h *Handler) respondError(w http.ResponseWriter, message, godata string, code int) {
	dat := homePageData{
		GOData:       godata,
		ErrorMessage: message,
	}
	w.WriteHeader(code)
	h.template.Execute(w, dat)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.WithField("path", r.URL.Path)
	req := r.WithContext(logging.ToContext(r.Context(), logger))
	path := strings.TrimPrefix(r.URL.Path, "/")
	if strings.HasPrefix(path, "assets") {
		file, err := fs.ReadFile(path)
		if err != nil {
			http.NotFound(w, req)
			return
		}
		fmt.Fprint(w, string(file))
		return
	} else {
		h.mux.ServeHTTP(w, req)
	}
}
