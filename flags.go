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

// Kopano capability flags as defined in provider/include/kcore.hpp. This only
// defines the flags actually used or understood by kcc-go.
const (
	KOPANO_CAP_LARGE_SESSIONID = 0x0010
	KOPANO_CAP_MULTI_SERVER    = 0x0040
	KOPANO_CAP_ENHANCED_ICS    = 0x0100
	KOPANO_CAP_UNICODE         = 0x0200
)

// DefaultClientCapabilities groups the default client caps sent by kcc.
var DefaultClientCapabilities uint64 = KOPANO_CAP_UNICODE |
	KOPANO_CAP_LARGE_SESSIONID |
	KOPANO_CAP_MULTI_SERVER |
	KOPANO_CAP_ENHANCED_ICS

// Kopano logon flags as defined in provider/include/kcore.hpp. This only
// defines the flags actually used or understood by kcc-go.
const (
	KOPANO_LOGON_NO_REGISTER_SESSION = 0x0002
)
