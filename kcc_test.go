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
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"testing"
)

var (
	testUsername     = "user1"
	testUserPassword = "pass"

	testSSOKCOIDCUsername   *string
	testSSOKCOIDCTokenValue *string

	testX509ClientCertificate *string
	testX509ClientPrivateKey  *string

	testSystemUsername = "SYSTEM"
)

func init() {
	username := os.Getenv("TEST_USERNAME")
	if username != "" {
		testUsername = username
	}
	password := os.Getenv("TEST_PASSWORD")
	if password != "" {
		testUserPassword = password
	}
	kcoicdUsername := os.Getenv("TEST_KCOIDC_USERNAME")
	if kcoicdUsername != "" {
		testSSOKCOIDCUsername = &kcoicdUsername
	}
	kcoidcTokenValue := os.Getenv("TEST_KCOIDC_TOKEN_VALUE")
	if kcoidcTokenValue != "" {
		testSSOKCOIDCTokenValue = &kcoidcTokenValue
	}
	x509ClientCertificate := os.Getenv("TEST_X509_CERTIFICATE")
	if x509ClientCertificate != "" {
		testX509ClientCertificate = &x509ClientCertificate
	}
	x509ClientPrivateKey := os.Getenv("TEST_X509_PRIVATE_KEY")
	if x509ClientPrivateKey != "" {
		testX509ClientPrivateKey = &x509ClientPrivateKey
	}
}

func logon(ctx context.Context, t testing.TB, c *KCC, username *string, userPassword *string, logonFlags KCFlag) (*KCC, *LogonResponse) {
	if username == nil {
		username = &testUsername
	}
	if userPassword == nil {
		userPassword = &testUserPassword
	}

	if c == nil {
		c = NewKCC(nil)
	}

	resp, err := c.Logon(ctx, *username, *userPassword, logonFlags)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Er != KCSuccess {
		t.Fatalf("logon returned wrong er: got %v want 0", resp.Er)
	}

	if resp.SessionID == KCNoSessionID {
		t.Errorf("logon returned invalid session ID")
	}

	if resp.ServerGUID == "" {
		t.Errorf("logon return invalid server GUID")
	}

	return c, resp
}

func x509Logon(ctx context.Context, t testing.TB, c *KCC, username *string, userPassword *string, certFile *string, keyFile *string, logonFlags KCFlag) (*KCC, *LogonResponse) {
	if username == nil {
		username = &testUsername
	}
	if userPassword == nil {
		empty := ""
		userPassword = &empty
	}

	if c == nil {
		transport := &http.Transport{}
		if defaultHTTPInsecureSkipVerify {
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: defaultHTTPInsecureSkipVerify,
			}
		}
		client, err := NewSOAPHTTPClient(nil, &http.Client{
			Transport: transport,
		})
		if err != nil {
			t.Fatalf("failed to create SOAP HTTP client: %v", err)
		}
		c = NewKCCWithClient(client)

		if certFile == nil {
			certFile = testX509ClientCertificate
		}
		if keyFile == nil {
			keyFile = testX509ClientPrivateKey
		}
		if certFile == nil || keyFile == nil {
			t.Skip("Missing TEST_X509_CERTIFICATE or TEST_X509_PRIVATE_KEY")
		}

		err = useX509KeyPair(transport, *certFile, *keyFile)
		if err != nil {
			t.Fatalf("failed to load X509 key pair: %v", err)
		}
	} else {
		if certFile != nil || keyFile != nil {
			t.Errorf("cannot use x509Logon with already initialized client and cert/key")
		}
	}

	return logon(ctx, t, c, username, userPassword, logonFlags)
}

func ssoKCOIDCLogon(ctx context.Context, t testing.TB, c *KCC, sessionID KCSessionID, username *string, tokenValue *string, logonFlags KCFlag) (*KCC, *LogonResponse) {
	if username == nil {
		username = testSSOKCOIDCUsername
	}
	if tokenValue == nil {
		tokenValue = testSSOKCOIDCTokenValue
	}

	if username == nil || tokenValue == nil {
		t.Skip("Missing TEST_KCOIDC_USERNAME or TEST_KCOIDC_TOKEN_VALUE")
	}

	if c == nil {
		c = NewKCC(nil)
	}

	resp, err := c.SSOLogon(ctx, KOPANO_SSO_TYPE_KCOIDC, *username, []byte(*tokenValue), sessionID, logonFlags)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Er != KCSuccess {
		t.Fatalf("sso logon returned wrong er: got %v want 0", resp.Er)
	}

	if resp.SessionID == KCNoSessionID {
		t.Errorf("sso logon returned invalid session ID")
	}

	if resp.ServerGUID == "" {
		t.Errorf("sso logon return invalid server GUID")
	}

	return c, resp
}

