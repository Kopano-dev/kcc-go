/*
 * Copyright 2017-2019 Kopano and its licensors
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

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"stash.kopano.io/kgol/kcc-go"
	"stash.kopano.io/kgol/kcc-go/cmd"
)

func main() {
	cmd.RootCmd.AddCommand(commandServe())

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func commandServe() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve [...args]",
		Short: "Start server and listen for requests",
		Run: func(cmd *cobra.Command, args []string) {
			if err := serve(cmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
	serveCmd.Flags().String("listen", "127.0.0.1:8769", "TCP listen address")
	serveCmd.Flags().String("server-uri", "", "Kopano server URI")
	serveCmd.Flags().String("server-auth-pem", "", "Full path to a PEM encoded x509 certificate with private key file")
	serveCmd.Flags().Bool("insecure", false, "Disable TLS certificate and hostname validation")

	return serveCmd
}

func serve(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: &logrus.TextFormatter{},
		Level:     logrus.DebugLevel,
	}

	logger.Infoln("serve start")

	var serverURI *url.URL
	var tlsConfig *tls.Config

	listenAddr, _ := cmd.Flags().GetString("listen")
	if serverURIString, err := cmd.Flags().GetString("server-uri"); err == nil && serverURIString != "" {
		// Parse serverURI
		serverURI, err = url.Parse(serverURIString)
		if err != nil {
			return err
		}
	}

	username := "SYSTEM"
	password := ""
	if usernameOverride := os.Getenv("KOPANO_USERNAME"); usernameOverride != "" {
		username = usernameOverride
	}
	if passwordOverride := os.Getenv("KOPANO_PASSWORD"); passwordOverride != "" {
		password = passwordOverride
	}

	switch serverURI.Scheme {
	case "https":
		tlsConfig = &tls.Config{
			ClientSessionCache: tls.NewLRUClientSessionCache(0),
		}

		tlsInsecureSkipVerify, _ := cmd.Flags().GetBool("insecure")
		if tlsInsecureSkipVerify {
			// NOTE(longsleep): This disable http2 client support. See https://github.com/golang/go/issues/14275 for reasons.
			tlsConfig.InsecureSkipVerify = true
			logger.Warnln("insecure mode, TLS client connections are susceptible to man-in-the-middle attacks")
			logger.Debugln("http2 client support is disabled (insecure mode)")
		}

		kcc.DefaultHTTPClient.Transport.(*http.Transport).TLSClientConfig = tlsConfig
		fallthrough
	case "http":
	case "file":
	default:
		return fmt.Errorf("unsupported server-uri scheme: %v", serverURI.Scheme)
	}

	if serverAuthPEM, err := cmd.Flags().GetString("server-auth-pem"); err == nil && serverAuthPEM != "" {
		if tlsConfig == nil {
			return fmt.Errorf("this server-uri cannot be used together with server-auth-cert, a https:// uri is required")
		}

		_, err := kcc.SetX509KeyPair(serverAuthPEM, serverAuthPEM, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to set server-auth-pem file: %v", err)
		}
		logger.Infoln("using TLS client certificate for server auth")
	}

	srv := NewServer(listenAddr, serverURI, logger)

	logger.Infof("serve started")
	return srv.Serve(ctx, username, password)
}
