package middleware

import (
	"context"
	"net/http"
	"time"
)

type HTTPMetricsRecorder interface {
	RecordHTTPRequest(ctx context.Context, method, route string, status int, duration time.Duration)
}

func Metrics(next http.Handler, recorder HTTPMetricsRecorder) http.Handler {
	if recorder == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		recorder.RecordHTTPRequest(r.Context(), r.Method, routeLabel(r), rec.status, time.Since(start))
	})
}
