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
