package reqparser

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/schema"
)

//+seyfert
type PATH_Request struct {
	//+expland RequestFields
}

//+seyfert
func (req PATH_Request) String() string {
	return "_PATH_"
}

//+seyfert
type PATH_Response struct {
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
		err := dec.Decode(req, r.Form)
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
func register_PATH_Handler(h PATH_Handler) {
	http.HandleFunc("/", generate_PATH_Handler(h))
}
