package main

import (
	"fmt"

	"log"
	"net/http"

	"github.com/Corentin-Damas/chirpy/database"

	"github.com/go-chi/chi"
)

type apiConfig struct {
	fileserverHits int
	DB             *database.DB
}

func main() {
	const addr = ":8080"

	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	apiCfg := apiConfig{
		fileserverHits: 0,
		DB:             db,
	}

	fgHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	r.Handle("/app", fgHandler)
	r.Handle("/app/*", fgHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Get("/reset", apiCfg.handlerReset)
	apiRouter.Get("/chirps", apiCfg.handleGetChrips)
	apiRouter.Post("/chirps", apiCfg.handlerChirpsCreate)
	r.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.handlerMetrics)
	r.Mount("/admin", adminRouter)

	corsMux := middlewareCors(r)

	srv := &http.Server{
		Addr:    addr,
		Handler: corsMux,
	}
	fmt.Printf("server started on port: %s ...", addr)
	log.Fatal(srv.ListenAndServe())
}
