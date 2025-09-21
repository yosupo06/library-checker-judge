package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"github.com/yosupo06/library-checker-judge/database"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultMetricType = "custom.googleapis.com/judge/task_queue/pending"

func main() {
	http.HandleFunc("/", queueMetricsHandler)

	port := envOrDefault("PORT", "8080")
	log.Printf("queue-metrics service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func queueMetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectID := firstNonEmpty(os.Getenv("GCP_PROJECT"), os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if projectID == "" {
		http.Error(w, "missing project id", http.StatusInternalServerError)
		return
	}
	metricType := envOrDefault("METRIC_TYPE", defaultMetricType)

	db := database.Connect(database.GetDSNFromEnv(), false)
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("db.DB(): %v", err)
		http.Error(w, "db failure", http.StatusInternalServerError)
		return
	}
	defer sqlDB.Close()

	stats, err := database.FetchMonitoringData(db)
	if err != nil {
		log.Printf("FetchMonitoringData: %v", err)
		http.Error(w, "fetch monitoring data", http.StatusInternalServerError)
		return
	}

	if err := writePendingMetric(ctx, projectID, metricType, float64(stats.TaskQueue.PendingTasks)); err != nil {
		log.Printf("writePendingMetric: %v", err)
		http.Error(w, "write metric", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "pending_tasks=%d", stats.TaskQueue.PendingTasks)
}

func writePendingMetric(ctx context.Context, projectID, metricType string, value float64) error {
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return fmt.Errorf("NewMetricClient: %w", err)
	}
	defer client.Close()

	if err := ensureMetricDescriptor(ctx, client, projectID, metricType); err != nil {
		return err
	}

	series := &monitoringpb.TimeSeries{
		Metric: &metricpb.Metric{Type: metricType},
		Resource: &monitoredrespb.MonitoredResource{
			Type: "global",
			Labels: map[string]string{
				"project_id": projectID,
			},
		},
		Points: []*monitoringpb.Point{
			{
				Interval: &monitoringpb.TimeInterval{EndTime: timestamppb.New(time.Now())},
				Value:    &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_DoubleValue{DoubleValue: value}},
			},
		},
	}

	req := &monitoringpb.CreateTimeSeriesRequest{
		Name:       fmt.Sprintf("projects/%s", projectID),
		TimeSeries: []*monitoringpb.TimeSeries{series},
	}

	if err := client.CreateTimeSeries(ctx, req); err != nil {
		return fmt.Errorf("CreateTimeSeries: %w", err)
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func ensureMetricDescriptor(ctx context.Context, client *monitoring.MetricClient, projectID, metricType string) error {
	name := fmt.Sprintf("projects/%s/metricDescriptors/%s", projectID, metricType)
	if _, err := client.GetMetricDescriptor(ctx, &monitoringpb.GetMetricDescriptorRequest{Name: name}); err != nil {
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("GetMetricDescriptor: %w", err)
		}

		_, err = client.CreateMetricDescriptor(ctx, &monitoringpb.CreateMetricDescriptorRequest{
			Name: fmt.Sprintf("projects/%s", projectID),
			MetricDescriptor: &metricpb.MetricDescriptor{
				Type:        metricType,
				DisplayName: "Judge Pending Tasks",
				Description: "Pending tasks waiting in the judge queue",
				Unit:        "1",
				MetricKind:  metricpb.MetricDescriptor_GAUGE,
				ValueType:   metricpb.MetricDescriptor_DOUBLE,
			},
		})
		if err != nil && status.Code(err) != codes.AlreadyExists {
			return fmt.Errorf("CreateMetricDescriptor: %w", err)
		}
	}
	return nil
}
