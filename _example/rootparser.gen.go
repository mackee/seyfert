package reqparser

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/schema"
)

//+seyfert
type rootRequest struct {
	//+expland RequestFields
}

//+seyfert
func (req rootRequest) String() string {
	return "_PATH_"
}

//+seyfert
type rootResponse struct {
	//+expand ResponseFields
}

//+seyfert
type rootHandler func(rootRequest) (rootResponse, error)

//+seyfert
func generateRootHandler(h rootHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		dec := schema.NewDecoder()
		req := rootRequest{}
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
func registerRootHandler(h rootHandler) {
	http.HandleFunc("/", generateRootHandler(h))
}
