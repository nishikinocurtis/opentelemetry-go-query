package trace

type FilterConfigFlag int

const (
	AttributeNotMatchFullTraceFilter FilterConfigFlag = 1 << iota
	AttributeFilter                                   = 1 << 1
	StructuralTraceFilter                             = 1 << 2
)

func (f FilterConfigFlag) WithFullTraceFilter() FilterConfigFlag {
	return f | AttributeNotMatchFullTraceFilter
}

func (f FilterConfigFlag) WithAttributeFilter() FilterConfigFlag {
	return f | AttributeFilter
}

func (f FilterConfigFlag) WithStructuralTraceFilter() FilterConfigFlag {
	return f | StructuralTraceFilter
}
