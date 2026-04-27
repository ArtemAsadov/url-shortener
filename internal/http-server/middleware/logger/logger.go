package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor) // 1. Обёртка сразу
			t1 := time.Now()                                        // 2. Старт таймера

			// 3. Выполняем запрос
			next.ServeHTTP(ww, r)

			// 4. Логируем ПОСЛЕ (уже без defer, просто по порядку)
			// Это работает, потому что ServeHTTP блокирует выполнение до конца
			log.Info("request_completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("request_id", middleware.GetReqID(r.Context())),
				slog.Int("status", ww.Status()),           // ← Теперь видим реальный статус
				slog.Int("bytes", ww.BytesWritten()),      // ← И размер ответа
				slog.Duration("duration", time.Since(t1)), // ← И время выполнения
			)
		})
	}
}
