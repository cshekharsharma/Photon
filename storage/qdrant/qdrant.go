package qdrant

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type restClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client

	retryMaxAttempts int
	retryBaseDelay   time.Duration
	retryMaxDelay    time.Duration
	retryJitter      float64
}

func NewRESTClient(cfg *ConnectionConfig) (*restClient, error) {
	if cfg == nil {
		return nil, errors.New("qdrant: nil config")
	}
	base := strings.TrimSpace(cfg.BaseURL)
	if base == "" {
		return nil, errors.New("qdrant: BaseURL is required")
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, errors.New("qdrant: invalid BaseURL")
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	dial := cfg.DialTimeout
	if dial <= 0 {
		dial = 5 * time.Second
	}
	idle := cfg.IdleConnTimeout
	if idle <= 0 {
		idle = 90 * time.Second
	}

	maxAttempts := cfg.RetryMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	baseDelay := cfg.RetryBaseDelay
	if baseDelay <= 0 {
		baseDelay = 150 * time.Millisecond
	}
	maxDelay := cfg.RetryMaxDelay
	if maxDelay <= 0 {
		maxDelay = 2 * time.Second
	}
	jitter := cfg.RetryJitter
	if jitter < 0 || jitter > 1 {
		jitter = 0.2
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   dial,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   50,
		IdleConnTimeout:       idle,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
	}

	return &restClient{
		baseURL: strings.TrimRight(base, "/"),
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   timeout,
		},
		retryMaxAttempts: maxAttempts,
		retryBaseDelay:   baseDelay,
		retryMaxDelay:    maxDelay,
		retryJitter:      jitter,
	}, nil
}

func (c *restClient) Health(ctx context.Context) error {
	// Qdrant commonly exposes /healthz
	var out any
	return c.doJSON(ctx, http.MethodGet, "/healthz", nil, &out)
}

func (c *restClient) UpsertPoints(ctx context.Context, req UpsertPointsRequest) error {
	if req.Collection == "" || len(req.Points) == 0 {
		return errors.New("qdrant: UpsertPoints invalid request")
	}
	body := map[string]any{"points": req.Points}
	path := "/collections/" + url.PathEscape(req.Collection) + "/points"
	if req.Wait {
		path += "?wait=true"
	}
	var out any
	return c.doJSON(ctx, http.MethodPut, path, body, &out)
}

func (c *restClient) DeletePoints(ctx context.Context, req DeletePointsRequest) error {
	if req.Collection == "" || len(req.IDs) == 0 {
		return errors.New("qdrant: DeletePoints invalid request")
	}
	body := map[string]any{"points": map[string]any{"ids": req.IDs}}
	path := "/collections/" + url.PathEscape(req.Collection) + "/points/delete"
	if req.Wait {
		path += "?wait=true"
	}
	var out any
	return c.doJSON(ctx, http.MethodPost, path, body, &out)
}

