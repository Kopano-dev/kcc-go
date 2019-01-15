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
	"bytes"
	"encoding/binary"
)

var (
	// MUIDECSAB is the GUID used in AB EntryIDs (ABEID). Definition copied
	// from kopanocore/common/include/kopano/ECGuid.h
	MUIDECSAB = DEFINE_GUID(0x50a921ac, 0xd340, 0x48ee, [8]byte{0xb3, 0x19, 0xfb, 0xa7, 0x53, 0x30, 0x44, 0x25})
)

type guidBytes struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// DEFINE_GUID is a helper to define byte representations of GUIDs.
func DEFINE_GUID(l uint32, w1, w2 uint16, b [8]byte) [16]byte {
	guid := guidBytes{
		l,
		w1,
		w2,
		b,
	}

	buf := bytes.NewBuffer(make([]byte, 0, 16))
	err := binary.Write(buf, binary.LittleEndian, guid)
	if err != nil {
		panic(err)
	}

	var res [16]byte
	copy(res[:], buf.Bytes())

	return res
}
