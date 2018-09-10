/*
 * Copyright 2018 Kopano and its licensors
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

// SSOType is the type of SSO to use with single sign on.
type SSOType string

func (sst SSOType) String() string {
	return string(sst)
}

// Known Kopano SSO types.
const (
	KOPANO_SSO_TYPE_NTML   SSOType = "NTLM"
	KOPANO_SSO_TYPE_KCOIDC SSOType = "KCOIDC"
	KOPANO_SSO_TYPE_KRB5   SSOType = ""
)
