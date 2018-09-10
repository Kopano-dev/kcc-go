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
	"os"
	"testing"
)

var (
	testUsername     = "user1"
	testUserPassword = "pass"

	testSSOKCOIDCUsername   *string
	testSSOKCOIDCTokenValue *string
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
}

func logon(ctx context.Context, t testing.TB, c *KCC, logonFlags KCFlag) (*KCC, *LogonResponse) {
	if c == nil {
		c = NewKCC(nil)
	}

	resp, err := c.Logon(ctx, testUsername, testUserPassword, logonFlags)
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

func ssoKCOIDCLogon(ctx context.Context, t testing.TB, c *KCC, sessionID KCSessionID, logonFlags KCFlag) (*KCC, *LogonResponse) {
	if testSSOKCOIDCUsername == nil || testSSOKCOIDCTokenValue == nil {
		t.Skip("Missing TEST_KCOIDC_USERNAME or TEST_KCOIDC_TOKEN_VALUE")
	}

	if c == nil {
		c = NewKCC(nil)
	}

	resp, err := c.SSOLogon(ctx, KOPANO_SSO_TYPE_KCOIDC, *testSSOKCOIDCUsername, []byte(*testSSOKCOIDCTokenValue), sessionID, logonFlags)
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
		_, session := logon(ctx, t, c, 0)
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
	_, resp := logon(context.Background(), t, nil, 0)
	t.Logf("Session ID  : %d", resp.SessionID)
	t.Logf("Server GUID : %v", resp.ServerGUID)
}

func TestSSOLogon(t *testing.T) {
	_, resp := ssoKCOIDCLogon(context.Background(), t, nil, KCNoSessionID, 0)
	t.Logf("Session ID  : %d", resp.SessionID)
	t.Logf("Server GUID : %v", resp.ServerGUID)

}

func TestLogoff(t *testing.T) {
	ctx := context.Background()

	c, session := logon(ctx, t, nil, 0)

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
		logon(ctx, b, c, KOPANO_LOGON_NO_REGISTER_SESSION)
	}
}

func BenchmarkKCOIDCSSOLogon(b *testing.B) {
	ctx := context.Background()

	c := NewKCC(nil)

	for n := 0; n < b.N; n++ {
		ssoKCOIDCLogon(ctx, b, c, KCNoSessionID, KOPANO_LOGON_NO_REGISTER_SESSION)
	}
}

func TestGetUserSelf(t *testing.T) {
	ctx := context.Background()

	c, session := logon(ctx, t, nil, 0)

	// NOTE(longsleep): Empty user EntryID returns data for the current user.
	_, resp := getUser(ctx, t, c, "", session.SessionID)

	if resp.User.Username != testUsername {
		t.Errorf("getUser returned wrong User.Username: got %v want %v", resp.User.Username, testUsername)
	}
}

func TestResolveUsernameSystemAndGetUser(t *testing.T) {
	ctx := context.Background()

	c, session := logon(ctx, t, nil, 0)

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
