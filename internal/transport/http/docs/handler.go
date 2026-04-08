package docs

import (
	"embed"
	"net/http"
)

//go:embed openapi.json
var openAPISpec embed.FS

type Handler struct {
	spec []byte
}

func NewHandler() *Handler {
	spec, err := openAPISpec.ReadFile("openapi.json")
	if err != nil {
		panic(err)
	}

	return &Handler{spec: spec}
}

func (h *Handler) ServeSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(h.spec)
}

func (h *Handler) ServeUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerUIHTML))
}

func (h *Handler) RedirectToUI(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
}

var swaggerUIHTML = `<!doctype html>
<html>
<head><meta charset="utf-8"><title>Task Service Swagger</title></head>
<body>
<h1>Task Service Swagger</h1>
<p>OpenAPI JSON: <a href="/swagger/openapi.json">/swagger/openapi.json</a></p>
</body>
</html>`
