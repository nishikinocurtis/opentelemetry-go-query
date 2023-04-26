// Qiutong Men 2023 April. 25

package attribute // import "go.opentelemetry.io/otel/attribute"

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
			// logging illegal type
			println("Illegal type: lower bound must be less than or equal to upper bound")
			return
		}
	} else {
		if lb.AsFloat64() > ub.AsFloat64() {
			// logging illegal type
			println("Illegal type: lower bound must be less than or equal to upper bound")
			return
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
	return true
}

func NewMapTraceAttributeFilter() TraceAttributeFilter {
	return &mapTraceAttributeFilter{
		matches: make(map[Key]TraceAttributeValueMatch),
	}
}
