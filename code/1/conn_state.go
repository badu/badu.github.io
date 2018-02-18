package http

func (c ConnState) String() string {
	return stateName[c]
}
