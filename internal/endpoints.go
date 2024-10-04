package endpoints

import (
	"net/http"

	"github.com/vizdos-enterprises/pdfify/internal/generate"
)

func Route(mux *http.ServeMux) {
	mux.HandleFunc("/generate", generate.GenerationEndpointHTTP{}.ServeHTTP)
}
