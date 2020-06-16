package monitoring_test

import (
	"context"
	"fmt"

	buildv1alpha1 "github.com/redhat-developer/build/pkg/apis/build/v1alpha1"
	"github.com/redhat-developer/build/pkg/monitoring"
)

type testConsumer struct {
	V1Alpha1BuildRunCreatedCallCount   int
	V1Alpha1BuildRunFailedCallCount    int
	V1Alpha1BuildRunRunningCallCount   int
	V1Alpha1BuildRunSucceededCallCount int
}

func newTestConsumer() *testConsumer {
	return &testConsumer{
		V1Alpha1BuildRunCreatedCallCount:   0,
		V1Alpha1BuildRunFailedCallCount:    0,
		V1Alpha1BuildRunRunningCallCount:   0,
		V1Alpha1BuildRunSucceededCallCount: 0,
	}
}

func (t *testConsumer) V1Alpha1BuildRunCreated(ctx context.Context, buildRun *buildv1alpha1.BuildRun) {
	fmt.Println("Incrementing V1Alpha1BuildRunCreatedCallCount")
	t.V1Alpha1BuildRunCreatedCallCount++
}

func (t *testConsumer) V1Alpha1BuildRunFailed(ctx context.Context, buildRun *buildv1alpha1.BuildRun) {
	fmt.Println("Incrementing V1Alpha1BuildRunFailedCallCount")
	t.V1Alpha1BuildRunFailedCallCount++
}

func (t *testConsumer) V1Alpha1BuildRunRunning(ctx context.Context, buildRun *buildv1alpha1.BuildRun) {
	fmt.Println("Incrementing V1Alpha1BuildRunRunningCallCount")
	t.V1Alpha1BuildRunRunningCallCount++
}

func (t *testConsumer) V1Alpha1BuildRunSucceeded(ctx context.Context, buildRun *buildv1alpha1.BuildRun) {
	fmt.Println("Incrementing V1Alpha1BuildRunSucceededCallCount")
	t.V1Alpha1BuildRunSucceededCallCount++
}

var _ = Describe("monitoring", func() {

	consumer := newTestConsumer()
	monitoring.RegisterConsumer(consumer)

	It("should invoke V1Alpha1BuildRunCreated for all consumers", func() {
		monitoring.OnV1Alpha1BuildRunCreated(context.TODO(), nil)
		Expect(consumer.V1Alpha1BuildRunCreatedCallCount).To(Equal(1))
	})

	It("should invoke V1Alpha1BuildRunFailed for all consumers", func() {
		monitoring.OnV1Alpha1BuildRunFailed(context.TODO(), nil)
		Expect(consumer.V1Alpha1BuildRunFailedCallCount).To(Equal(1))
	})

	It("should invoke V1Alpha1BuildRunRunning for all consumers", func() {
		monitoring.OnV1Alpha1BuildRunRunning(context.TODO(), nil)
		Expect(consumer.V1Alpha1BuildRunRunningCallCount).To(Equal(1))
	})

	It("should invoke V1Alpha1BuildRunSucceeded for all consumers", func() {
		monitoring.OnV1Alpha1BuildRunSucceeded(context.TODO(), nil)
		Expect(consumer.V1Alpha1BuildRunSucceededCallCount).To(Equal(1))
	})

})
