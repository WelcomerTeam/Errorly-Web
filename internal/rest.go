package errorly

import (
	"net/http"

	"github.com/gorilla/mux"
)

func me(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}

func createEndpoints(router *mux.Router) {
	router.HandleFunc("/", me)

	return
}
