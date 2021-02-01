package couchdb

import (
	"github.com/spf13/viper"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const (
	paramCouchDbUrl   = "couchdb_url"
	paramCouchDbUser  = "couchdb_user"
	paramCouchDbRoles = "couchdb_roles"
)

var couchDbParams = []string{paramCouchDbUrl, paramCouchDbUser, paramCouchDbRoles}

type couchDbConfig struct {
	Url   *url.URL
	User  string
	Roles string
}

type CouchDbProxy struct {
	config *couchDbConfig
	proxy  *httputil.ReverseProxy
}

func (proxy *CouchDbProxy) ProxyRequest(writer http.ResponseWriter, request *http.Request) {
	auth := request.Header.Get(authorizationHeader)
	if auth == "" {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	parts := strings.Split(request.RequestURI, "/")
	if len(parts) <= 1 {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	allowed, err := isAccessAllowed(parts[1], auth)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if !allowed {
		writer.WriteHeader(http.StatusForbidden)
		return
	}
	request.Header["X-Auth-CouchDB-Roles"] = []string{proxy.config.Roles}
	request.Header["X-Auth-CouchDB-UserName"] = []string{proxy.config.User}

	proxy.proxy.ServeHTTP(writer, request)
}

func NewCouchDbProxy() *CouchDbProxy {
	config := readCouchDbConfig()

	return &CouchDbProxy{
		config: config,
		proxy:  httputil.NewSingleHostReverseProxy(config.Url),
	}
}

func readCouchDbConfig() *couchDbConfig {
	for _, env := range couchDbParams {
		err := viper.BindEnv(env)
		if err != nil {
			panic(err)
		}
	}

	viper.SetDefault(paramCouchDbUrl, "http://couchdb:5984")
	viper.SetDefault(paramCouchDbUser, "admin")
	viper.SetDefault(paramCouchDbRoles, "_admin")

	couchdbUrl, err := url.Parse(viper.GetString(paramCouchDbUrl))
	if err != nil {
		panic(err)
	}

	return &couchDbConfig{
		Url:   couchdbUrl,
		User:  viper.GetString(paramCouchDbUser),
		Roles: viper.GetString(paramCouchDbRoles),
	}
}
