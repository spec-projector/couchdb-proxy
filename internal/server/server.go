package server

import (
	"couchdb-proxy/internal/couchdb"
	"couchdb-proxy/internal/pg"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

const serverPort = ":8080"

func init() {
	viper.SetEnvPrefix("proxy")
}

func Run() {
	proxy := couchdb.NewCouchDbProxy()
	pgPool := pg.GetConnectionPool()

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		proxy.ProxyRequest(pgPool, writer, request)
	})

	log.Printf("starting http server on port %s ...", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, nil))
}
