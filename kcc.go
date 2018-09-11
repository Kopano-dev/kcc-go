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

package kcc // import "stash.kopano.io/kgol/kcc-go"

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
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
	return fmt.Sprintf("KCC(%s:%s)", c.app, c.uri)
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
