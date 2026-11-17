package main

import (
  "fmt"
  "log"
  "net/http"
  "sync/atomic"
)

type apiConfig struct {
  fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    cfg.fileserverHits.Add(1)
    next.ServeHTTP(w, r)
  })
}

func main() {
  mux := http.NewServeMux()
  apiCfg := apiConfig{}

  mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

  mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(http.StatusText(http.StatusOK)))
  })

  mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, req *http.Request) {
    w.Header().Add("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(fmt.Sprintf(`
      <html>
        <body>
          <h1>Welcome, Chirpy Admin</h1>
          <p>Chirpy has been visited %d times!</p>
        </body>
      </html>`,
      apiCfg.fileserverHits.Load())))
  })

  mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, req *http.Request) {
    w.Header().Add("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    apiCfg.fileserverHits.Store(0)
  })

  s := http.Server{
    Handler: mux,
    Addr: ":8080",
  }
  log.Fatal(s.ListenAndServe())
}
