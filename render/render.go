package render

import (
	"context"
	"github.com/a-h/templ"
	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
	"net/http"
)

type Render struct {
	Session *scs.SessionManager
}

func (ren *Render) Template(w http.ResponseWriter, r *http.Request, template templ.Component) error {

	// Create a context and set value(s) that will be available to all templates
	ctx := context.WithValue(r.Context(), "CSRFToken", nosurf.Token(r))
	ctx = context.WithValue(ctx, "success", ren.Session.PopString(r.Context(), "success"))
	ctx = context.WithValue(ctx, "warning", ren.Session.PopString(r.Context(), "warning"))
	ctx = context.WithValue(ctx, "error", ren.Session.PopString(r.Context(), "error"))

	return template.Render(ctx, w)
}
