package request

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/charukak/probe/cli/internal/parser"
)

type Opts struct {
	Verbose bool
}

func BuildAndExec(sf *parser.SourceFile, opts *Opts) error {
	client := &http.Client{}

	if opts.Verbose {
		client.Transport = &verboseTransport{Base: http.DefaultTransport}
	}

	for _, r := range sf.Children {
		hr, err := http.NewRequest(
			r.RequestLine.Method,
			r.RequestLine.Target,
			bytes.NewReader(r.Body.Octets),
		)

		if err != nil {
			return err
		}

		for _, h := range r.HeaderLines {
			hr.Header.Add(h.Key, h.Value)
		}

		resp, err := client.Do(hr)
		if err != nil {
			return err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", body)
	}

	return nil
}
