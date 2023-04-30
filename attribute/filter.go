// Qiutong Men 2023 April. 25

package attribute // import "go.opentelemetry.io/otel/attribute"
import "net/http"

type MatchValueFlag int

const (
	NoValue MatchValueFlag = iota
	EQUALITY
	RANGE
)

type TraceAttributeFilter interface {
	AddRangeMatch(key Key, lb Value, ub Value)
	AddEqualityMatch(key Key, value Value)
	AddKeyMatch(key Key)
	RemoveMatch(key Key)
	Match(key Key, value Value) bool
	BatchMatch(attrs []KeyValue, callback func(KeyValue) error)
	BatchNotMatch(attrs []KeyValue, callback func() error)
	Clear()
	HandleRequest(req *http.Request) error
}

type TraceAttributeValueMatch struct {
	mvf MatchValueFlag
	lb  Value // range lower bound, inclusive
	ub  Value // range upper bound, exclusive

}

type mapTraceAttributeFilter struct {
	// matches is a map from attribute key to a value specifier
	matches map[Key]TraceAttributeValueMatch
}

// AddRangeMatch appends a legal range match to the filter
func (f *mapTraceAttributeFilter) AddRangeMatch(key Key, lb Value, ub Value) {
	// check if key exists
	if _, ok := f.matches[key]; ok {
		// logging duplicate key
		println("Duplicate key: key already exists in the filter")
		// overwrite the existing key
	}

	// do type checking: lb and ub must be of the same type, and must be comparable
	if lb.Type() != ub.Type() {
		// logging illegal type
		println("Illegal type: lower bound and upper bound must be of the same type")
		return
	}

	if lb.Type() != INT64 && lb.Type() != FLOAT64 {
		// logging illegal type
		println("Illegal type: lower bound and upper bound must be of type INT64 or FLOAT64")
		return
	}

	// check if lb <= ub
	if lb.Type() == INT64 {
		if lb.AsInt64() > ub.AsInt64() {
			// reverse the order of lb and ub
			f.matches[key] = TraceAttributeValueMatch{RANGE, ub, lb}
		}
	} else {
		if lb.AsFloat64() > ub.AsFloat64() {
			// reverse the order of lb and ub
			f.matches[key] = TraceAttributeValueMatch{RANGE, ub, lb}
		}
	}

	f.matches[key] = TraceAttributeValueMatch{RANGE, lb, ub}
}

// AddEqualityMatch appends a legal equality match to the filter
func (f *mapTraceAttributeFilter) AddEqualityMatch(key Key, value Value) {
	// check if key exists
	if _, ok := f.matches[key]; ok {
		// logging duplicate key
		println("Duplicate key: key already exists in the filter")
		// overwrite the existing key
	}

	// do type checking: value must be of type BOOL, INT64, or STRING
	if value.Type() != BOOL && value.Type() != INT64 && value.Type() != STRING {
		// logging illegal type
		println("Illegal type: value must be of type BOOL, INT64, or STRING")
		return
	}

	f.matches[key] = TraceAttributeValueMatch{EQUALITY, value, value}
}

// AddKeyMatch appends a legal key match to the filter, without value
func (f *mapTraceAttributeFilter) AddKeyMatch(key Key) {
	// check if key exists
	if _, ok := f.matches[key]; ok {
		// logging duplicate key
		println("Duplicate key: key already exists in the filter")
		// overwrite the existing key
	}

	f.matches[key] = TraceAttributeValueMatch{NoValue,
		InvalidValue(), InvalidValue()}
}

// RemoveMatch removes a match from the filter
func (f *mapTraceAttributeFilter) RemoveMatch(key Key) {
	delete(f.matches, key)
}

// Match returns true if the key-value pair matches the filter
func (f *mapTraceAttributeFilter) Match(key Key, value Value) bool {
	if match, ok := f.matches[key]; ok {
		switch match.mvf {
		case NoValue:
			return true
		case EQUALITY:
			return match.lb == value
		case RANGE:
			if value.Type() == INT64 && match.lb.Type() == INT64 {
				return match.lb.AsInt64() <= value.AsInt64() &&
					value.AsInt64() <= match.ub.AsInt64()
			} else if value.Type() == FLOAT64 && match.lb.Type() == FLOAT64 {
				return match.lb.AsFloat64() <= value.AsFloat64() &&
					value.AsFloat64() <= match.ub.AsFloat64()
			}
		}
	}
	return false
}

// Clear clears all filters
func (f *mapTraceAttributeFilter) Clear() {
	// directly assign a new map to the map, the old map will be garbage collected
	f.matches = make(map[Key]TraceAttributeValueMatch)
}

// HandleRequest execute the filter operations and returns an error if the request is unsupported
// No effect as base type
func (f *mapTraceAttributeFilter) HandleRequest(req *http.Request) error {
	return nil
}

// BatchMatch execute callback for all attributes matching one filter
func (f *mapTraceAttributeFilter) BatchMatch(attrs []KeyValue, callback func(KeyValue) error) {
	for _, attr := range attrs {
		if f.Match(attr.Key, attr.Value) {
			err := callback(attr)
			if err != nil {
				return
			}
		}
	}
}

// BatchNotMatch execute callback if any existing filter is not matched, callback is executed only once
func (f *mapTraceAttributeFilter) BatchNotMatch(attrs []KeyValue, callback func() error) {
	matchTarget := len(f.matches)
	for _, attr := range attrs {
		if _, ok := f.matches[attr.Key]; !ok {
			return
		}
		if f.Match(attr.Key, attr.Value) {
			matchTarget--
		}
	}
	if matchTarget != 0 {
		if err := callback(); err != nil {
			println("Error in callback function in BatchNotMatch: ", err.Error())
			return
		}
	}
}

func NewMapTraceAttributeFilter() TraceAttributeFilter {
	return &mapTraceAttributeFilter{
		matches: make(map[Key]TraceAttributeValueMatch),
	}
}
