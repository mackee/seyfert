package reqparser

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/schema"
)

//+seyfert
type RootRequest struct {
	HogeID int `schema:"hoge_id"`
	Page   int `schema:"page"`
}

//+seyfert
func (req RootRequest) String() string {
	return "_PATH_"
}

//+seyfert
type RootResponse struct { //hgehoge
	HogeID int `json:"hoge_id"`
}

//+seyfert
type RootHandler func(RootRequest) (RootResponse, error)

//+seyfert
func generateRootHandler(h RootHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		dec := schema.NewDecoder()
		req := RootRequest{}
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
func RegisterRootHandler(h RootHandler) {
	http.HandleFunc("/", generateRootHandler(h))
}
