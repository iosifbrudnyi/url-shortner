package redirect

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	resp "github.com/iosifbrudnyi/url-shortner/internal/lib/api/response"
	"github.com/iosifbrudnyi/url-shortner/internal/lib/logger/sl"
	"github.com/iosifbrudnyi/url-shortner/internal/storage"
	"log/slog"
	"net/http"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		alias := chi.URLParam(r, "alias")
		if alias == "" {
			errMsg := "alias is empty"
			log.Info(errMsg)
			render.JSON(w, r, resp.Error(errMsg))
			return
		}

		url, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			errMsg := "url not found"
			log.Error(errMsg)
			render.JSON(w, r, resp.Error(errMsg))
			return
		}
		if err != nil {
			errMsg := "failed to get url"
			log.Error(errMsg, sl.Err(err))
		}

		log.Info("got url", slog.String("url", url))

		http.Redirect(w, r, url, http.StatusFound)
	}
}
