package server

import (
	"couchdb-proxy/internal/couchdb"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

func init() {
	viper.SetEnvPrefix("proxy")
}

func Run() {
	proxy := couchdb.NewCouchDbProxy()

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		proxy.ProxyRequest(writer, request)
	})

	log.Printf("starting http server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