func getUser(ctx context.Context, t testing.TB, c *KCC, userEntryID string, sessionID KCSessionID) (*KCC, *GetUserResponse) {
	if c == nil {
		c = NewKCC(nil)
	}

	if sessionID == 0 {
		_, session := logon(ctx, t, c, nil, nil, 0)
		sessionID = session.SessionID
	}

	resp, err := c.GetUser(ctx, userEntryID, sessionID)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Er != KCSuccess {
		t.Fatalf("getUser returned wrong er: got %v want 0", resp.Er)
	}

	if resp.User == nil {
		t.Fatal("getUser returned no user")
	}

	if resp.User.ID == 0 {
		t.Errorf("getUser user returned invalid User.ID: got %v", resp.User.ID)
	}

	if resp.User.UserEntryID == "" {
		t.Errorf("getUser user returned invalid User.UserID")
	}

	return c, resp
}

func TestLogon(t *testing.T) {
	_, resp := logon(context.Background(), t, nil, nil, nil, 0)
	t.Logf("Session ID  : %d", resp.SessionID)
	t.Logf("Server GUID : %v", resp.ServerGUID)
}

func TestLogonWithX509KeyPair(t *testing.T) {
	_, resp := x509Logon(context.Background(), t, nil, nil, nil, nil, nil, 0)
	t.Logf("Session ID  : %d", resp.SessionID)
	t.Logf("Server GUID : %v", resp.ServerGUID)
}

func TestLogonSystemWithX509KeyPair(t *testing.T) {
	_, resp := x509Logon(context.Background(), t, nil, &testSystemUsername, nil, nil, nil, 0)
	t.Logf("Session ID  : %d", resp.SessionID)
	t.Logf("Server GUID : %v", resp.ServerGUID)
}

func TestSSOLogon(t *testing.T) {
	_, resp := ssoKCOIDCLogon(context.Background(), t, nil, KCNoSessionID, nil, nil, 0)
	t.Logf("Session ID  : %d", resp.SessionID)
	t.Logf("Server GUID : %v", resp.ServerGUID)
}

func TestLogoff(t *testing.T) {
	ctx := context.Background()

	c, session := logon(ctx, t, nil, nil, nil, 0)

	resp, err := c.Logoff(ctx, session.SessionID)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Er != KCSuccess {
		t.Fatalf("logoff returned wrong er: got %v want 0", resp.Er)
	}
}

func BenchmarkLogon(b *testing.B) {
	ctx := context.Background()

	c := NewKCC(nil)

	for n := 0; n < b.N; n++ {
		logon(ctx, b, c, nil, nil, KOPANO_LOGON_NO_REGISTER_SESSION)
	}
}

func BenchmarkX509Logon(b *testing.B) {
	ctx := context.Background()

	c, _ := x509Logon(ctx, b, nil, nil, nil, nil, nil, KOPANO_LOGON_NO_REGISTER_SESSION)

	for n := 0; n < b.N; n++ {
		x509Logon(ctx, b, c, nil, nil, nil, nil, KOPANO_LOGON_NO_REGISTER_SESSION)
	}
}

func BenchmarkKCOIDCSSOLogon(b *testing.B) {
	ctx := context.Background()

	c := NewKCC(nil)

	for n := 0; n < b.N; n++ {
		// NOTE(longsleep): Currently SSO logon supports no flags, thus we
		// pass 0. So for now this creates sessions on the server.
		ssoKCOIDCLogon(ctx, b, c, KCNoSessionID, nil, nil, 0)
	}
}

func TestGetUserSelf(t *testing.T) {
	ctx := context.Background()

	c, session := logon(ctx, t, nil, nil, nil, 0)

	// NOTE(longsleep): Empty user EntryID returns data for the current user.
	_, resp := getUser(ctx, t, c, "", session.SessionID)

	if resp.User.Username != testUsername {
		t.Errorf("getUser returned wrong User.Username: got %v want %v", resp.User.Username, testUsername)
	}
}

func TestResolveUsernameSystemAndGetUser(t *testing.T) {
	ctx := context.Background()

	c, session := logon(ctx, t, nil, nil, nil, 0)

	resp, err := c.ResolveUsername(ctx, "SYSTEM", session.SessionID)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Er != KCSuccess {
		t.Fatalf("resolveUsername returned wrong er: got %v want 0", resp.Er)
	}

	if resp.ID == 0 {
		t.Errorf("resolveUsername returned invalid ID: got %v", resp.ID)
	}

	if resp.UserEntryID == "" {
		t.Errorf("getUser user returned invalid UserID")
	}

	_, userResp := getUser(ctx, t, c, resp.UserEntryID, session.SessionID)

	if userResp.User.Username != "SYSTEM" {
		t.Errorf("getUser of resolved SYSTEM user did not return the SYSTEM user: got %v", userResp.User.Username)
	}

	if userResp.User.ID != resp.ID {
		t.Errorf("resolveUsername user.ID does not match getUser result: got %v want %v", userResp.User.ID, resp.ID)
	}

	if userResp.User.UserEntryID != resp.UserEntryID {
		t.Errorf("resolveUsername user.UserEntryID does not match getUser result: got %v want %v", userResp.User.UserEntryID, resp.UserEntryID)
	}
}
