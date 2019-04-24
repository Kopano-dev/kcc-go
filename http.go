/*
 * Copyright 2017 Kopano and its licensors
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License, version 3,
 * as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package kcc

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Default HTTP client settings.
var (
	DefaultHTTPTimeoutSeconds         int64 = 10
	DefaultHTTPMaxIdleConns                 = 100
	DefaultHTTPMaxIdleConnsPerHost          = 100
	DefaultHTTPIdleConnTimeoutSeconds int64 = 90
	DefaultHTTPDialTimeoutSeconds     int64 = 30
	DefaultHTTPKeepAliveSeconds       int64 = 120
	DefaultHTTPDualStack                    = true
)

// DefaultHTTPClient is the default Client as used by KCC for HTTP SOAP requests.
var DefaultHTTPClient *http.Client

// DefaultHTTPTransport is the default Transpart as used by KCC for HTTP SOAP requests.
var DefaultHTTPTransport *http.Transport

func init() {
	debug = os.Getenv("KCC_GO_DEBUG") != ""

	if s := os.Getenv("KCC_GO_HTTP_TIMEOUT"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			DefaultHTTPTimeoutSeconds = n
		}
	}
	if s := os.Getenv("KCC_GO_HTTP_MAX_IDLE_CONNS"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 0); err == nil {
			DefaultHTTPMaxIdleConns = int(n)
		}
	}
	if s := os.Getenv("KCC_GO_HTTP_MAX_IDLE_CONNS_PER_HOST"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 0); err == nil {
			DefaultHTTPMaxIdleConnsPerHost = int(n)
		}
	}
	if s := os.Getenv("KCC_GO_HTTP_IDLE_CONN_TIMEOUT"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			DefaultHTTPIdleConnTimeoutSeconds = n
		}
	}
	if s := os.Getenv("KCC_GO_HTTP_DIAL_TIMEOUT"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			DefaultHTTPDialTimeoutSeconds = n
		}
	}
	if s := os.Getenv("KCC_GO_HTTP_KEEPALIVE"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			DefaultHTTPKeepAliveSeconds = n
		}
	}
	if s := os.Getenv("KCC_GO_HTTP_DUALSTACK"); s != "" {
		switch s {
		case "off", "false", "no":
			DefaultHTTPDualStack = false
		case "on", "true", "yes":
			DefaultHTTPDualStack = true
		}
	}

	dialer := &net.Dialer{
		Timeout:   time.Duration(DefaultHTTPDialTimeoutSeconds) * time.Second,
		KeepAlive: time.Duration(DefaultHTTPKeepAliveSeconds) * time.Second,
		DualStack: DefaultHTTPDualStack,
	}

	DefaultHTTPTransport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          DefaultHTTPMaxIdleConns,
		MaxIdleConnsPerHost:   DefaultHTTPMaxIdleConnsPerHost,
		IdleConnTimeout:       time.Duration(DefaultHTTPIdleConnTimeoutSeconds) * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	DefaultHTTPClient = &http.Client{
		Timeout:   time.Duration(DefaultHTTPTimeoutSeconds) * time.Second,
		Transport: DefaultHTTPTransport,
	}

	if debug {
		fmt.Printf("HTTP client: %+v\n", DefaultHTTPClient)
		fmt.Printf("HTTP client transport: %+v\n", DefaultHTTPTransport)
		fmt.Printf("HTTP client transport dial: %+v\n", dialer)
	}
}
