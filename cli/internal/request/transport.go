package request

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
)

type verboseTransport struct {
	Base http.RoundTripper
}

func (t *verboseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stdout, "%s\n", reqDump)

	resp, err := t.Base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stdout, "%s\n", respDump)

	return resp, nil
}
