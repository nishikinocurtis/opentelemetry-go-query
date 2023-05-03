// Qiutong Men 2023 April. 25

package global // import "go.opentelemetry.io/otel/internal/global"
import (
	"encoding/json"
	"errors"
	"go.opentelemetry.io/otel/attribute"
	"net/http"
	"sync"
)

type updateFilterRequests struct {
	Filters []struct {
		Key    attribute.Key `json:"key"`
		Type   string        `json:"type"`
		Values []any         `json:"values"`
	} `json:"filters"`
}

type removeFilterRequests struct {
	Filters []attribute.Key `json:"filters"`
}

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

func newTraceEventFilter() *traceAttributeFilter {
	return &traceAttributeFilter{
		taf: attribute.NewMapTraceAttributeFilter(),
	}
}

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

func (t *traceAttributeFilter) BatchMatch(attrs []attribute.KeyValue, callback func(attribute.KeyValue) error) {
	t.rwx.RLock()
	defer t.rwx.RUnlock()
	t.taf.BatchMatch(attrs, callback)
}

func (t *traceAttributeFilter) BatchNotMatch(attrs []attribute.KeyValue, callback func() error) {
	t.rwx.RLock()
	defer t.rwx.RUnlock()
	t.taf.BatchNotMatch(attrs, callback)
}

func (t *traceAttributeFilter) updateFilter(ufrs updateFilterRequests) error {
	t.rwx.Lock()
	defer t.rwx.Unlock()
	for _, filter := range ufrs.Filters {
		if len(filter.Values) == 0 {
			t.taf.AddKeyMatch(filter.Key)
		} else {
			switch filter.Type {
			case "string": // for string type filter, only the first value is used
				t.taf.AddEqualityMatch(filter.Key, attribute.StringValue(filter.Values[0].(string)))
			case "bool": // for bool type filter, only the first value is used
				t.taf.AddEqualityMatch(filter.Key, attribute.BoolValue(filter.Values[0].(bool)))
			case "int64":
				// for int64 type filter, depending on the number of values, either equality or range match is used
				if len(filter.Values) == 1 {
					t.taf.AddEqualityMatch(filter.Key,
						attribute.Int64Value(int64(filter.Values[0].(float64))))
				} else {
					t.taf.AddRangeMatch(filter.Key,
						attribute.Int64Value(int64(filter.Values[0].(float64))),
						attribute.Int64Value(int64(filter.Values[1].(float64))))
				}
			case "float64":
				// for float64 type filter, depending on the number of values, either equality or range match is used
				if len(filter.Values) == 1 {
					t.taf.AddEqualityMatch(filter.Key,
						attribute.Float64Value(filter.Values[0].(float64)))
				} else {
					t.taf.AddRangeMatch(filter.Key,
						attribute.Float64Value(filter.Values[0].(float64)),
						attribute.Float64Value(filter.Values[1].(float64)))
				}
			default:
				return errors.New("Unsupported type: " + filter.Type)
			}
		}
	}
	return nil
}

func (t *traceAttributeFilter) removeFilter(rfrs removeFilterRequests) error {
	t.rwx.Lock()
	defer t.rwx.Unlock()
	for _, filter := range rfrs.Filters {
		t.taf.RemoveMatch(filter)
	}
	return nil
}

func (t *traceAttributeFilter) Clear() {
	t.rwx.Lock()
	defer t.rwx.Unlock()
	t.taf.Clear()
}

func (t *traceAttributeFilter) HandleRequest(r *http.Request) error {
	reqOp := r.URL.Query().Get("op")
	println("opCode: " + reqOp)
	switch reqOp {
	case "update":
		println("Start updating filter")
		var ufrs updateFilterRequests
		if err := json.NewDecoder(r.Body).Decode(&ufrs); err != nil {
			return err
		}
		if err := t.updateFilter(ufrs); err != nil {
			return err
		}
		return nil
	case "remove":
		var rfrs removeFilterRequests
		if err := json.NewDecoder(r.Body).Decode(&rfrs); err != nil {
			return err
		}
		if err := t.removeFilter(rfrs); err != nil {
			return err
		}
		return nil
	case "clear":
		t.Clear()
		return nil
	default:
		return errors.New("Unsupported opCode: " + reqOp)
	}
}
