package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricsRegistry struct {
	// Transações
	TransactionDuration prometheus.Histogram
	TransactionTotal    prometheus.Counter
	TransactionStatus   prometheus.CounterVec

	// Erros
	ErrorsTotal prometheus.CounterVec

	// Ledger
	ConsumerLag prometheus.Gauge
	OutboxPending prometheus.Gauge
	DLQMessagesTotal prometheus.Counter

	// Account
	BalanceGauge prometheus.GaugeVec

	// Database
	DatabaseQueryDuration prometheus.HistogramVec

	// Kafka Producer
	KafkaMessagesProduced prometheus.CounterVec

	// Circuit Breaker
	CircuitBreakerState prometheus.GaugeVec
}

func NewMetricsRegistry(serviceName string) (*MetricsRegistry, error) {
	registry := &MetricsRegistry{}

	transactionDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    fmt.Sprintf("%s_request_duration_seconds", serviceName),
			Help:    "Latência de processamento de transações em segundos",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
	)
	registry.TransactionDuration = transactionDuration

	transactionTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_transactions_total", serviceName),
			Help: "Total de transações processadas",
		},
	)
	registry.TransactionTotal = transactionTotal

	transactionStatus := *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_requests_total", serviceName),
			Help: "Total de transações por status",
		},
		[]string{"status", "method"},
	)
	registry.TransactionStatus = transactionStatus

	errorsTotal := *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_errors_total", serviceName),
			Help: "Total de erros",
		},
		[]string{"error_type"},
	)
	registry.ErrorsTotal = errorsTotal

	consumerLag := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_consumer_lag", serviceName),
			Help: "Lag do consumer Kafka em mensagens",
		},
	)
	registry.ConsumerLag = consumerLag

	outboxPending := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_outbox_pending", serviceName),
			Help: "Número de mensagens pendentes no outbox",
		},
	)
	registry.OutboxPending = outboxPending

	dlqMessagesTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_dlq_messages_total", serviceName),
			Help: "Total de mensagens em Dead Letter Queue",
		},
	)
	registry.DLQMessagesTotal = dlqMessagesTotal

	balanceGauge := *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_account_balance", serviceName),
			Help: "Saldo de cada conta",
		},
		[]string{"account_id"},
	)
	registry.BalanceGauge = balanceGauge

	databaseQueryDuration := *prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Latência de banco de dados em segundos",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation"},
	)
	registry.DatabaseQueryDuration = databaseQueryDuration

	kafkaMessagesProduced := *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_produced_total",
			Help: "Total de mensagens produzidas no Kafka por tópico",
		},
		[]string{"topic"},
	)
	registry.KafkaMessagesProduced = kafkaMessagesProduced

	circuitBreakerState := *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Estado dos circuit breakers (1 = nesse estado)",
		},
		[]string{"service", "state"},
	)
	registry.CircuitBreakerState = circuitBreakerState

	err := prometheus.DefaultRegisterer.Register(transactionDuration)
	if err != nil && !isAlreadyRegistered(err) {
		return nil, err
	}

	err = prometheus.DefaultRegisterer.Register(transactionTotal)
	if err != nil && !isAlreadyRegistered(err) {
		return nil, err
	}

	err = prometheus.DefaultRegisterer.Register(&transactionStatus)
	if err != nil && !isAlreadyRegistered(err) {
		return nil, err
	}

	err = prometheus.DefaultRegisterer.Register(&errorsTotal)
	if err != nil && !isAlreadyRegistered(err) {
		return nil, err
	}

	err = prometheus.DefaultRegisterer.Register(consumerLag)
	if err != nil && !isAlreadyRegistered(err) {
		return nil, err
	}

	err = prometheus.DefaultRegisterer.Register(outboxPending)
	if err != nil && !isAlreadyRegistered(err) {
		return nil, err
	}

	err = prometheus.DefaultRegisterer.Register(dlqMessagesTotal)
	if err != nil && !isAlreadyRegistered(err) {
		return nil, err
	}

	err = prometheus.DefaultRegisterer.Register(&balanceGauge)
	if err != nil && !isAlreadyRegistered(err) {
		return nil, err
	}

	err = prometheus.DefaultRegisterer.Register(&databaseQueryDuration)
	if err != nil && !isAlreadyRegistered(err) {
		return nil, err
	}

	err = prometheus.DefaultRegisterer.Register(&kafkaMessagesProduced)
	if err != nil && !isAlreadyRegistered(err) {
		return nil, err
	}

	err = prometheus.DefaultRegisterer.Register(&circuitBreakerState)
	if err != nil && !isAlreadyRegistered(err) {
		return nil, err
	}

	return registry, nil
}

func isAlreadyRegistered(err error) bool {
	_, ok := err.(prometheus.AlreadyRegisteredError)
	return ok
}
