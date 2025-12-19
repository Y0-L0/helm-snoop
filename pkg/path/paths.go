package path

type Paths []*Path

func (ps Paths) Len() int      { return len(ps) }
func (ps Paths) Swap(i, j int) { ps[i], ps[j] = ps[j], ps[i] }
func (ps Paths) Less(i, j int) bool {
	pi, pj := ps[i], ps[j]
	if pi == nil {
		return false
	}
	if pj == nil {
		return true
	}
	return ps[i].Compare(*ps[j]) < 0
}

func (ps *Paths) Append(paths ...*Path) {
	*ps = append(*ps, paths...)
}
