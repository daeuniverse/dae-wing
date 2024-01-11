package ws

import (
	"net/http"
)

type Handler struct{}

var _ http.Handler = &Handler{}

func (*Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if w.Header().Get("x-auth-result") == "0" {
		w.WriteHeader(401)
	}
	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
}
