package httputil

import (
	"io"
	"net/http"
)

func Put(url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", bodyType)
	return http.DefaultClient.Do(req)
}
