package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/flybot/data-steward-go/common/database"
	"github.com/flybot/data-steward-go/common/queue"
	"github.com/flybot/data-steward-go/common/token"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

type Server struct {
	cfg        ServerConfig
	db         database.Maker
	tokenMaker token.Maker
	queue      queue.Maker
	websrv     *http.Server
}

type MessageResponse struct {
	Msg string `json:"msg"`
}

func JsonResponse(w http.ResponseWriter, data interface{}, c int) {
	dj, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	fmt.Fprintf(w, "%s", dj)
}

var srv Server

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func WebRouter() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(httprate.Limit(50, 10*time.Second,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Too many requests"}`))
		}),
	))

	router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	router.Mount("/public", PublicRouter())
	router.Mount("/auth", AuthRouter())

	// Set up static file serving
	/*workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "assets"))
	FileServer(router, "/assets", filesDir)*/

	return router
}
func PublicRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"msg": "Some JSON data"}`))
	})

	//r.Post("/task", AddTask)

	return r
}

func CreateWebServer(port string) *http.Server {
	var websrv *http.Server

	websrv = &http.Server{
		Handler:      WebRouter(),
		Addr:         port,
		WriteTimeout: 15 * time.Second, // timeout for SSE
		ReadTimeout:  15 * time.Second,
	}

	return websrv
}

func main() {
	// load config
	srv.cfg = LoadConfig("app.yaml")
	// connect to database
	srv.db = database.InitPostgresql()
	srv.db.Connect(srv.cfg.DB.Host, srv.cfg.DB.Port, srv.cfg.DB.User, srv.cfg.DB.Password, srv.cfg.DB.Database)
	// initialize token system
	tm, err := token.NewJWTMaker(srv.cfg.Token.Secret)
	if err != nil {
		log.Fatal("Token system initialization failed")
	}
	srv.tokenMaker = tm
	// create a web server
	srv.websrv = CreateWebServer(srv.cfg.WEB.Port)

	go func() {
		if srv.cfg.WEB.UseTLS {
			log.Fatal(srv.websrv.ListenAndServeTLS("../cert/server-cert.pem", "../cert/server-key.pem"))
		} else {
			log.Fatal(srv.websrv.ListenAndServe())
		}
	}()
	log.Println("Web server started")

	// Wait for an interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Attempt a graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//close database connection
	srv.db.Close()
	log.Fatal(srv.websrv.Shutdown(ctx))

}
