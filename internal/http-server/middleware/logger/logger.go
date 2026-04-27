package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	// ЭТАП 1: Вызывается ОДИН раз при старте приложения
	log.Debug(">>> [MIDDLEWARE FACTORY] New() called",
		slog.String("stage", "factory_init"),
	)

	return func(next http.Handler) http.Handler {
		// ЭТАП 2: Вызывается ОДИН раз при сборке роутера
		log.Debug(">>> [WRAPPER] Creating handler wrapper",
			slog.String("stage", "wrapper_init"),
			slog.String("next_handler_type", httpHandlerType(next)), // вспомогательная функция ниже
		)

		// Создаём "локальный" логгер с префиксом
		log := log.With(slog.String("component", "middleware/logger"))
		log.Info("logger middleware enabled", slog.String("stage", "wrapper_ready"))

		fn := func(w http.ResponseWriter, r *http.Request) {
			// ЭТАП 3: Вызывается НА КАЖДЫЙ запрос

			// 3.1. Начало обработки запроса
			log.Debug(">>> [REQUEST] Start processing",
				slog.String("stage", "request_start"),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			// 3.2. Обогащаем логгер данными из запроса
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			entry.Debug(">>> [REQUEST] Logger entry created", slog.String("stage", "entry_ready"))

			// 3.3. Оборачиваем ResponseWriter
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			entry.Debug(">>> [REQUEST] ResponseWriter wrapped",
				slog.String("stage", "writer_wrapped"),
				slog.Int("proto_major", r.ProtoMajor),
			)

			// 3.4. Замер времени
			t1 := time.Now()
			entry.Debug(">>> [REQUEST] Timer started",
				slog.String("stage", "timer_start"),
				slog.Time("start_time", t1),
			)

			// 3.5. Дефер для логирования ПОСЛЕ выполнения
			defer func() {
				duration := time.Since(t1)
				entry.Info(">>> [REQUEST] Completed",
					slog.String("stage", "request_end"),
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", duration.String()),
					slog.Float64("duration_ms", float64(duration.Milliseconds())),
				)
			}()

			// 3.6. ПЕРЕДАЁМ УПРАВЛЕНИЕ ДАЛЬШЕ (самый важный момент!)
			entry.Debug(">>> [REQUEST] Calling next.ServeHTTP()",
				slog.String("stage", "calling_next"),
			)

			next.ServeHTTP(ww, r)

			entry.Debug(">>> [REQUEST] Returned from next.ServeHTTP()",
				slog.String("stage", "returned_from_next"),
			)
		}

		log.Debug("<<< [WRAPPER] Returning http.HandlerFunc", slog.String("stage", "wrapper_done"))
		return http.HandlerFunc(fn)
	}
}

// ВСПОМОГАТЕЛЬНАЯ: чтобы видеть тип хендлера в логах
func httpHandlerType(h http.Handler) string {
	if h == nil {
		return "nil"
	}
	// Простое отражение типа без импорта reflect для краткости
	// В проде можно убрать или заменить на более умную логику
	return "http.Handler"
}
