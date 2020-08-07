package newrelic

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"

	buildv1alpha1 "github.com/redhat-developer/build/pkg/apis/build/v1alpha1"
	"github.com/redhat-developer/build/pkg/config"
	"github.com/redhat-developer/build/pkg/ctxlog"
	"github.com/redhat-developer/build/pkg/monitoring"

	newrelicconfig "github.com/newrelic/newrelic-client-go/pkg/config"
	"github.com/newrelic/newrelic-client-go/pkg/events"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	knativeapis "knative.dev/pkg/apis"
)

var (
	// sample message: "step-image-digest-exporter-57xvn" exited with code 1 (image: "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/imagedigestexporter@sha256:0c3d2f9cafca27f10d99dd9147c22111eb403a6581c0ad728302f3bd755aa536"); for logs run: kubectl -n 92701b11-1997 logs npm-simple-normal-icr2-1-m4g2p-pod-h8dbb -c step-image-digest-exporter-57xvn"
	failedStepMessageRegex = regexp.MustCompile("^\"step-([0-9a-zA-Z-]{1,})\"\\sexited\\swith\\scode\\s([0-9]{1,})\\s")
)

type newrelicConsumer struct {
	client    events.Events
	eventType string
}

func (n *newrelicConsumer) createEventData(buildRun *buildv1alpha1.BuildRun, taskRun *tektonv1beta1.TaskRun) ([]byte, error) {
	nrEvent := make(map[string]interface{})

	nrEvent["eventType"] = n.eventType
	nrEvent["timestamp"] = buildRun.CreationTimestamp.Unix() * 1000

	nrEvent["cluster"] = os.Getenv("CLUSTER_NAME")
	nrEvent["controllerPod"] = os.Getenv("POD_NAME")

	nrEvent["namespace"] = buildRun.Namespace

	nrEvent["buildrun"] = buildRun.Name
	nrEvent["buildrunId"] = string(buildRun.UID)
	nrEvent["build"] = buildRun.Spec.BuildRef.Name
	if buildRun.Status.BuildSpec != nil {
		nrEvent["buildstrategy"] = buildRun.Status.BuildSpec.StrategyRef.Name
	}

	if buildRun.Status.Succeeded == corev1.ConditionTrue {
		nrEvent["successful"] = "true"
	}
	if buildRun.Status.Succeeded == corev1.ConditionFalse {
		nrEvent["successful"] = "false"
	}

	nrEvent["creationTime"] = buildRun.CreationTimestamp.Unix()
	if buildRun.Status.StartTime != nil {
		nrEvent["startTime"] = buildRun.Status.StartTime.Unix()
	}
	if buildRun.Status.CompletionTime != nil {
		nrEvent["completionTime"] = buildRun.Status.CompletionTime.Unix()
	}

	if taskRun != nil {
		if taskRun.Status.PodName != "" {
			nrEvent["taskRunPod"] = taskRun.Status.PodName
		}

		// Add information about completed steps
		if taskRun.Status.Steps != nil {
			for _, step := range taskRun.Status.Steps {
				if step.Terminated != nil {
					stepName := createNewRelicKey(step.Name)

					nrEvent[stepName+"Status"] = step.Terminated.ExitCode
					nrEvent[stepName+"StartTime"] = step.Terminated.StartedAt.Unix()
					nrEvent[stepName+"CompletionTime"] = step.Terminated.FinishedAt.Unix()
				}
			}
		}

		// the failed step can be found in the conditions, this is a little uggly
		condition := taskRun.Status.GetCondition(knativeapis.ConditionSucceeded)
		if condition != nil && condition.Status == corev1.ConditionFalse && failedStepMessageRegex.MatchString(condition.Message) {
			groups := failedStepMessageRegex.FindStringSubmatch(condition.Message)

			stepName := createNewRelicKey(groups[1])
			exitCode, err := strconv.Atoi(groups[2])
			if err != nil {
				return nil, err
			}

			nrEvent[stepName+"Status"] = exitCode
			nrEvent["failedContainer"] = removeRandomSuffix(groups[1])
		}

		// look for insights from the task run result
		for _, taskRunResult := range taskRun.Status.TaskRunResults {
			if strings.HasPrefix(taskRunResult.Name, "insights-") {
				key := taskRunResultNameToNewRelicKey(taskRunResult.Name)
				value := strings.TrimSpace(taskRunResult.Value)

				intValue, err := strconv.Atoi(value)
				if err == nil {
					nrEvent[key] = intValue
					continue
				}

				floatValue, err := strconv.ParseFloat(value, 64)
				if err == nil {
					nrEvent[key] = floatValue
					continue
				}

				nrEvent[key] = value
			}
		}
	}

	return json.Marshal(nrEvent)
}

