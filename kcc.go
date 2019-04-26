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

package kcc // import "stash.kopano.io/kgol/kcc-go"

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	// DefaultURI is the default Kopano server URI to be used when no URI is
	// given when constructing a KCC instance.
	DefaultURI = "http://127.0.0.1:236"
	// DefaultAppName is the default client app name as sent to the server.
	DefaultAppName = "kcc-go"
	// Version specifies the version string of this client implementation.
	Version = "0.0.0-dev"
	// ClientVersion specifies the version of this clients API implementation,
	ClientVersion = 8
)

var debug = false

func init() {
	uri := os.Getenv("KOPANO_SERVER_DEFAULT_URI")
	if uri != "" {
		DefaultURI = uri
	}
	debug = os.Getenv("KCC_GO_DEBUG") != ""
}

// A KCC is the client implementation base object containing the HTTP connection
// pool and other references to interface with a Kopano server via SOAP.
type KCC struct {
	uri string

	Client       SOAPClient
	Capabilities KCFlag

	app [2]string
}

// NewKCC constructs a KCC instance with the provided URI. If no URI is passed,
// the current DefaultURI value will tbe used.
func NewKCC(uri *url.URL) *KCC {
	if uri == nil {
		uri, _ = url.Parse(DefaultURI)
	}
	soap, _ := NewSOAPClient(uri)

	c := &KCC{
		uri: uri.String(),
		app: [2]string{DefaultAppName, Version},

		Client:       soap,
		Capabilities: DefaultClientCapabilities,
	}

	return c
}

func (c *KCC) String() string {
	return fmt.Sprintf("KCC(%s)", c.uri)
}

// LoadX509KeyPair enables TLS client authentication for the associated client
// using the provided certificate and private key. The files must contain PEM
// encoded data.
func (c *KCC) LoadX509KeyPair(certFile, keyFile string) error {
	client, _ := c.Client.(*SOAPHTTPClient)
	if client == nil {
		return fmt.Errorf("SOAP client type %T does not support TLS client auth", c.Client)
	}
	if !strings.HasPrefix(client.URI, "https://") {
		return fmt.Errorf("SOAP client not using https")
	}

	transport := client.Client.Transport.(*http.Transport)
	config := transport.TLSClientConfig
	if config == nil {
		config = &tls.Config{}
	} else {
		config = config.Clone()
	}
	if _, err := SetX509KeyPair(certFile, keyFile, config); err != nil {
		return err
	}

	transport.TLSClientConfig = config
	return nil
}

// SetClientApp sets the clients app details as sent with requests to the
// accociated server.
func (c *KCC) SetClientApp(name, version string) error {
	c.app = [2]string{name, version}
	return nil
}

// Logon creates a session with the Kopano server using the provided credentials.
func (c *KCC) Logon(ctx context.Context, username, password string, logonFlags KCFlag) (*LogonResponse, error) {
	payload := `<ns:logon><szUsername>` +
		username +
		`</szUsername><szPassword>` +
		password +
		`</szPassword><szImpersonateUser/><ulCapabilities>` +
		c.Capabilities.String() +
		`</ulCapabilities><ulFlags>` +
		logonFlags.String() +
		`</ulFlags><szClientApp>` +
		c.app[0] +
		`</szClientApp><szClientAppVersion>` +
		c.app[1] +
		`</szClientAppVersion><clientVersion>` +
		string(ClientVersion) +
		`</clientVersion></ns:logon>`

	var logonResponse LogonResponse
	err := c.Client.DoRequest(ctx, &payload, &logonResponse)

	return &logonResponse, err
}

