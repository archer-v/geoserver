package geoserver

type GsError struct {
	err  string
	dump string
}

func (e GsError) Error() string {
	return e.err
}

func (e GsError) Dump() string {
	return e.dump
}

var statusErrorMapping = map[int]GsError{
	statusNotAllowed:    {err: "Method Not Allowed"},
	statusNotFound:      {err: "Not Found"},
	statusUnauthorized:  {err: "Unauthorized"},
	statusInternalError: {err: "Internal Server Error"},
	statusForbidden:     {err: "Forbidden"},
}
