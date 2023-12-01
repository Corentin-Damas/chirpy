package main

import (
	"fmt"

	"log"
	"net/http"

	"github.com/go-chi/chi"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	const addr = ":8080"

	r := chi.NewRouter()

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	fgHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	r.Handle("/app", fgHandler)
	r.Handle("/app/*", fgHandler)
	r.Get("/healthz", handlerReadiness)
	r.Get("/metrics", apiCfg.handlerMetrics)
	r.Get("/reset", apiCfg.handlerReset)

	corsMux := middlewareCors(r)

	srv := &http.Server{
		Addr:    addr,
		Handler: corsMux,
	}
	fmt.Println("server started on port ")
	err := srv.ListenAndServe()
	log.Fatal(err)
}