func (c *restClient) Search(ctx context.Context, req SearchRequest) ([]ScoredPoint, error) {
	if req.Collection == "" || len(req.Vector) == 0 || req.Limit <= 0 {
		return nil, errors.New("qdrant: Search invalid request")
	}

	body := map[string]any{
		"vector":       req.Vector,
		"limit":        req.Limit,
		"offset":       req.Offset,
		"with_payload": req.WithPayload,
		"with_vector":  req.WithVector,
	}
	if req.ScoreThreshold != nil {
		body["score_threshold"] = *req.ScoreThreshold
	}
	if len(req.Filter) > 0 {
		var f any
		if err := json.Unmarshal(req.Filter, &f); err != nil {
			return nil, errors.New("qdrant: invalid filter JSON")
		}
		body["filter"] = f
	}

	var resp struct {
		Result []ScoredPoint `json:"result"`
		Status string        `json:"status"`
	}
	path := "/collections/" + url.PathEscape(req.Collection) + "/points/search"
	if err := c.doJSON(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

func (c *restClient) CreateCollection(ctx context.Context, req CreateCollectionRequest) error {
	if strings.TrimSpace(req.Collection) == "" {
		return errors.New("qdrant: CreateCollection invalid collection")
	}
	body, err := req.Config.ToJSON()
	if err != nil {
		return err
	}

	var out any
	path := "/collections/" + url.PathEscape(req.Collection)
	return c.doJSON(ctx, http.MethodPut, path, body, &out)
}

func (c *restClient) DeleteCollection(ctx context.Context, collection string) error {
	if strings.TrimSpace(collection) == "" {
		return errors.New("qdrant: DeleteCollection invalid collection")
	}
	var out any
	path := "/collections/" + url.PathEscape(collection)
	return c.doJSON(ctx, http.MethodDelete, path, nil, &out)
}

func (c *restClient) GetCollectionInfo(ctx context.Context, collection string) (CollectionInfo, error) {
	if strings.TrimSpace(collection) == "" {
		return CollectionInfo{}, errors.New("qdrant: GetCollectionInfo invalid collection")
	}
	var resp struct {
		Result struct {
			Status string `json:"status"`
		} `json:"result"`
		Status string `json:"status"`
	}
	path := "/collections/" + url.PathEscape(collection)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return CollectionInfo{}, err
	}
	return CollectionInfo{Status: resp.Result.Status}, nil
}

// ---------------- internal: hardened request executor ----------------

func (c *restClient) doJSON(ctx context.Context, method, path string, in any, out any) error {
	var payload []byte
	var err error
	if in != nil {
		payload, err = json.Marshal(in)
		if err != nil {
			return err
		}
	}

	var lastErr error
	for attempt := 1; attempt <= c.retryMaxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		if in != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if c.apiKey != "" {
			// Qdrant supports "api-key"
			req.Header.Set("api-key", c.apiKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if !isRetryableNetErr(err) || attempt == c.retryMaxAttempts {
				break
			}
			if err := sleepCtx(ctx, c.backoff(attempt)); err != nil {
				return err
			}
			continue
		}

		b, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2MB cap
		closeErr := resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if attempt == c.retryMaxAttempts {
				break
			}
			if err := sleepCtx(ctx, c.backoff(attempt)); err != nil {
				return err
			}
			continue
		}
		if closeErr != nil {
			lastErr = closeErr
			if attempt == c.retryMaxAttempts {
				break
			}
			if err := sleepCtx(ctx, c.backoff(attempt)); err != nil {
				return err
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			httpErr := &HTTPError{StatusCode: resp.StatusCode, Body: string(b)}
			mapped := mapHTTPStatus(resp.StatusCode)
			if mapped != nil {
				lastErr = errors.Join(mapped, httpErr)
			} else {
				lastErr = httpErr
			}

			// Retry only on 429/5xx
			if (resp.StatusCode == 429 || resp.StatusCode >= 500) && attempt < c.retryMaxAttempts {
				if err := sleepCtx(ctx, c.backoff(attempt)); err != nil {
					return err
				}
				continue
			}
			break
		}

		if out != nil {
			// Treat empty (or all-whitespace) body as successful when caller expects a response.
			if len(bytes.TrimSpace(b)) == 0 {
				return nil
			}
			if err := json.Unmarshal(b, out); err != nil {
				return err
			}
		}
		return nil
	}

	if lastErr != nil {
		return lastErr
	}
	return errors.New("qdrant: request failed")
}

func (c *restClient) backoff(attempt int) time.Duration {
	// attempt: 1 => after first failure
	d := c.retryBaseDelay * (1 << (attempt - 1))
	if d > c.retryMaxDelay {
		d = c.retryMaxDelay
	}
	if c.retryJitter > 0 {
		j := (rand.Float64()*2 - 1) * c.retryJitter
		d = time.Duration(float64(d) * (1 + j))
		if d < 0 {
			d = 0
		}
	}
	return d
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func isRetryableNetErr(err error) bool {
	var ne net.Error

	if errors.As(err, &ne) {
		return ne.Timeout()
	}

	s := strings.ToLower(err.Error())
	return strings.Contains(strings.ToLower(s), "connection reset") ||
		strings.Contains(strings.ToLower(s), "broken pipe") ||
		strings.Contains(strings.ToLower(s), "timeout") ||
		strings.Contains(strings.ToLower(s), "tls handshake") ||
		strings.Contains(strings.ToLower(s), "connection refused")
}
