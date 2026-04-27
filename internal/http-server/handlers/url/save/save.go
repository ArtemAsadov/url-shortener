package save

import (
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/alias"
	"url-shortener/internal/lib/api/response"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	render "github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"` //Rand if empty
}

type URLSaver interface {
	SaveUrl(urlToSave string, alias string) (int64, error)
}

// TODO: move to config
const aliasLength = 6

var aliasGen *alias.Generator = alias.New(aliasLength)

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.new"

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())), // Как он достает из машинерии всей этой request_id
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))

			return //render.JSON не прерывает работу
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validatoinErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.Error("invalid message"))
			render.JSON(w, r, resp.ValidationError(validatoinErr))

			return //render.JSON не прерывает работу
		}

		reqAlias := req.Alias
		if reqAlias == "" {
			reqAlias = aliasGen.GenerateFromURL(req.URL)
		}

		//TODO: отсекать повторение если в базе есть такой alias по такому URL ?
		id, err := urlSaver.SaveUrl(req.URL, reqAlias)
		if errors.Is(err, storage.ErrUrlExist) {
			log.Info("url already exist", slog.String("url", req.URL))

			//не палим в ошибке чуствительные данные наружу err
			render.JSON(w, r, resp.Error("url already exist"))

			return //render.JSON не прерывает работу
		}

		if err != nil {

			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add url"))

			return //render.JSON не прерывает работу
		}

		log.Info("url added", slog.Int64("id", id))

		render.JSON(w, r, toResponse(response.Ok(), req.Alias))
	}
}

func toResponse(resp resp.Response, alias string) Response {
	return Response{
		Response: resp,
		Alias:    alias,
	}
}
