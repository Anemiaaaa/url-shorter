package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"
)

// TODO: move to config
const aliasLength = 8

type Request struct {
	URLToSave string `json:"url" validate:"required,url"`
	Alias     string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2 --name=URLServer --output=./mocks
type URLServer interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Логирование запроса
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Декодирование JSON тела запроса
		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to decode request", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("Request body decoded", slog.Any("request", req))

		// Валидация
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		// Генерация строки если алиас не указан
		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		// Сохранение URL
		id, err := urlSaver.SaveURL(req.URLToSave, alias)

		// Проверка на существование URL
		if errors.Is(err, storage.ErrURLExists) {
			log.Error("URL already exists", slog.String("url", req.URLToSave))

			render.JSON(w, r, resp.Error("url already exists"))

			return
		}

		// Другие ошибки
		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add url"))

			return
		}

		log.Info("URL saved", slog.Int64("id", id), slog.String("alias", alias))

		// Формирование ответа
		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
