package rapidus

import (
	"github.com/fouched/toolkit/v2"
	"net/http"
)

var t toolkit.Tools

func (r *Rapidus) ReadJSON(w http.ResponseWriter, req *http.Request, data interface{}) error {
	return t.ReadJSON(w, req, data)
}

func (r *Rapidus) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	return t.WriteJSON(w, status, data, headers...)
}

func (r *Rapidus) WriteXML(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	return t.WriteXML(w, status, data, headers...)
}

func (r *Rapidus) DownloadStaticFile(w http.ResponseWriter, req *http.Request, pathName, displayName string) {
	t.DownloadStaticFile(w, req, pathName, displayName)
}

func (r *Rapidus) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (r *Rapidus) Error404(w http.ResponseWriter) {
	r.ErrorStatus(w, http.StatusNotFound)
}

func (r *Rapidus) Error500(w http.ResponseWriter) {
	r.ErrorStatus(w, http.StatusInternalServerError)
}

func (r *Rapidus) ErrorUnauthorized(w http.ResponseWriter) {
	r.ErrorStatus(w, http.StatusUnauthorized)
}

func (r *Rapidus) ErrorForbidden(w http.ResponseWriter) {
	r.ErrorStatus(w, http.StatusForbidden)
}
