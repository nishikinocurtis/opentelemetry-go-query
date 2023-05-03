// Qiutong Men 2023 April 26.

package otel // import "go.opentelemetry.io/otel"
import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/internal/global"
)

// GetTraceAttributeFilter returns the global TraceAttributeFilter.
// Currently, it will return the only implementation
// (the default) wrapped mapTraceAttributeFilter.
func GetTraceAttributeFilter() attribute.TraceAttributeFilter {
	return global.TraceAttributeFilter()
}

// SetTraceAttributeFilter sets the global TraceAttributeFilter.
// TODO: there's no effect of this function.
func SetTraceAttributeFilter(f attribute.TraceAttributeFilter) {
	global.SetTraceAttributeFilter(f)
}

func WithAttributeFilter() global.FilterConfigFlag {
	return global.AttributeFilter
}

func WithAttributeNotMatchFullTraceFilter() global.FilterConfigFlag {
	return global.AttributeNotMatchFullTraceFilter
}

func WithStructuralTraceFilter() global.FilterConfigFlag {
	return global.StructuralTraceFilter
}

func SetAttributeFilterConfig(flags ...global.FilterConfigFlag) {
	var flag global.FilterConfigFlag = 0
	for _, f := range flags {
		flag |= f
	}
	global.SetFilterConfigFlags(flag)
}
