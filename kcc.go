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
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

var (
	// DefaultURI is the default Kopano server URI to be used when no URI is
	// given when constructing a KCC instance.
	DefaultURI = "http://kc-dev-1.lxd:11236"
	// Version specifies the version string of this client implementation.
	Version = "0.0.0-dev"
)

func init() {
	uri := os.Getenv("KOPANO_SERVER_DEFAULT_URI")
	if uri != "" {
		DefaultURI = uri
	}
}

// A KCC is the client implementation base object containing the HTTP connection
// pool and other references to interface with a Kopano server via SOAP.
type KCC struct {
	uri string

	Client *http.Client
}

// NewKCC constructs a KCC instance with the provided URI. If no URI is passed,
// the current DefaultURI value will tbe used.
func NewKCC(uri *string) *KCC {
	if uri == nil {
		uri = &DefaultURI
	}

	return &KCC{
		uri: *uri,

		Client: DefaultHTTPClient,
	}
}

// Logon creates a session with the Kopano server using the provided credentials.
func (c *KCC) Logon(ctx context.Context, username, password string) (*LogonResponse, error) {
	payload := `<ns:logon><szUsername>` +
		username +
		`</szUsername><szPassword>` +
		password +
		`</szPassword><szImpersonateUser/><ulCapabilities>26451</ulCapabilities><ulFlags>0</ulFlags><szClientApp>kcc-go</szClientApp><szClientAppVersion>` +
		Version +
		`</szClientAppVersion></ns:logon>`

	req, _ := newSOAPRequest(ctx, c.uri, &payload)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http response status: %v", resp.StatusCode)
	}

	var logonResponse *LogonResponse
	err = parseSOAPResponse(resp.Body, &logonResponse)
	if err != nil {
		return nil, err
	}

	return logonResponse, nil
}

// Logoff terminates the provided session with the Kopano server.
func (c *KCC) Logoff(ctx context.Context, sessionID uint64) (*LogoffResponse, error) {
	payload := `<ns:logoff><ulSessionId>` +
		strconv.FormatUint(sessionID, 10) +
		`</ulSessionId></ns:logoff>`

	req, _ := newSOAPRequest(ctx, c.uri, &payload)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http response status: %v", resp.StatusCode)
	}

	var logoffResponse *LogoffResponse
	err = parseSOAPResponse(resp.Body, &logoffResponse)
	if err != nil {
		return nil, err
	}

	return logoffResponse, nil
}

// ResolveUsername looks up the user ID of the provided username using the
// provided session.
func (c *KCC) ResolveUsername(ctx context.Context, username string, sessionID uint64) (*ResolveUserResponse, error) {
	payload := `<ns:resolveUsername><lpszUsername>` +
		username +
		`</lpszUsername><ulSessionId>` +
		strconv.FormatUint(sessionID, 10) +
		`</ulSessionId></ns:resolveUsername>`

	req, _ := newSOAPRequest(ctx, c.uri, &payload)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http response status: %v", resp.StatusCode)
	}

	var resolveUserResponse *ResolveUserResponse
	err = parseSOAPResponse(resp.Body, &resolveUserResponse)
	if err != nil {
		return nil, err
	}

	return resolveUserResponse, nil
}

// GetUser fetches a user's detail meta data of the provided user ID using the
// provided session.
func (c *KCC) GetUser(ctx context.Context, userID uint64, sessionID uint64) (*GetUserResponse, error) {
	payload := `<ns:getUser><ulUserId>` +
		strconv.FormatUint(userID, 10) +
		`</ulUserId><ulSessionId>` +
		strconv.FormatUint(sessionID, 10) +
		`</ulSessionId></ns:getUser>`

	req, _ := newSOAPRequest(ctx, c.uri, &payload)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http response status: %v", resp.StatusCode)
	}

	var getUserResponse *GetUserResponse
	err = parseSOAPResponse(resp.Body, &getUserResponse)
	if err != nil {
		return nil, err
	}

	return getUserResponse, nil
}
