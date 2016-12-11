package main

import (
	"net/http"

	rp "github.com/mackee/seyfert/_example/reqparser"
)

func main() {
	rp.RegisterRootHandler(func(req rp.RootRequest) (rp.RootResponse, error) {
		resp := rp.RootResponse{
			HogeID: req.HogeID,
		}
		return resp, nil
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
