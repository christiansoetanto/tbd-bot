package util

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	QAMovesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tbd_bot_qa_moves_total",
		Help: "Total number of Q&A questions moved and archived.",
	})

	UsersVettedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tbd_bot_users_vetted_total",
		Help: "Total number of users successfully vetted.",
	})

	TSUsersVettedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tbd_bot_ts_users_vetted_total",
		Help: "Total number of Terra Sancta users successfully vetted.",
	})

	CronExecutionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tbd_bot_cron_executions_total",
		Help: "Total number of cron job executions.",
	}, []string{"cron_name"})

	CMActionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tbd_bot_cm_actions_total",
		Help: "Total number of CM actions performed.",
	}, []string{"action"})

	MessagesProcessedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tbd_bot_messages_processed_total",
		Help: "Total number of messages processed.",
	}, []string{"handler"})

	ComponentInteractionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tbd_bot_component_interactions_total",
		Help: "Total number of component interactions.",
	}, []string{"component"})

	HandlerDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tbd_bot_handler_duration_seconds",
		Help:    "Histogram of handler execution durations in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"handler", "status"})

	HandlerRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tbd_bot_handler_requests_total",
		Help: "Total number of handler requests processed.",
	}, []string{"handler", "status"})

	registerOnce sync.Once
)

// InitMetrics registers custom Prometheus metrics with the default registerer.
func InitMetrics() {
	registerOnce.Do(func() {
		prometheus.MustRegister(QAMovesTotal)
		prometheus.MustRegister(UsersVettedTotal)
		prometheus.MustRegister(TSUsersVettedTotal)
		prometheus.MustRegister(CronExecutionsTotal)
		prometheus.MustRegister(CMActionsTotal)
		prometheus.MustRegister(MessagesProcessedTotal)
		prometheus.MustRegister(ComponentInteractionsTotal)
		prometheus.MustRegister(HandlerDuration)
		prometheus.MustRegister(HandlerRequestsTotal)
		prometheus.MustRegister(DiscordConnected)
		prometheus.MustRegister(GatewayEventsTotal)
		prometheus.MustRegister(ExternalAPIFailuresTotal)
		prometheus.MustRegister(lastHeartbeatAckTimestamp)
		prometheus.MustRegister(heartbeatLatencySeconds)
		prometheus.MustRegister(ExternalHeartbeatTotal)
		prometheus.MustRegister(externalHeartbeatLastPing)
		prometheus.MustRegister(externalHeartbeatEnabled)
	})
}

func init() {
	InitMetrics()
}

// IncQAMoves increments the tbd_bot_qa_moves_total counter by 1.
func IncQAMoves() {
	QAMovesTotal.Inc()
}

// IncUsersVetted increments the tbd_bot_users_vetted_total counter by 1.
func IncUsersVetted() {
	UsersVettedTotal.Inc()
}

// IncTSUsersVetted increments the tbd_bot_ts_users_vetted_total counter by 1.
func IncTSUsersVetted() {
	TSUsersVettedTotal.Inc()
}

// IncCronExecutions increments the tbd_bot_cron_executions_total counter for a given cron_name.
func IncCronExecutions(cronName string) {
	CronExecutionsTotal.WithLabelValues(cronName).Inc()
}

// IncCMActions increments the tbd_bot_cm_actions_total counter for a given action.
func IncCMActions(action string) {
	CMActionsTotal.WithLabelValues(action).Inc()
}

// IncMessagesProcessed increments the tbd_bot_messages_processed_total counter for a given handler.
func IncMessagesProcessed(handlerName string) {
	MessagesProcessedTotal.WithLabelValues(handlerName).Inc()
}

// IncComponentInteractions increments the tbd_bot_component_interactions_total counter for a given component.
func IncComponentInteractions(component string) {
	ComponentInteractionsTotal.WithLabelValues(component).Inc()
}

// RecordHandlerExecution observes latency and increments request total counter for a handler.
func RecordHandlerExecution(handlerName string, startTime time.Time, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	duration := time.Since(startTime).Seconds()
	HandlerDuration.WithLabelValues(handlerName, status).Observe(duration)
	HandlerRequestsTotal.WithLabelValues(handlerName, status).Inc()
}

// DecorateInteractionHandler wraps an interaction command handler returning an error with RED metrics observation.
func DecorateInteractionHandler(name string, fn func(s *discordgo.Session, i *discordgo.InteractionCreate) error) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) (err error) {
		start := time.Now()
		defer func() {
			RecordHandlerExecution(name, start, err)
		}()
		err = fn(s, i)
		return err
	}
}

// DecorateEventHandler wraps a Discord event handler function with RED metrics observation.
func DecorateEventHandler[T any](name string, fn func(s *discordgo.Session, event T)) func(s *discordgo.Session, event T) {
	return func(s *discordgo.Session, event T) {
		start := time.Now()
		defer func() {
			RecordHandlerExecution(name, start, nil)
		}()
		fn(s, event)
	}
}
