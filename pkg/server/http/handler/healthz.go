package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Healthz returns 200
// In the future we should report the status of the database and any other dependency
func Healthz(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// swagger:route GET /healthz service health
	//
	// A simple, unauthenticated endpoint to check if the service is running.
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http
	//
	//     Responses:
	//       default: genericError
	//       200: ok

	w.Write([]byte("ok\n"))
}
