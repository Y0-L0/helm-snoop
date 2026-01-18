package parser

func (s *Unittest) TestCalcPosition_SingleLine() {
	line, col := CalcPosition("hello", 2)
	s.Equal(1, line)
	s.Equal(3, col)
}

func (s *Unittest) TestCalcPosition_MultiLine() {
	src := "line1\nline2\nline3"
	line, col := CalcPosition(src, 7) // 'i' in line2
	s.Equal(2, line)
	s.Equal(2, col)
}

func (s *Unittest) TestCalcPosition_StartOfFile() {
	line, col := CalcPosition("hello", 0)
	s.Equal(1, line)
	s.Equal(1, col)
}
