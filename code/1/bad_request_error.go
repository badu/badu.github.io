package http

func (e badRequestError) Error() string { return "Bad Request: " + string(e) }
