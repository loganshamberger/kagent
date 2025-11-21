/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/kagent-dev/kagent/go/api/v1alpha1"
)

type Checker struct {
	timeout time.Duration
	client  *http.Client
}

func NewChecker(timeout time.Duration) *Checker {
	return &Checker{
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: timeout,
				}).DialContext,
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableKeepAlives:   true,
				DisableCompression:  true,
				MaxIdleConnsPerHost: 1,
			},
		},
	}
}

func (c *Checker) CheckEndpoints(ctx context.Context, endpoints []v1alpha1.AgentEndpoint) (bool, error) {
	if len(endpoints) == 0 {
		return false, fmt.Errorf("no endpoints to check")
	}

	for _, endpoint := range endpoints {
		healthy, err := c.checkEndpoint(ctx, endpoint)
		if err != nil {
			continue
		}
		if healthy {
			return true, nil
		}
	}

	return false, fmt.Errorf("all endpoints unhealthy")
}

func (c *Checker) checkEndpoint(ctx context.Context, endpoint v1alpha1.AgentEndpoint) (bool, error) {
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(checkCtx, http.MethodHead, endpoint.URL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check endpoint: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 500, nil
}
