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

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) logonHandler(rw http.ResponseWriter, req *http.Request) {
	var failedErr error

	authorizationArray := req.Header["Authorization"]

	for {
		if len(authorizationArray) == 0 {
			break
		}

		authorization := strings.TrimSpace(authorizationArray[0])
		credentials := strings.Split(authorization, " ")

		if len(credentials) != 2 || credentials[0] != "Basic" {
			break
		}

		auth, err := base64.StdEncoding.DecodeString(credentials[1])
		if err != nil {
			failedErr = err
			break
		}

		userpass := strings.Split(string(auth), ":")
		if len(userpass) != 2 {
			break
		}

		response, err := s.c.Logon(req.Context(), userpass[0], userpass[1])
		if err != nil {
			failedErr = err
			break
		}

		if response.Er != 0 {
			failedErr = fmt.Errorf("mapi error code: %x", response.Er)
			break
		}

		rw.WriteHeader(http.StatusOK)
		return
	}

	if failedErr != nil {
		s.logger.WithError(failedErr).Infoln("logon request error")
	}

	rw.Header().Set("WWW-Authenticate", "Basic realm=\"user\"")
	http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func (s *Server) logoffHandler(rw http.ResponseWriter, req *http.Request) {
	sessionIDString := req.URL.Query().Get("id")
	if sessionIDString == "" {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	sessionID, err := strconv.ParseUint(sessionIDString, 10, 64)
	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	response, err := s.c.Logoff(req.Context(), sessionID)
	if err != nil {
		s.logger.WithError(err).Errorln("logoffHandler request logoff failed")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if response.Er != 0 {
		s.logger.WithError(fmt.Errorf("mapi error code: %x", response.Er)).Errorln("logoffHandler request logoff mapi error")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *Server) userinfoHandler(rw http.ResponseWriter, req *http.Request) {
	username := req.URL.Query().Get("username")
	if username == "" {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	session := s.getSession()
	if session == nil {
		s.logger.WithError(fmt.Errorf("no server session")).Errorln("userinfoHandler request error")
		http.Error(rw, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	resolve, err := s.c.ResolveUsername(req.Context(), username, session.ID())
	if err != nil {
		s.logger.WithError(err).Errorln("userinfoHandler request resolveUserName failed")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if resolve.Er != 0 {
		s.logger.WithError(fmt.Errorf("mapi error code: %x", resolve.Er)).Errorln("userinfoHandler request resolveUserName mapi error")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := s.c.GetUser(req.Context(), resolve.UserEntryID, session.ID())
	if err != nil {
		s.logger.WithError(err).Errorln("userinfoHandler request getUser failed")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if response.Er != 0 {
		s.logger.WithError(fmt.Errorf("mapi error code: %x", response.Er)).Errorln("userinfoHandler request getUser mapi error")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(rw)
	enc.SetIndent("", "  ")
	err = enc.Encode(response.User)
	if err != nil {
		s.logger.WithError(err).Errorln("userInfoHandler request failed writing response")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}
