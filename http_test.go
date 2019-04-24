/*
 * Copyright 2019 Kopano and its licensors
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
	"crypto/tls"
	"fmt"
	"os"
)

var defaultHTTPInsecureSkipVerify = false

func init() {
	if s := os.Getenv("KCC_GO_HTTP_INSECURE_SKIP_VERIFY"); s != "" {
		switch s {
		case "off", "false", "no":
			defaultHTTPInsecureSkipVerify = false
		case "on", "true", "yes":
			defaultHTTPInsecureSkipVerify = true
		}
	}

	if defaultHTTPInsecureSkipVerify {
		transport := DefaultHTTPTransport
		config := transport.TLSClientConfig
		if config == nil {
			config = &tls.Config{}
		} else {
			config = config.Clone()
		}
		config.InsecureSkipVerify = defaultHTTPInsecureSkipVerify
		transport.TLSClientConfig = config
		fmt.Printf("Warning: kcc-go default HTTP client transport has disabled TLS verification\n")
	}
}
