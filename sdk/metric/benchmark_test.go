// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metric // import "go.opentelemetry.io/otel/sdk/metric"

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

var viewBenchmarks = []struct {
	Name  string
	Views []View
}{
	{"NoView", []View{}},
	{
		"DropView",
		[]View{NewView(
			Instrument{Name: "*"},
			Stream{Aggregation: aggregation.Drop{}},
		)},
	},
	{
		"AttrFilterView",
		[]View{NewView(
			Instrument{Name: "*"},
			Stream{AttributeFilter: func(kv attribute.KeyValue) bool {
				return kv.Key == attribute.Key("K")
			}},
		)},
	},
}

func BenchmarkSyncMeasure(b *testing.B) {
	for _, bc := range viewBenchmarks {
		b.Run(bc.Name, benchSyncViews(bc.Views...))
	}
}

func benchSyncViews(views ...View) func(*testing.B) {
	ctx := context.Background()
	rdr := NewManualReader()
	provider := NewMeterProvider(WithReader(rdr), WithView(views...))
	meter := provider.Meter("benchSyncViews")
	return func(b *testing.B) {
		iCtr, err := meter.Int64Counter("int64-counter")
		assert.NoError(b, err)
		b.Run("Int64Counter", benchMeasAttrs(func() measF {
			return func(attr []attribute.KeyValue) func() {
				return func() { iCtr.Add(ctx, 1, attr...) }
			}
		}()))

		fCtr, err := meter.Float64Counter("float64-counter")
		assert.NoError(b, err)
		b.Run("Float64Counter", benchMeasAttrs(func() measF {
			return func(attr []attribute.KeyValue) func() {
				return func() { fCtr.Add(ctx, 1, attr...) }
			}
		}()))

		iUDCtr, err := meter.Int64UpDownCounter("int64-up-down-counter")
		assert.NoError(b, err)
		b.Run("Int64UpDownCounter", benchMeasAttrs(func() measF {
			return func(attr []attribute.KeyValue) func() {
				return func() { iUDCtr.Add(ctx, 1, attr...) }
			}
		}()))

		fUDCtr, err := meter.Float64UpDownCounter("float64-up-down-counter")
		assert.NoError(b, err)
		b.Run("Float64UpDownCounter", benchMeasAttrs(func() measF {
			return func(attr []attribute.KeyValue) func() {
				return func() { fUDCtr.Add(ctx, 1, attr...) }
			}
		}()))

		iHist, err := meter.Int64Histogram("int64-histogram")
		assert.NoError(b, err)
		b.Run("Int64Histogram", benchMeasAttrs(func() measF {
			return func(attr []attribute.KeyValue) func() {
				return func() { iHist.Record(ctx, 1, attr...) }
			}
		}()))

		fHist, err := meter.Float64Histogram("float64-histogram")
		assert.NoError(b, err)
		b.Run("Float64Histogram", benchMeasAttrs(func() measF {
			return func(attr []attribute.KeyValue) func() {
				return func() { fHist.Record(ctx, 1, attr...) }
			}
		}()))
	}
}

type measF func(attr []attribute.KeyValue) func()

func benchMeasAttrs(meas measF) func(*testing.B) {
	return func(b *testing.B) {
		b.Run("Attributes/0", func(b *testing.B) {
			f := meas(nil)
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				f()
			}
		})
		b.Run("Attributes/1", func(b *testing.B) {
			attrs := []attribute.KeyValue{attribute.Bool("K", true)}
			f := meas(attrs)
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				f()
			}
		})
		b.Run("Attributes/10", func(b *testing.B) {
			n := 10
			attrs := make([]attribute.KeyValue, 0)
			attrs = append(attrs, attribute.Bool("K", true))
			for i := 2; i < n; i++ {
				attrs = append(attrs, attribute.Int(strconv.Itoa(i), i))
			}
			f := meas(attrs)
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				f()
			}
		})
	}
}

func BenchmarkCollect(b *testing.B) {
	for _, bc := range viewBenchmarks {
		b.Run(bc.Name, benchCollectViews(bc.Views...))
	}
}

