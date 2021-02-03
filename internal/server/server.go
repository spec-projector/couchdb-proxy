package server

import (
	"couchdb-proxy/internal/couchdb"
	"couchdb-proxy/internal/pg"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

func init() {
	viper.SetEnvPrefix("proxy")
}

func Run() {
	proxy := couchdb.NewCouchDbProxy()
	pgPool := pg.GetConnectionPool()

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		proxy.ProxyRequest(pgPool, writer, request)
	})

	log.Printf("starting http server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
