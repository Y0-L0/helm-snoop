package parser

func includeFn(...interface{}) interface{} { panic("not implemented") }
func tplFn(...interface{}) interface{}     { panic("not implemented") }
func noopFn(...interface{}) interface{}    { return nil }

// getFn appends a literal key to an absolute .Values path.
func getFn(args ...interface{}) interface{} {
	if len(args) != 2 {
		panic("not implemented / invalid template")
	}
	base, ok := args[0].(AbsPath)
	if !ok {
		panic("not implemented")
	}
	key, ok := args[1].(LiteralSet)
	if !ok || len(key.Values) != 1 {
		panic("not implemented")
	}
	segs := append(append([]string{}, base.Segs...), key.Values[0])
	return AbsPath{Segs: segs}
}

// indexFn appends one or more literal keys to an absolute .Values path.
func indexFn(args ...interface{}) interface{} {
	if len(args) < 2 {
		return nil
	}
	base, ok := args[0].(AbsPath)
	if !ok {
		panic("not implemented")
	}
	segs := append([]string{}, base.Segs...)
	for _, a := range args[1:] {
		lit, ok := a.(LiteralSet)
		if !ok || len(lit.Values) != 1 {
			panic("not implemented")
		}
		segs = append(segs, lit.Values[0])
	}
	return AbsPath{Segs: segs}
}

// Analysis-aware helpers used by our evaluator.
func defaultFn(args ...interface{}) interface{} {
	if len(args) == 0 {
		return nil
	}
	return args[len(args)-1]
}

func passthrough1Fn(args ...interface{}) interface{} {
	if len(args) == 0 {
		return nil
	}
	return args[0]
}
