package engine

type impl struct {
	queries []string
}

func New() *impl {
	return &impl{}
}
