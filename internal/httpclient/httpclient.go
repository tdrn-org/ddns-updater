// httpclient.go
//
// Copyright (C) 2023-2026 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package httpclient

import (
	"crypto/tls"
	"net/http"
	"time"
)

const DefaultTimeout time.Duration = 0

func PrepareClient(timeout time.Duration, insecureSkipVerify bool) *http.Client {
	transport := &http.Transport{
		ResponseHeaderTimeout: timeout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

func ReqUserAgent(req *http.Request, userAgent string) *http.Request {
	req.Header.Set("User-Agent", userAgent)
	return req
}

func ReqAuthorization(req *http.Request, username string, password string) *http.Request {
	req.SetBasicAuth(username, password)
	return req
}
