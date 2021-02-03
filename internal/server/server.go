package server

import (
	"couchdb-proxy/internal/couchdb"
	"couchdb-proxy/internal/pg"
	"errors"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
)

const (
	serverPort = ":8080"
	authHeader = "Authorization"
	authPrefix = "Bearer "
)

var proxy *couchdb.CouchDbProxy
var pgPool *pgxpool.Pool

func Run() {
	http.HandleFunc("/", handler)

	log.Printf("starting http server on port %s ...", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, nil))
}

func init() {
	viper.SetEnvPrefix("proxy")

	proxy = couchdb.NewCouchDbProxy()
	pgPool = pg.GetConnectionPool()
}

func handler(writer http.ResponseWriter, request *http.Request) {
	authToken := extractAuthToken(request)
	if authToken == "" {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	database, err := extractDatabase(request)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = proxy.ProxyRequest(pgPool, authToken, database, writer, request)
	if err != nil {
		switch err.(type) {
		case *couchdb.ForbiddenError:
			writer.WriteHeader(http.StatusForbidden)
		default:
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}

func extractAuthToken(request *http.Request) string {
	auth := request.Header.Get(authHeader)
	auth = strings.TrimPrefix(auth, authPrefix)

	return auth
}

func extractDatabase(request *http.Request) (string, error) {
	parts := strings.Split(request.RequestURI, "/")
	if len(parts) <= 1 {
		return "", errors.New("Can't determine database")
	}

	return parts[1], nil
}
