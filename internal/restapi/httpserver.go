package restapi

import(

	"net/http"

	"golang.org/x/net/html"
	log "github.com/sirupsen/logrus"

)



type DeduplicatorAPI struct {

	httpServer *http.Server

}


func NewDeduplicatorAPI()(*DeduplicatorAPI, error){

	return &DeduplicatorAPI{
		httpServer: nil,
	}, nil 

}

// Start the API 
func (api *DeduplicatorAPI) Start() {

	address := "localhost:8080"

	api.httpServer = &http.Server{Addr: address}

	// Add default listener 
	http.HandleFunc("/", api.default_handler)

	log.Fatal(api.httpServer.ListenAndServe())

}

func (api *DeduplicatorAPI) default_handler(w http.ResponseWriter, r *http.Request){

	log.Infof("Hello, %q", html.EscapeString(r.URL.Path))

}
