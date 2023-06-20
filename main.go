package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"p-cinema-go/api"
	"p-cinema-go/api/admin"
	"p-cinema-go/rdbms"
	"p-cinema-go/service"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	rdbms.InitDb()
	defer rdbms.Close()

	r := mux.NewRouter()
	r.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			body, _ := ioutil.ReadAll(r.Body)
			r.Body = ioutil.NopCloser(strings.NewReader(string(body)))
			fmt.Printf("uri: %s\nrequest: %+v\nbody: %s\n", r.RequestURI, r, string(body))
			// <-time.NewTicker(time.Second * 2).C
			h.ServeHTTP(rw, r)
		})
	})
	r.Methods(http.MethodGet).Path("/movies").HandlerFunc(service.PublicApi(api.GetAvaliableMovies)).Queries("date", "")
	r.Methods(http.MethodGet).Path("/movieDetail/{moviePath}").HandlerFunc(service.PublicApi(api.GetMovieDetail))
	r.Methods(http.MethodGet).Path("/searchMovie").HandlerFunc(service.PublicApi(api.SearchMovie)).Queries("movieId", "", "date", "", "theatreId", "")
	r.Methods(http.MethodGet).Path("/theatreField").HandlerFunc(service.PublicApi(api.GetTheatreField)).Queries("date", "", "theatreId", "")
	r.Methods(http.MethodGet).Path("/theatres").HandlerFunc(service.PublicApi(api.GetTheatres))
	r.Methods(http.MethodGet).Path("/theatreHouses").HandlerFunc(service.PublicApi(api.GetTheatreHouses)).Queries("theatreId", "")
	r.Methods(http.MethodGet).Path("/fieldSettingPlan/{id}").HandlerFunc(service.PublicApi(api.GetFieldSettingPlan))
	r.Methods(http.MethodGet).Path("/attachmentHandler/{filename}").HandlerFunc(api.AttachmentHandler)

	adminR := r.PathPrefix("/admin").Subrouter()
	adminR.Methods(http.MethodPost).Path("/login").HandlerFunc(service.PublicApi(admin.Login))
	adminR.Methods(http.MethodPost).Path("/createOrUpdateTheatre").HandlerFunc(service.PrivateApi(api.CreateOrUpdateTheatre))
	adminR.Methods(http.MethodGet).Path("/theatresDetail/{id}").HandlerFunc(service.PrivateApi(api.GetTheatresDetail))
	adminR.Methods(http.MethodPost).Path("/createOrUpdateHouse").HandlerFunc(service.PrivateApi(api.CreateOrUpdateHouse))
	adminR.Methods(http.MethodGet).Path("/houseDetail/{id}").HandlerFunc(service.PrivateApi(api.HouseDetail))
	adminR.Methods(http.MethodPost).Path("/createOrUpdateMovie").HandlerFunc(service.PrivateApi(api.CreateOrUpdateMovie))
	adminR.Methods(http.MethodGet).Path("/movieDetail/{id}").HandlerFunc(service.PrivateApi(api.GetMovieDetailById))
	adminR.Methods(http.MethodPost).Path("/createOrUpdateField").HandlerFunc(service.PrivateApi(api.CreateOrUpdateField))
	adminR.Methods(http.MethodPost).Path("/uploadAttachment").HandlerFunc(service.PrivateApi(api.UploadAttachment))

	r.PathPrefix("/").HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Println("404::::::: ", r.RequestURI)
		rw.WriteHeader(http.StatusNotFound)
	})
	log.Fatal(http.ListenAndServe(":"+os.Getenv("SERVICE_HOST"),
		handlers.CORS(
			handlers.AllowedHeaders([]string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "X-Requested-With"}),
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
		)(r),
	))
}
