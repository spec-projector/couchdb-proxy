package couchdb

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

type ForbiddenError struct{}

func (err *ForbiddenError) Error() string { return "Forbidden" }

func (proxy *CouchDbProxy) ProxyRequest(pool *pgxpool.Pool, authToken string, database string, writer http.ResponseWriter, request *http.Request) (err error) {
	allowed, err := isAccessAllowed(pool, database, authToken)
	if err != nil {
		return
	}

	if !allowed {
		return &ForbiddenError{}
	}

	request.Header["X-Auth-CouchDB-Roles"] = []string{proxy.config.Roles}
	request.Header["X-Auth-CouchDB-UserName"] = []string{proxy.config.User}

	proxy.proxy.ServeHTTP(writer, request)

	return
}

func NewCouchDbProxy() *CouchDbProxy {
	config := readCouchDbConfig()

	log.Printf("create couchdb proxy: url=%s, user=%s", config.Url, config.User)

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
