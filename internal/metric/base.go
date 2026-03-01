package metric

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	// Метрика для RPS (счетчик запросов)
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "go_base_requests_total",
			Help: "Общее количество запросов к эндпоинту",
		},
		[]string{"method", "status"}, // Лейблы для фильтрации
	)

	// Метрика для времени ответа (гистограмма)
	responseDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "go_base_response_duration_seconds",
			Help:    "Длительность ответа в секундах",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5}, // Бакеты важны для точности
		},
		[]string{"method"},
	)
)

// IncRequestTotal ...
func IncRequestTotal(method, code string) {
	requestsTotal.WithLabelValues(method, code).Inc()
}

// SetResponseDurationSeconds ...
func SetResponseDurationSeconds(duraion time.Duration, method string) {
	responseDurationSeconds.WithLabelValues(method).Observe(duraion.Seconds())
}

// UnaryServerInterceptor возвращает gRPC unary interceptor, который собирает метрики
// для ваших Prometheus-метрик.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		// Извлекаем "чистое" имя метода, например /user.v1.UserService/CreateUser
		fullMethod := info.FullMethod

		// Выполняем обработчик
		resp, err := handler(ctx, req)

		duration := time.Since(start)

		// Определяем gRPC-код статуса
		st, _ := status.FromError(err)
		code := st.Code()
		codeStr := code.String() // "OK", "INVALID_ARGUMENT", "NOT_FOUND" и т.д.

		// Увеличиваем счётчик запросов
		IncRequestTotal(fullMethod, codeStr)

		// Записываем длительность
		SetResponseDurationSeconds(duration, fullMethod)

		return resp, err
	}
}