func (n *newrelicConsumer) V1Alpha1BuildRunCreated(ctx context.Context, buildRun *buildv1alpha1.BuildRun) {
}

func (n *newrelicConsumer) V1Alpha1BuildRunFailed(ctx context.Context, buildRun *buildv1alpha1.BuildRun, taskRun *tektonv1beta1.TaskRun) {
	ctx = ctxlog.NewContext(ctx, "newrelic")
	ctxlog.Debug(ctx, "V1Alpha1BuildRunFailed", "namespace", buildRun.Namespace, "name", buildRun.Name)

	event, err := n.createEventData(buildRun, taskRun)
	if err != nil {
		ctxlog.Error(ctx, err, "Event data cannot be created.")
		return
	}

	if err := n.client.EnqueueEvent(ctx, event); err != nil {
		ctxlog.Error(ctx, err, "Event cannot be queued.")
	}
}

func (n *newrelicConsumer) V1Alpha1BuildRunRunning(ctx context.Context, buildRun *buildv1alpha1.BuildRun, taskRun *tektonv1beta1.TaskRun) {
}

func (n *newrelicConsumer) V1Alpha1BuildRunSucceeded(ctx context.Context, buildRun *buildv1alpha1.BuildRun, taskRun *tektonv1beta1.TaskRun) {
	ctx = ctxlog.NewContext(ctx, "newrelic")
	ctxlog.Debug(ctx, "V1Alpha1BuildRunSucceeded", "namespace", buildRun.Namespace, "name", buildRun.Name)

	event, err := n.createEventData(buildRun, taskRun)
	if err != nil {
		ctxlog.Error(ctx, err, "Event data cannot be created.")
		return
	}

	if err := n.client.EnqueueEvent(ctx, event); err != nil {
		ctxlog.Error(ctx, err, "Event cannot be queued.")
	}
}

// InitNewRelic initializes the New Relic stuff
func InitNewRelic(ctx context.Context, config *config.Config) error {
	ctx = ctxlog.NewContext(ctx, "newrelic")

	if config.NewRelic.AccountID != 0 && config.NewRelic.EventType != "" && config.NewRelic.InsightsAPIKey != "" && config.NewRelic.InsightsBaseURL != "" {
		ctxlog.Info(ctx, "Registering New Relic monitoring consumer")

		cfg := newrelicconfig.New()
		cfg.InsightsInsertKey = config.NewRelic.InsightsAPIKey
		// cfg.LogLevel = "trace"
		cfg.Region().SetInsightsBaseURL(config.NewRelic.InsightsBaseURL)

		client := events.New(cfg)

		if err := client.BatchMode(ctx, config.NewRelic.AccountID); err != nil {
			return err
		}

		consumer := &newrelicConsumer{
			client:    client,
			eventType: config.NewRelic.EventType,
		}

		monitoring.RegisterConsumer(consumer)
	}

	return nil
}

func removeRandomSuffix(stepName string) string {
	// the step might be a generated name ending with - and five random numbers/characters, these are well-known, so fix them here:
	if strings.HasPrefix(stepName, "create-dir-image-") {
		return "create-dir-image"
	} else if strings.HasPrefix(stepName, "git-source-source-") {
		return "git-source-source"
	} else if strings.HasPrefix(stepName, "image-digest-exporter-") {
		return "image-digest-exporter"
	}

	return stepName
}

func createNewRelicKey(stepName string) string {
	stepName = removeRandomSuffix(stepName)

	// convert to New Relic key, e. g. create-dir-image -> containerCreateDirImage
	stepNameParts := strings.Split(stepName, "-")
	for i, stepNamePart := range stepNameParts {
		if len(stepNamePart) == 1 {
			stepNameParts[i] = strings.ToUpper(stepNamePart)
		} else if len(stepNamePart) > 1 {
			stepNameParts[i] = strings.ToUpper(string(stepNamePart[0])) + string(stepNamePart[1:])
		}
	}

	return "container" + strings.Join(stepNameParts, "")
}

func taskRunResultNameToNewRelicKey(taskRunResultName string) string {
	taskRunResultNameParts := strings.Split(string(taskRunResultName[9:]), "-")
	for i, taskRunResultNamePart := range taskRunResultNameParts {
		if i == 0 {
			taskRunResultNameParts[i] = strings.ToLower(taskRunResultNamePart)
		} else if len(taskRunResultNamePart) == 1 {
			taskRunResultNameParts[i] = strings.ToUpper(taskRunResultNamePart)
		} else if len(taskRunResultNamePart) > 1 {
			taskRunResultNameParts[i] = strings.ToUpper(string(taskRunResultNamePart[0])) + strings.ToLower(string(taskRunResultNamePart[1:]))
		}
	}

	return strings.Join(taskRunResultNameParts, "")
}
