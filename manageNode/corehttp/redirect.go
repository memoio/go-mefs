package corehttp

import (
	"net"
	"net/http"

	core "github.com/memoio/go-mefs/core"
)

func RedirectOption(path string, redirect string) ServeOption {
	handler := &redirectHandler{redirect}
	return func(n *core.MefsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		mux.Handle("/"+path+"/", handler)
		return mux, nil
	}
}

type redirectHandler struct {
	path string
}

func (i *redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, i.path, 302)
}
