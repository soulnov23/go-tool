package pipeline

type stage struct {
	source      any
	sink        any
	preFilters  []Filter
	postFilters []Filter
}

func newStage() *stage {
	return &stage{}
}

func (s *stage) process() error {
	return nil
}
