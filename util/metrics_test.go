package util_test

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestBusinessMetricsRegistered(t *testing.T) {
	// Verify QAMovesTotal is registered and increments correctly
	beforeQA := testutil.ToFloat64(util.QAMovesTotal)
	util.IncQAMoves()
	afterQA := testutil.ToFloat64(util.QAMovesTotal)
	if afterQA-beforeQA != 1 {
		t.Fatalf("expected QAMovesTotal to increment by 1, got before=%f, after=%f", beforeQA, afterQA)
	}

	// Verify UsersVettedTotal is registered and increments correctly
	beforeVetted := testutil.ToFloat64(util.UsersVettedTotal)
	util.IncUsersVetted()
	afterVetted := testutil.ToFloat64(util.UsersVettedTotal)
	if afterVetted-beforeVetted != 1 {
		t.Fatalf("expected UsersVettedTotal to increment by 1, got before=%f, after=%f", beforeVetted, afterVetted)
	}

	// Gather metrics from default gatherer and verify metric names exist
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	foundQA := false
	foundVetted := false
	for _, mf := range metricFamilies {
		if mf.GetName() == "tbd_bot_qa_moves_total" {
			foundQA = true
		}
		if mf.GetName() == "tbd_bot_users_vetted_total" {
			foundVetted = true
		}
	}

	if !foundQA {
		t.Errorf("metric tbd_bot_qa_moves_total not found in default gatherer")
	}
	if !foundVetted {
		t.Errorf("metric tbd_bot_users_vetted_total not found in default gatherer")
	}
}

func TestREDMetricsDecorators(t *testing.T) {
	// Test DecorateInteractionHandler with success
	dummySuccessInteraction := util.DecorateInteractionHandler("test_interaction_success", func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		return nil
	})
	err := dummySuccessInteraction(nil, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Test DecorateInteractionHandler with error
	dummyErrorInteraction := util.DecorateInteractionHandler("test_interaction_error", func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		return errors.New("something went wrong")
	})
	err = dummyErrorInteraction(nil, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	// Test DecorateEventHandler
	dummyEvent := util.DecorateEventHandler("test_event", func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// do nothing
	})
	dummyEvent(nil, nil)

	// Gather metrics and verify RED metrics presence
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	foundDuration := false
	foundRequests := false
	for _, mf := range metricFamilies {
		if mf.GetName() == "tbd_bot_handler_duration_seconds" {
			foundDuration = true
		}
		if mf.GetName() == "tbd_bot_handler_requests_total" {
			foundRequests = true
		}
	}

	if !foundDuration {
		t.Errorf("metric tbd_bot_handler_duration_seconds not found in default gatherer")
	}
	if !foundRequests {
		t.Errorf("metric tbd_bot_handler_requests_total not found in default gatherer")
	}

	// Check specific counter values using testutil
	successCount := testutil.ToFloat64(util.HandlerRequestsTotal.WithLabelValues("test_interaction_success", "success"))
	if successCount != 1 {
		t.Errorf("expected 1 success count for test_interaction_success, got %f", successCount)
	}

	errorCount := testutil.ToFloat64(util.HandlerRequestsTotal.WithLabelValues("test_interaction_error", "error"))
	if errorCount != 1 {
		t.Errorf("expected 1 error count for test_interaction_error, got %f", errorCount)
	}

	eventCount := testutil.ToFloat64(util.HandlerRequestsTotal.WithLabelValues("test_event", "success"))
	if eventCount != 1 {
		t.Errorf("expected 1 success count for test_event, got %f", eventCount)
	}
}

func TestAllExpandedBusinessMetrics(t *testing.T) {
	util.IncCronExecutions("test_cron")
	util.IncTSUsersVetted()
	util.IncCMActions("test_action")
	util.IncMessagesProcessed("test_msg")
	util.IncComponentInteractions("test_component")

	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	expectedMetrics := map[string]bool{
		"tbd_bot_cron_executions_total":        false,
		"tbd_bot_ts_users_vetted_total":        false,
		"tbd_bot_cm_actions_total":             false,
		"tbd_bot_messages_processed_total":     false,
		"tbd_bot_component_interactions_total": false,
	}

	for _, mf := range metricFamilies {
		if _, ok := expectedMetrics[mf.GetName()]; ok {
			expectedMetrics[mf.GetName()] = true
		}
	}

	for name, found := range expectedMetrics {
		if !found {
			t.Errorf("metric %s not found in default gatherer", name)
		}
	}
}


