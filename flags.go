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
	"strconv"
)

// KCFlag is the type representing flags known to Kopano Core.
type KCFlag uint64

func (kcf KCFlag) String() string {
	return strconv.FormatUint(uint64(kcf), 10)
}

// Kopano capability flags as defined in provider/include/kcore.hpp. This only
// defines the flags actually used or understood by kcc-go.
const (
	KOPANO_CAP_LARGE_SESSIONID KCFlag = 0x0010
	KOPANO_CAP_MULTI_SERVER    KCFlag = 0x0040
	KOPANO_CAP_ENHANCED_ICS    KCFlag = 0x0100
	KOPANO_CAP_UNICODE         KCFlag = 0x0200
)

// DefaultClientCapabilities groups the default client caps sent by kcc.
var DefaultClientCapabilities = KOPANO_CAP_UNICODE |
	KOPANO_CAP_LARGE_SESSIONID |
	KOPANO_CAP_MULTI_SERVER |
	KOPANO_CAP_ENHANCED_ICS

// Kopano logon flags as defined in provider/include/kcore.hpp. This only
// defines the flags actually used or understood by kcc-go.
const (
	KOPANO_LOGON_NO_REGISTER_SESSION KCFlag = 0x0002
)
