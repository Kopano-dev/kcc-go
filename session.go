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
	"sync"
	"time"
)

var (
	// SessionAutorefreshInterval defines the interval when sessions are auto
	// refreshed automatically.
	SessionAutorefreshInterval = 5 * time.Minute
)

// Session holds the data structures to keep a session open on the accociated
// Kopano server.
type Session struct {
	id         uint64
	serverGUID string
	active     bool

	mutex     sync.RWMutex
	ctx       context.Context
	ctxCancel context.CancelFunc
	c         *KCC
}

// NewSession connects to the provided server with the provided parameters,
// creates a new Session which will be automatically refreshed until detroyed.
func NewSession(ctx context.Context, c *KCC, username, password string) (*Session, error) {
	if c == nil {
		c = NewKCC(nil)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	resp, err := c.Logon(ctx, username, password)
	if err != nil {
		return nil, fmt.Errorf("create session logon failed: %v", err)
	}

	if resp.Er != 0 {
		return nil, fmt.Errorf("create session logon mapi error: %x", resp.Er)
	}

	if resp.SessionID == 0 {
		return nil, fmt.Errorf("create session logon returned invalid session ID")
	}

	if resp.ServerGUID == "" {
		return nil, fmt.Errorf("create sesion logon return invalid server GUID")
	}

	sessionCtx, cancel := context.WithCancel(ctx)

	s := &Session{
		id:         resp.SessionID,
		serverGUID: resp.ServerGUID,

		active:    true,
		ctx:       sessionCtx,
		ctxCancel: cancel,
		c:         c,
	}

	ticker := time.NewTicker(SessionAutorefreshInterval)
	stop := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				stop <- true
			case <-ticker.C:
				err := s.refresh()
				if err != nil {
					s.Destroy(ctx)
					stop <- true
				}
			case <-stop:
				return
			}
		}
	}()

	return s, nil
}

// Context returns the accociated Session's context.
func (s *Session) Context() context.Context {
	return s.ctx
}

// IsActive retruns true when the accociated Session is not destroyed and if the
// last refresh was successfull.
func (s *Session) IsActive() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.active
}

// ID returns the accociated Session's ID.
func (s *Session) ID() uint64 {
	return s.id
}

// Destroy logs off the accociated Session at the accociated Server and stops
// auto refreshing by cancelling the accociated Session's Context. An error is
// retruned if the logoff request fails.
func (s *Session) Destroy(ctx context.Context) error {
	s.mutex.Lock()
	if !s.active {
		s.mutex.Unlock()
		return nil
	}
	s.active = false
	s.mutex.Unlock()
	s.ctxCancel()

	resp, err := s.c.Logoff(ctx, s.id)
	if err != nil {
		return fmt.Errorf("logoff session logoff failed: %v", err)
	}

	if resp.Er != 0 {
		return fmt.Errorf("logoff session logoff error: %x", resp.Er)
	}

	return nil
}

func (s *Session) String() string {
	return fmt.Sprintf("Session(%v)", s.id)
}

func (s *Session) refresh() error {
	s.mutex.RLock()
	active := s.active
	s.mutex.RUnlock()
	if !active {
		return nil
	}

	resp, err := s.c.ResolveUsername(s.ctx, "SYSTEM", s.id)
	if err != nil {
		return fmt.Errorf("refresh session resolveUsername failed: %v", err)
	}

	if resp.Er != 0 {
		return fmt.Errorf("refresh session resolveUserrname mapi error: %x", resp.Er)
	}

	return nil
}
