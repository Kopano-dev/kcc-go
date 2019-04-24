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
)

// SetX509KeyPair reads and parses a public/private key pair from a pair of
// files and adds the resulting certificate to the provided TLs config. If the
// provided TLS config is nil, a new empty one will be created and returned.
func SetX509KeyPair(certFile, keyFile string, config *tls.Config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return config, err
	}

	if config == nil {
		config = &tls.Config{}
	}
	config.Certificates = []tls.Certificate{cert}

	return config, nil
}