func benchCollectViews(views ...View) func(*testing.B) {
	setup := func(name string) (metric.Meter, Reader) {
		r := NewManualReader()
		mp := NewMeterProvider(WithReader(r), WithView(views...))
		return mp.Meter(name), r
	}
	ctx := context.Background()
	return func(b *testing.B) {
		b.Run("Int64Counter/1", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Int64Counter")
			i, err := m.Int64Counter("int64-counter")
			assert.NoError(b, err)
			i.Add(ctx, 1, attr...)
			return r
		}))
		b.Run("Int64Counter/10", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Int64Counter")
			i, err := m.Int64Counter("int64-counter")
			assert.NoError(b, err)
			for n := 0; n < 10; n++ {
				i.Add(ctx, 1, attr...)
			}
			return r
		}))

		b.Run("Float64Counter/1", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Float64Counter")
			i, err := m.Float64Counter("float64-counter")
			assert.NoError(b, err)
			i.Add(ctx, 1, attr...)
			return r
		}))
		b.Run("Float64Counter/10", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Float64Counter")
			i, err := m.Float64Counter("float64-counter")
			assert.NoError(b, err)
			for n := 0; n < 10; n++ {
				i.Add(ctx, 1, attr...)
			}
			return r
		}))

		b.Run("Int64UpDownCounter/1", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Int64UpDownCounter")
			i, err := m.Int64UpDownCounter("int64-up-down-counter")
			assert.NoError(b, err)
			i.Add(ctx, 1, attr...)
			return r
		}))
		b.Run("Int64UpDownCounter/10", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Int64UpDownCounter")
			i, err := m.Int64UpDownCounter("int64-up-down-counter")
			assert.NoError(b, err)
			for n := 0; n < 10; n++ {
				i.Add(ctx, 1, attr...)
			}
			return r
		}))

		b.Run("Float64UpDownCounter/1", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Float64UpDownCounter")
			i, err := m.Float64UpDownCounter("float64-up-down-counter")
			assert.NoError(b, err)
			i.Add(ctx, 1, attr...)
			return r
		}))
		b.Run("Float64UpDownCounter/10", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Float64UpDownCounter")
			i, err := m.Float64UpDownCounter("float64-up-down-counter")
			assert.NoError(b, err)
			for n := 0; n < 10; n++ {
				i.Add(ctx, 1, attr...)
			}
			return r
		}))

		b.Run("Int64Histogram/1", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Int64Histogram")
			i, err := m.Int64Histogram("int64-histogram")
			assert.NoError(b, err)
			i.Record(ctx, 1, attr...)
			return r
		}))
		b.Run("Int64Histogram/10", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Int64Histogram")
			i, err := m.Int64Histogram("int64-histogram")
			assert.NoError(b, err)
			for n := 0; n < 10; n++ {
				i.Record(ctx, 1, attr...)
			}
			return r
		}))

		b.Run("Float64Histogram/1", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Float64Histogram")
			i, err := m.Float64Histogram("float64-histogram")
			assert.NoError(b, err)
			i.Record(ctx, 1, attr...)
			return r
		}))
		b.Run("Float64Histogram/10", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Float64Histogram")
			i, err := m.Float64Histogram("float64-histogram")
			assert.NoError(b, err)
			for n := 0; n < 10; n++ {
				i.Record(ctx, 1, attr...)
			}
			return r
		}))

		b.Run("Int64ObservableCounter", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Int64ObservableCounter")
			_, err := m.Int64ObservableCounter(
				"int64-observable-counter",
				instrument.WithInt64Callback(int64Cback(attr)),
			)
			assert.NoError(b, err)
			return r
		}))

		b.Run("Float64ObservableCounter", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Float64ObservableCounter")
			_, err := m.Float64ObservableCounter(
				"float64-observable-counter",
				instrument.WithFloat64Callback(float64Cback(attr)),
			)
			assert.NoError(b, err)
			return r
		}))

		b.Run("Int64ObservableUpDownCounter", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Int64ObservableUpDownCounter")
			_, err := m.Int64ObservableUpDownCounter(
				"int64-observable-up-down-counter",
				instrument.WithInt64Callback(int64Cback(attr)),
			)
			assert.NoError(b, err)
			return r
		}))

		b.Run("Float64ObservableUpDownCounter", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Float64ObservableUpDownCounter")
			_, err := m.Float64ObservableUpDownCounter(
				"float64-observable-up-down-counter",
				instrument.WithFloat64Callback(float64Cback(attr)),
			)
			assert.NoError(b, err)
			return r
		}))

		b.Run("Int64ObservableGauge", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Int64ObservableGauge")
			_, err := m.Int64ObservableGauge(
				"int64-observable-gauge",
				instrument.WithInt64Callback(int64Cback(attr)),
			)
			assert.NoError(b, err)
			return r
		}))

		b.Run("Float64ObservableGauge", benchCollectAttrs(func(attr []attribute.KeyValue) Reader {
			m, r := setup("benchCollectViews/Float64ObservableGauge")
			_, err := m.Float64ObservableGauge(
				"float64-observable-gauge",
				instrument.WithFloat64Callback(float64Cback(attr)),
			)
			assert.NoError(b, err)
			return r
		}))
	}
}

func int64Cback(attr []attribute.KeyValue) instrument.Int64Callback {
	return func(_ context.Context, o instrument.Int64Observer) error {
		o.Observe(1, attr...)
		return nil
	}
}

func float64Cback(attr []attribute.KeyValue) instrument.Float64Callback {
	return func(_ context.Context, o instrument.Float64Observer) error {
		o.Observe(1, attr...)
		return nil
	}
}

func benchCollectAttrs(setup func([]attribute.KeyValue) Reader) func(*testing.B) {
	ctx := context.Background()
	out := new(metricdata.ResourceMetrics)
	run := func(reader Reader) func(b *testing.B) {
		return func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				_ = reader.Collect(ctx, out)
			}
		}
	}
	return func(b *testing.B) {
		b.Run("Attributes/0", run(setup(nil)))

		attrs := []attribute.KeyValue{attribute.Bool("K", true)}
		b.Run("Attributes/1", run(setup(attrs)))

		for i := 2; i < 10; i++ {
			attrs = append(attrs, attribute.Int(strconv.Itoa(i), i))
		}
		b.Run("Attributes/10", run(setup(attrs)))
	}
}
