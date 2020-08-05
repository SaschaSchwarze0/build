package prometheus

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	buildv1alpha1 "github.com/redhat-developer/build/pkg/apis/build/v1alpha1"
	"github.com/redhat-developer/build/pkg/config"
	"github.com/redhat-developer/build/pkg/ctxlog"
	"github.com/redhat-developer/build/pkg/monitoring"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	labelNamespace         = "namespace"
	labelBuildName         = "build_name"
	labelBuildStrategyName = "build_strategy_name"
)

type prometheusConsumer struct {
	buildRunCreatedToRunning   *prometheus.HistogramVec
	buildRunCreatedToSucceeded *prometheus.HistogramVec
	buildRunFailedTotal        *prometheus.CounterVec
	buildRunSucceededTotal     *prometheus.CounterVec
	buildRunTotal              *prometheus.CounterVec
}

func (p *prometheusConsumer) V1Alpha1BuildRunCreated(ctx context.Context, buildRun *buildv1alpha1.BuildRun) {
	ctx = ctxlog.NewContext(ctx, "prometheus")
	ctxlog.Debug(ctx, "V1Alpha1BuildRunCreated", "namespace", buildRun.Namespace, "name", buildRun.Name)

	p.buildRunTotal.With(prometheus.Labels{
		labelNamespace: buildRun.Namespace,
		labelBuildName: buildRun.Spec.BuildRef.Name,
		// TODO is not yet set here
		// labelBuildStrategyName: buildRun.Status.BuildSpec.StrategyRef.Name,
		labelBuildStrategyName: "",
	}).Inc()
}

func (p *prometheusConsumer) V1Alpha1BuildRunFailed(ctx context.Context, buildRun *buildv1alpha1.BuildRun, taskRun *tektonv1beta1.TaskRun) {
	ctx = ctxlog.NewContext(ctx, "prometheus")
	ctxlog.Debug(ctx, "V1Alpha1BuildRunFailed", "namespace", buildRun.Namespace, "name", buildRun.Name)

	p.buildRunFailedTotal.With(prometheus.Labels{
		labelNamespace:         buildRun.Namespace,
		labelBuildName:         buildRun.Spec.BuildRef.Name,
		labelBuildStrategyName: buildRun.Status.BuildSpec.StrategyRef.Name,
	}).Inc()
}

func (p *prometheusConsumer) V1Alpha1BuildRunRunning(ctx context.Context, buildRun *buildv1alpha1.BuildRun, taskRun *tektonv1beta1.TaskRun) {
	ctx = ctxlog.NewContext(ctx, "prometheus")
	ctxlog.Debug(ctx, "V1Alpha1BuildRunRunning", "namespace", buildRun.Namespace, "name", buildRun.Name)

	duration := buildRun.Status.StartTime.Time.Sub(buildRun.CreationTimestamp.Time).Seconds()

	ctxlog.Debug(ctx, "Timing info", "StartTime", buildRun.Status.StartTime, "CreationTimestamp", buildRun.CreationTimestamp, "Duration", duration)

	p.buildRunCreatedToRunning.With(prometheus.Labels{
		labelNamespace:         buildRun.Namespace,
		labelBuildName:         buildRun.Spec.BuildRef.Name,
		labelBuildStrategyName: buildRun.Status.BuildSpec.StrategyRef.Name,
	}).Observe(duration)
}

func (p *prometheusConsumer) V1Alpha1BuildRunSucceeded(ctx context.Context, buildRun *buildv1alpha1.BuildRun, taskRun *tektonv1beta1.TaskRun) {
	ctx = ctxlog.NewContext(ctx, "prometheus")
	ctxlog.Debug(ctx, "V1Alpha1BuildRunSucceeded", "namespace", buildRun.Namespace, "name", buildRun.Name)

	duration := buildRun.Status.CompletionTime.Time.Sub(buildRun.CreationTimestamp.Time).Seconds()

	p.buildRunCreatedToSucceeded.With(prometheus.Labels{
		labelNamespace:         buildRun.Namespace,
		labelBuildName:         buildRun.Spec.BuildRef.Name,
		labelBuildStrategyName: buildRun.Status.BuildSpec.StrategyRef.Name,
	}).Observe(duration)

	p.buildRunSucceededTotal.With(prometheus.Labels{
		labelNamespace:         buildRun.Namespace,
		labelBuildName:         buildRun.Spec.BuildRef.Name,
		labelBuildStrategyName: buildRun.Status.BuildSpec.StrategyRef.Name,
	}).Inc()
}

// InitPrometheus initializes the prometheus stuff
func InitPrometheus(config *config.Config) {
	// TODO check if prometheus is enabled in config and only initialize if so

	fmt.Println("SASCHA in init of prometheus")

	consumer := &prometheusConsumer{
		buildRunFailedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "buildrun_failed_total",
				Help: "Total number of failed buildruns.",
			},
			[]string{
				labelNamespace,
				labelBuildName,
				labelBuildStrategyName,
			},
		),

		buildRunSucceededTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "buildrun_succeeded_total",
				Help: "Total number of successful buildruns.",
			},
			[]string{
				labelNamespace,
				labelBuildName,
				labelBuildStrategyName,
			},
		),

		buildRunTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "buildrun_total",
				Help: "Total number of created buildruns.",
			},
			[]string{
				labelNamespace,
				labelBuildName,
				labelBuildStrategyName,
			},
		),

		buildRunCreatedToRunning: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "buildrun_created_to_running",
				Help:    "Duration in seconds from the creation of a buildrun to it entering the Running state.",
				Buckets: prometheus.LinearBuckets(0, 1, 11),
			},
			[]string{
				labelNamespace,
				labelBuildName,
				labelBuildStrategyName,
			},
		),

		buildRunCreatedToSucceeded: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "buildrun_created_to_succeeded",
				Help:    "Duration in seconds from the creation of a buildrun to it ending successful.",
				Buckets: prometheus.ExponentialBuckets(10, 1.5, 14),
			},
			[]string{
				labelNamespace,
				labelBuildName,
				labelBuildStrategyName,
			},
		),
	}

	metrics.Registry.MustRegister(
		consumer.buildRunFailedTotal,
		consumer.buildRunSucceededTotal,
		consumer.buildRunTotal,
		consumer.buildRunCreatedToRunning,
		consumer.buildRunCreatedToSucceeded,
	)

	monitoring.RegisterConsumer(consumer)

}
