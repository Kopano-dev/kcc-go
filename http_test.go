/*
 * Copyright 2019 Kopano and its licensors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *	http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kcc

import (
	"crypto/tls"
	"fmt"
	"os"
)

var defaultHTTPInsecureSkipVerify = false

func init() {
	transport := DefaultHTTPTransport
	config := transport.TLSClientConfig
	if config == nil {
		config = &tls.Config{}
		transport.TLSClientConfig = config
	}

	config.ClientSessionCache = tls.NewLRUClientSessionCache(0)

	if s := os.Getenv("KCC_GO_HTTP_INSECURE_SKIP_VERIFY"); s != "" {
		switch s {
		case "off", "false", "no":
			defaultHTTPInsecureSkipVerify = false
		case "on", "true", "yes":
			defaultHTTPInsecureSkipVerify = true
		}
	}

	if defaultHTTPInsecureSkipVerify {
		config.InsecureSkipVerify = defaultHTTPInsecureSkipVerify
		fmt.Printf("Warning: kcc-go default HTTP client transport has disabled TLS verification\n")
	}
}
