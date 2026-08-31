package httpadapter

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"payment-sandbox/internal/domain"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func Observability(next http.Handler, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		ctx, span := otel.Tracer("payment-sandbox/http").Start(ctx, r.Method+" "+r.URL.Path,
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
				attribute.String("request.id", requestID),
			),
		)
		defer span.End()

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		rec.Header().Set("X-Request-Id", requestID)

		defer func() {
			if recovered := recover(); recovered != nil {
				span.RecordError(toError(recovered))
				span.SetStatus(codes.Error, "panic")
				logger.ErrorContext(ctx, "request panic",
					"request_id", requestID,
					"method", r.Method,
					"path", r.URL.Path,
					"panic", recovered,
				)
				writeJSON(rec, http.StatusInternalServerError, map[string]any{"error": domain.NewError(http.StatusInternalServerError, "internal_error", "internal server error")})
				return
			}

			status := rec.status
			if status == 0 {
				status = http.StatusOK
			}
			if status >= 500 {
				span.SetStatus(codes.Error, http.StatusText(status))
			}
			span.SetAttributes(
				attribute.Int("http.status_code", status),
				attribute.Int64("http.response_size", int64(rec.bytes)),
			)

			traceID, spanID := traceIDs(ctx)
			logger.InfoContext(ctx, "request completed",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration_ms", time.Since(start).Milliseconds(),
				"trace_id", traceID,
				"span_id", spanID,
			)
		}()

		next.ServeHTTP(rec, r.WithContext(ctx))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusRecorder) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

func traceIDs(ctx context.Context) (string, string) {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return "", ""
	}
	return spanCtx.TraceID().String(), spanCtx.SpanID().String()
}

func toError(v any) error {
	if err, ok := v.(error); ok {
		return err
	}
	return domain.NewError(http.StatusInternalServerError, "panic", "panic")
}
