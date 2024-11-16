package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	resp "github.com/iosifbrudnyi/url-shortner/internal/lib/api/response"
	"github.com/iosifbrudnyi/url-shortner/internal/lib/logger/sl"
	"github.com/iosifbrudnyi/url-shortner/internal/lib/random"
	"github.com/iosifbrudnyi/url-shortner/internal/storage"
	"io"
	"log/slog"
	"net/http"
)

// TODO: move to config
const aliasLength = 6

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias"`
}

type URLSaver interface {
	SaveURL(url, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			errMsg := "request body is empty"
			log.Error(errMsg)
			render.JSON(w, r, resp.Error(errMsg))
			return
		}
		if err != nil {
			errMsg := "failed to decode request body"
			log.Error(errMsg, sl.Err(err))
			render.JSON(w, r, resp.Error(errMsg))
			return
		}

		log.Info("request body decoded", slog.Any("req", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, resp.Error(validateErr.Error()))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			errMsg := "url already exists"
			log.Info(errMsg, slog.String("url", req.URL))
			render.JSON(w, r, resp.Error(errMsg))
			return
		}
		if err != nil {
			errMsg := "failed to add url"
			log.Error(errMsg, sl.Err(err))
			render.JSON(w, r, resp.Error(errMsg))
			return
		}

		log.Info("url added", slog.Int64("id", id))
		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
