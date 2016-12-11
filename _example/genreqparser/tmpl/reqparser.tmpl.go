package reqparser

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/schema"
)

//+seyfert
type PATH_Request struct {
	//+expand RequestFields
}

//+seyfert
func (req PATH_Request) Path() string {
	return "_ROUTEPATH_"
}

//+seyfert
type PATH_Response struct { //hgehoge
	//+expand ResponseFields
}

//+seyfert
type PATH_Handler func(PATH_Request) (PATH_Response, error)

//+seyfert
func generate_PATH_Handler(h PATH_Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		dec := schema.NewDecoder()
		req := PATH_Request{}
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			enc.Encode(ErrorResponse{err.Error()})
			return
		}
		err = dec.Decode(&req, r.Form)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			enc.Encode(ErrorResponse{err.Error()})
			return
		}
		res, err := h(req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			enc.Encode(ErrorResponse{err.Error()})
			return
		}
		enc.Encode(res)
	}
}

//+seyfert
func Register_PATH_Handler(h PATH_Handler) {
	http.HandleFunc("_ROUTEPATH_", generate_PATH_Handler(h))
}
