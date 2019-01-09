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
	"sort"
	"strconv"
	"strings"
	"time"

	"stash.kopano.io/kgol/kcc-go"
)

func (s *Server) logonHandler(rw http.ResponseWriter, req *http.Request) {
	var failedErr error
	var noSession bool

	authorizationArray := req.Header["Authorization"]
	if sessionQueryString := req.URL.Query().Get("session"); sessionQueryString == "0" {
		noSession = true
	}

	for {
		if len(authorizationArray) == 0 {
			rw.Header().Set("WWW-Authenticate", "Basic realm=\"Kopano\"")
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		authorization := strings.TrimSpace(authorizationArray[0])
		credentials := strings.Split(authorization, " ")

		if len(credentials) != 2 || credentials[0] != "Basic" {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		auth, err := base64.StdEncoding.DecodeString(credentials[1])
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		userpass := strings.Split(string(auth), ":")
		if len(userpass) != 2 {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var logonFlags kcc.KCFlag
		if noSession {
			logonFlags |= kcc.KOPANO_LOGON_NO_REGISTER_SESSION
		}
		response, err := s.c.Logon(req.Context(), userpass[0], userpass[1], logonFlags)
		if err != nil {
			failedErr = err
			break
		}
		if response.Er == kcc.KCERR_LOGON_FAILED {
			rw.Header().Set("WWW-Authenticate", "Basic realm=\"Kopano\"")
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		} else if response.Er != kcc.KCSuccess {
			failedErr = response.Er
			break
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)

		if noSession {
			return
		}

		enc := json.NewEncoder(rw)
		enc.SetIndent("", "  ")
		err = enc.Encode(response)
		if err != nil {
			s.logger.WithError(err).Errorln("logon request failed writing response")
		}

		return
	}

	if failedErr != nil {
		s.logger.WithError(failedErr).Infoln("logon request error")
	}

	http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

	response, err := s.c.Logoff(req.Context(), kcc.KCSessionID(sessionID))
	if err != nil {
		s.logger.WithError(err).Errorln("logoffHandler request logoff failed")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if response.Er != kcc.KCSuccess {
		s.logger.WithError(response.Er).Errorln("logoffHandler request logoff mapi error")
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

	retries := 0
	for {
		session := s.getSession()
		if session == nil || !session.IsActive() {
			s.logger.WithError(fmt.Errorf("no server session")).Errorln("userinfoHandler request error")
			http.Error(rw, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}

		var failedErr error
		for {
			resolve, err := s.c.ResolveUsername(req.Context(), username, session.ID())
			if err != nil {
				s.logger.WithError(err).Errorln("userinfoHandler request resolveUserName failed")
				failedErr = err
				break

			}
			if resolve.Er == kcc.KCERR_NOT_FOUND {
				http.Error(rw, resolve.Er.Error(), http.StatusNotFound)
				return
			} else if resolve.Er != kcc.KCSuccess {
				s.logger.WithError(resolve.Er).Errorln("userinfoHandler request resolveUserName mapi error")
				failedErr = resolve.Er
				break
			}

			response, err := s.c.GetUser(req.Context(), resolve.UserEntryID, session.ID())
			if err != nil {
				s.logger.WithError(err).Errorln("userinfoHandler request getUser failed")
				failedErr = err
				break
			}
			if response.Er != kcc.KCSuccess {
				s.logger.WithError(response.Er).Errorln("userinfoHandler request getUser mapi error")
				failedErr = resolve.Er
				break
			}

			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)

			enc := json.NewEncoder(rw)
			enc.SetIndent("", "  ")
			err = enc.Encode(response.User)
			if err != nil {
				s.logger.WithError(err).Errorln("userInfoHandler request failed writing response")
				return
			}

			return
		}

		if failedErr != nil {
			switch failedErr {
			case kcc.KCERR_END_OF_SESSION:
				session.Destroy(req.Context(), false)
			default:
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		// If reach here, its a retry.
		select {
		case <-time.After(50 * time.Millisecond):
			// Retry now.
		case <-req.Context().Done():
			// Abort.
			return
		}

		retries++
		if retries > 3 {
			s.logger.WithField("retry", retries).Errorln("userInfoHandler giving up")
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		s.logger.WithField("retry", retries).Debugln("userInfoHandler retry in progress")
	}
}

func (s *Server) errorSenseHandler(rw http.ResponseWriter, req *http.Request) {
	er := req.URL.Query().Get("er")

	if er == "" {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var intEr uint64
	var errInt error
	if strings.HasPrefix(er, "0x") || strings.HasPrefix(er, "0X") {
		intEr, errInt = strconv.ParseUint(er[2:], 16, 64)
	} else {
		intEr, errInt = strconv.ParseUint(er, 10, 64)
	}

	err := kcc.KCError(intEr)
	if errInt != nil {
		http.Error(rw, errInt.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(rw, "%s\n", err)
}

func (s *Server) errorsList(rw http.ResponseWriter, req *http.Request) {
	var keys []kcc.KCError
	for k := range kcc.KCErrorNameMap {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		a := uint64(keys[i])
		b := uint64(keys[j])

		return a < b
	})

	for _, k := range keys {
		fmt.Fprintf(rw, "%#v : %d : %s\n", k, k, k)
	}
}

func (s *Server) abResolveNamesHandler(rw http.ResponseWriter, req *http.Request) {
	names := req.URL.Query()["name"]
	if len(names) == 0 {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	props := []kcc.PT{
		kcc.PR_ADDRTYPE,
		kcc.PR_EMAIL_ADDRESS,
		kcc.PR_SMTP_ADDRESS,
		kcc.PR_ENTRYID,
		kcc.PR_INSTANCE_KEY,
		kcc.PR_OBJECT_TYPE,
		kcc.PR_RECORD_KEY,
		kcc.PR_SEARCH_KEY,
		0x6783000a, // ??
	}

	request := make(map[kcc.PT]interface{})
	for _, name := range names {
		request[kcc.PR_DISPLAY_NAME] = name
	}

	var resolveNamesFlags kcc.KCFlag

	retries := 0
	for {
		session := s.getSession()
		if session == nil || !session.IsActive() {
			s.logger.WithError(fmt.Errorf("no server session")).Errorln("userinfoHandler request error")
			http.Error(rw, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}

		var failedErr error
		for {
			response, err := s.c.ABResolveNames(req.Context(), props, request, kcc.MAPI_UNRESOLVED, session.ID(), resolveNamesFlags)
			if err != nil {
				s.logger.WithError(err).Errorln("abResolveNamesHandler request abResolveNames failed")
				failedErr = err
				break
			}

			if response.Er != kcc.KCSuccess {
				s.logger.WithError(response.Er).Errorln("abResolveNamesHandler request abResolveNames mapi error")
				failedErr = response.Er
				break
			}

			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)

			enc := json.NewEncoder(rw)
			enc.SetIndent("", "  ")
			err = enc.Encode(response)
			if err != nil {
				s.logger.WithError(err).Errorln("abResolveNamesHandler request failed writing response")
				return
			}

			return
		}

		if failedErr != nil {
			switch failedErr {
			case kcc.KCERR_END_OF_SESSION:
				session.Destroy(req.Context(), false)
			default:
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		// If reach here, its a retry.
		select {
		case <-time.After(50 * time.Millisecond):
			// Retry now.
		case <-req.Context().Done():
			// Abort.
			return
		}

		retries++
		if retries > 3 {
			s.logger.WithField("retry", retries).Errorln("userInfoHandler giving up")
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		s.logger.WithField("retry", retries).Debugln("userInfoHandler retry in progress")
	}
}
