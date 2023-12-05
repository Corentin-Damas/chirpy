package main

import (
	"fmt"

	"log"
	"net/http"
	"os"

	"github.com/Corentin-Damas/chirpy/database"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits int
	DB             *database.DB
	jwtSecret      string
	polkaKey       string
}

func main() {
	const addr = ":8080"

	godotenv.Load(".env")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
		log.Fatal("POLKA_KEY environment variable is not set")
	}

	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
		DB:             db,
		jwtSecret:      jwtSecret,
		polkaKey:       polkaKey,
	}

	r := chi.NewRouter()
	fgHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	r.Handle("/app", fgHandler)
	r.Handle("/app/*", fgHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Get("/reset", apiCfg.handlerReset)

	apiRouter.Post("/polka/webhooks", apiCfg.handlerWebhook)

	apiRouter.Post("/revoke", apiCfg.handlerRevoke)
	apiRouter.Post("/refresh", apiCfg.handlerRefresh)
	apiRouter.Post("/login", apiCfg.handlerLogin)

	apiRouter.Post("/users", apiCfg.handlerUsersCreate)
	apiRouter.Put("/users", apiCfg.handlerUsersUpdate)

	apiRouter.Delete("/chirps/{chirpID}", apiCfg.handlerChirpsDelete)
	apiRouter.Post("/chirps", apiCfg.handlerChirpsCreate)
	apiRouter.Get("/chirps", apiCfg.handlerChirpsRetrieve)
	apiRouter.Get("/chirps/{chirpID}", apiCfg.handlerChirpsGet)
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
