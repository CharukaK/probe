package request

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
)

const (
	colorReset = "\033[0m"
	colorReq   = "\033[36m"
	colorResp  = "\033[32m"
)

type verboseTransport struct {
	Base http.RoundTripper
}

func (t *verboseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}
	writePrefixed(os.Stdout, reqDump, ">", colorReq)

	resp, err := t.Base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	respDump, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return nil, err
	}

	writePrefixed(os.Stdout, respDump, "<", colorResp)

	return resp, nil
}

func writePrefixed(w io.Writer, dump []byte, prefix, color string) {
	scanner := bufio.NewScanner(bytes.NewReader(dump))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		fmt.Fprintf(w, "%s%s %s%s\n", color, prefix, scanner.Text(), colorReset)
	}
	fmt.Fprintln(w) // blank separator line between this dump and the next
}
