package monitoring

import (
	"context"

	buildv1alpha1 "github.com/redhat-developer/build/pkg/apis/build/v1alpha1"
	"github.com/redhat-developer/build/pkg/ctxlog"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

// Consumer is an interface with functions for all monitoring events that a consumer needs to implement
type Consumer interface {

	// V1Alpha1BuildRunCreated is a callback invoked on creation of a build run
	V1Alpha1BuildRunCreated(context.Context, *buildv1alpha1.BuildRun)

	// V1Alpha1BuildRunFailed is a callback invoked on failed completion of a build run, the task run may be nil
	V1Alpha1BuildRunFailed(context.Context, *buildv1alpha1.BuildRun, *tektonv1beta1.TaskRun)

	// V1Alpha1BuildRunFailed is a callback invoked when a build run enters the Running state
	V1Alpha1BuildRunRunning(context.Context, *buildv1alpha1.BuildRun, *tektonv1beta1.TaskRun)

	// V1Alpha1BuildRunFailed is a callback invoked on successful completion of a build run
	V1Alpha1BuildRunSucceeded(context.Context, *buildv1alpha1.BuildRun, *tektonv1beta1.TaskRun)
}

var consumers []Consumer

// RegisterConsumer is used to register a monitoring consumer. It will subsequently be called for all
// monitoring events.
func RegisterConsumer(consumer Consumer) {
	consumers = append(consumers, consumer)
}

// OnV1Alpha1BuildRunCreated is the monitoring event for the creation of a BuildRun
func OnV1Alpha1BuildRunCreated(ctx context.Context, buildRun *buildv1alpha1.BuildRun) {
	ctx = ctxlog.NewContext(ctx, "monitoring")
	ctxlog.Debug(ctx, "OnV1Alpha1BuildRunCreated", "namespace", buildRun.Namespace, "name", buildRun.Name, "consumerCount", len(consumers))

	for _, consumer := range consumers {
		consumer.V1Alpha1BuildRunCreated(ctx, buildRun)
	}
}

// OnV1Alpha1BuildRunFailed is the monitoring event for the for a BuildRun to be failed
func OnV1Alpha1BuildRunFailed(ctx context.Context, buildRun *buildv1alpha1.BuildRun, taskRun *tektonv1beta1.TaskRun) {
	ctx = ctxlog.NewContext(ctx, "monitoring")
	ctxlog.Debug(ctx, "OnV1Alpha1BuildRunFailed", "namespace", buildRun.Namespace, "name", buildRun.Name, "consumerCount", len(consumers))

	for _, consumer := range consumers {
		consumer.V1Alpha1BuildRunFailed(ctx, buildRun, taskRun)
	}
}

// OnV1Alpha1BuildRunRunning is the monitoring event for a BuildRun to become Running
func OnV1Alpha1BuildRunRunning(ctx context.Context, buildRun *buildv1alpha1.BuildRun, taskRun *tektonv1beta1.TaskRun) {
	ctx = ctxlog.NewContext(ctx, "monitoring")
	ctxlog.Debug(ctx, "OnV1Alpha1BuildRunRunning", "namespace", buildRun.Namespace, "name", buildRun.Name, "consumerCount", len(consumers))

	for _, consumer := range consumers {
		consumer.V1Alpha1BuildRunRunning(ctx, buildRun, taskRun)
	}
}

// OnV1Alpha1BuildRunSucceeded is the monitoring event for a BuildRun to be successful
func OnV1Alpha1BuildRunSucceeded(ctx context.Context, buildRun *buildv1alpha1.BuildRun, taskRun *tektonv1beta1.TaskRun) {
	ctx = ctxlog.NewContext(ctx, "monitoring")
	ctxlog.Debug(ctx, "OnV1Alpha1BuildRunSucceeded", "namespace", buildRun.Namespace, "name", buildRun.Name, "consumerCount", len(consumers))

	for _, consumer := range consumers {
		consumer.V1Alpha1BuildRunSucceeded(ctx, buildRun, taskRun)
	}
}
