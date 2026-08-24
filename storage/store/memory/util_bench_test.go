package memory

import (
	"testing"

	"hanzo.ai/status/config/endpoint"
	"hanzo.ai/status/storage"
	"hanzo.ai/status/storage/store/common/paging"
)

func BenchmarkShallowCopyEndpointStatus(b *testing.B) {
	ep := &testEndpoint
	status := endpoint.NewStatus(ep.Group, ep.Name)
	for range storage.DefaultMaximumNumberOfResults {
		AddResult(status, &testSuccessfulResult, storage.DefaultMaximumNumberOfResults, storage.DefaultMaximumNumberOfEvents)
	}
	for b.Loop() {
		ShallowCopyEndpointStatus(status, paging.NewEndpointStatusParams().WithResults(1, 20))
	}
	b.ReportAllocs()
}
