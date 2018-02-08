package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Healthz returns 200
// In the future we should report the status of the database and any other dependency
func Healthz(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("ok\n"))
}
