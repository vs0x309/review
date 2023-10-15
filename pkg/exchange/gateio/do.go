package gateio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func (a *API) doPublicGET(ctx context.Context, endpoint string, payload url.Values, result any) error {
	a.mu.Lock()
	defer func() {
		go func() {
			time.Sleep(doPause)
			a.mu.Unlock()
		}()
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+endpoint, nil)
	if err != nil {
		return err
	}

	req.URL.RawQuery = payload.Encode()

	req.Header.Add("Accept", "application/json")

	return a.do(req, result)
}

func (a *API) do(req *http.Request, result any) error {
	rsp, err := a.cli.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	if Debug {
		log.Printf("%s %s %d\n%s", req.Method, req.URL, rsp.StatusCode, string(body))
	}

	if rsp.StatusCode != 200 {
		checkErr := struct {
			Label   string `json:"label"`
			Message string `json:"message"`
		}{}

		if err = json.Unmarshal(body, &checkErr); err == nil {
			if len(checkErr.Label) > 0 {
				return fmt.Errorf("%s %s %d [%s]", req.Method, req.URL, rsp.StatusCode, checkErr.Label)
			}
		}

		return fmt.Errorf("%s %s %d", req.Method, req.URL, rsp.StatusCode)
	}

	if result == nil {
		return nil
	}

	return json.Unmarshal(body, result)
}