// SSOLogon creates a session with the Kopano server using the provided credentials.
func (c *KCC) SSOLogon(ctx context.Context, prefix SSOType, username string, input []byte, sessionID KCSessionID, logonFlags KCFlag) (*LogonResponse, error) {
	if logonFlags != 0 {
		return nil, fmt.Errorf("logon flags are not support by sso logon")
	}

	// Add prefix value.
	lpInput := make([]byte, 0, len(prefix)+len(input))
	lpInput = append(lpInput, prefix.String()...)
	lpInput = append(lpInput, input...)

	// NOTE(longsleep): There is currently no way to specify flags when using
	// SSOLogon. This means, a new session is created when none was given and
	// the call will fail with error if the given session does not exist.
	payload := `<ns:ssoLogon><szUsername>` +
		username +
		`</szUsername><lpInput>` +
		base64.StdEncoding.EncodeToString(lpInput) +
		`</lpInput><szImpersonateUser/><ulCapabilities>` +
		c.Capabilities.String() +
		`</ulCapabilities><szClientApp>` +
		c.app[0] +
		`</szClientApp><szClientAppVersion>` +
		c.app[1] +
		`</szClientAppVersion><clientVersion>` +
		string(ClientVersion) +
		`</clientVersion><ulSessionId>` +
		sessionID.String() +
		`</ulSessionId></ns:ssoLogon>`

	var logonResponse LogonResponse
	err := c.Client.DoRequest(ctx, &payload, &logonResponse)

	return &logonResponse, err
}

// Logoff terminates the provided session with the Kopano server.
func (c *KCC) Logoff(ctx context.Context, sessionID KCSessionID) (*LogoffResponse, error) {
	payload := `<ns:logoff><ulSessionId>` +
		sessionID.String() +
		`</ulSessionId></ns:logoff>`

	var logoffResponse LogoffResponse
	err := c.Client.DoRequest(ctx, &payload, &logoffResponse)

	return &logoffResponse, err
}

// ResolveUsername looks up the user ID of the provided username using the
// provided session.
func (c *KCC) ResolveUsername(ctx context.Context, username string, sessionID KCSessionID) (*ResolveUserResponse, error) {
	payload := `<ns:resolveUsername><lpszUsername>` +
		username +
		`</lpszUsername><ulSessionId>` +
		sessionID.String() +
		`</ulSessionId></ns:resolveUsername>`

	var resolveUserResponse ResolveUserResponse
	err := c.Client.DoRequest(ctx, &payload, &resolveUserResponse)

	return &resolveUserResponse, err
}

// GetUser fetches a user's detail meta data of the provided user Entry
// ID using the provided session.
func (c *KCC) GetUser(ctx context.Context, userEntryID string, sessionID KCSessionID) (*GetUserResponse, error) {
	payload := `<ns:getUser><sUserId>` +
		userEntryID +
		`</sUserId><ulSessionId>` +
		sessionID.String() +
		`</ulSessionId></ns:getUser>`

	var getUserResponse GetUserResponse
	err := c.Client.DoRequest(ctx, &payload, &getUserResponse)

	return &getUserResponse, err
}

// ABResolveNames searches the AB for the provided props using the provided
// request data and flags.
func (c *KCC) ABResolveNames(ctx context.Context, props []PT, request map[PT]interface{}, requestFlags ABFlag, sessionID KCSessionID, resolveNamesFlags KCFlag) (*ABResolveNamesResponse, error) {
	payload := `<ns:abResolveNames>` +
		`<ulSessionId>` +
		sessionID.String() +
		`</ulSessionId>` +
		`<lpaPropTag SOAP-ENC:arrayType="xsd:unsignedInt[` + fmt.Sprintf("%d", len(props)) + `]">`
	for _, prop := range props {
		payload += fmt.Sprintf("<item>%d</item>\n", prop)
	}
	payload += `</lpaPropTag>` +
		`<lpsRowSet SOAP-ENC:arrayType="propVal[][` + fmt.Sprintf("%d", len(request)) + `]">`
	for prop, value := range request {
		payload += `<item SOAP-ENC:arrayType="propVal[1]"><item>` +
			fmt.Sprintf("<ulPropTag>%d</ulPropTag>", prop)
		switch tv := value.(type) {
		case string:
			payload += fmt.Sprintf("<lpszA>%s</lpszA>", tv)
		default:
			return nil, fmt.Errorf("unsupported type in request map value: %v", value)
		}
		payload += `</item></item>`
	}
	payload += `</lpsRowSet>` +
		`<lpaFlags>` +
		fmt.Sprintf("<item>%s</item>", requestFlags) +
		`</lpaFlags>` +
		fmt.Sprintf("<ulFlags>%s</ulFlags>", resolveNamesFlags) +
		`</ns:abResolveNames>`

	var abResolveNamesResponse ABResolveNamesResponse
	err := c.Client.DoRequest(ctx, &payload, &abResolveNamesResponse)

	return &abResolveNamesResponse, err
}
