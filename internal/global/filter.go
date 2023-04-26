// Qiutong Men 2023 April. 25

package global // import "go.opentelemetry.io/otel/internal/global"
import (
	"go.opentelemetry.io/otel/attribute"
	"sync"
)

// traceAttributeFilter is a default (currently unique) TraceAttributeFilter
// that maintains a Read-Write lock to protect the filter from concurrent
// access, while maintaining efficiency for read-only access (which is much
// more frequent in production).
// This object is written to follow OpenTelemetry's design pattern for global
// states.
type traceAttributeFilter struct {
	rwx sync.RWMutex
	taf attribute.TraceAttributeFilter
}

var _ attribute.TraceAttributeFilter = (*traceAttributeFilter)(nil)

func newTraceAttributeFilter() *traceAttributeFilter {
	return &traceAttributeFilter{
		taf: attribute.NewMapTraceAttributeFilter(),
	}
}

func (t *traceAttributeFilter) AddRangeMatch(key attribute.Key, lb attribute.Value, ub attribute.Value) {
	t.rwx.Lock()
	defer t.rwx.Unlock()
	t.taf.AddRangeMatch(key, lb, ub)
}

func (t *traceAttributeFilter) AddEqualityMatch(key attribute.Key, value attribute.Value) {
	t.rwx.Lock()
	defer t.rwx.Unlock()
	t.taf.AddEqualityMatch(key, value)
}

func (t *traceAttributeFilter) AddKeyMatch(key attribute.Key) {
	t.rwx.Lock()
	defer t.rwx.Unlock()
	t.taf.AddKeyMatch(key)
}

func (t *traceAttributeFilter) RemoveMatch(key attribute.Key) {
	t.rwx.Lock()
	defer t.rwx.Unlock()
	t.taf.RemoveMatch(key)
}

func (t *traceAttributeFilter) Match(key attribute.Key, value attribute.Value) bool {
	t.rwx.RLock()
	defer t.rwx.RUnlock()
	return t.taf.Match(key, value)
}
