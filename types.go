/*
 * Copyright 2019 Kopano and its licensors
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

// MAPIType is the type representing MAPI types as used by Kopano Core.
type MAPIType uint32

func (mt MAPIType) String() string {
	return strconv.FormatUint(uint64(mt), 10)
}

// Possibe type values as defined in mapi4linux/include/mapidefs.h. We
// only define the ones know and understood by kcc-go.
const (
	MAPI_MAILUSER MAPIType = 0x00000006
)
