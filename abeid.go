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
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
)

type ABEID interface {
	ABFlags() byte
	GUID() [16]byte
	Type() MAPIType
	ID() uint32
	ExID() []byte
	String() string
	Hex() string
}

// An abeidV1 defines an AB EntryID. See
type abeidV1 struct {
	header *abeidHeader
	dataV1 *abeidV1Data
	exID   []byte
}

// ABFlags returns the first byte of the associated ABEIDs abflag data.
func (abeid *abeidV1) ABFlags() byte {
	return abeid.header.ABFlags[0]
}

// GUID returns the associated ABEID GUID value.
func (abeid *abeidV1) GUID() [16]byte {
	return abeid.header.GUID
}

// Type returns the associated ABEID Type.
func (abeid *abeidV1) Type() MAPIType {
	return abeid.dataV1.Type
}

// ID returns the associated ABEID ID numeric field value.
func (abeid *abeidV1) ID() uint32 {
	return abeid.dataV1.ID
}

// ExID returns the associated ABEID external ID field as byte value.
func (abeid *abeidV1) ExID() []byte {
	return abeid.exID
}

func (abeid *abeidV1) String() string {
	buf := new(bytes.Buffer)
	enc1 := base64.NewEncoder(base64.StdEncoding, buf)
	binary.Write(enc1, binary.LittleEndian, abeid.header)
	binary.Write(enc1, binary.LittleEndian, abeid.dataV1)
	enc2 := base64.NewEncoder(base64.StdEncoding, enc1)
	enc2.Write(abeid.exID)
	enc2.Close()
	enc1.Close()

	return buf.String()
}

func (abeid *abeidV1) Hex() string {
	buf := new(bytes.Buffer)
	enc1 := hex.NewEncoder(buf)
	binary.Write(enc1, binary.LittleEndian, abeid.header)
	binary.Write(enc1, binary.LittleEndian, abeid.dataV1)
	enc2 := base64.NewEncoder(base64.StdEncoding, enc1)
	enc2.Write(abeid.exID)
	enc2.Close()

	return buf.String()
}

// A abeidHeader is the byte representation of an AB EntryID start including
// version. See
// https://docs.microsoft.com/en-us/office/client-developer/outlook/mapi/entryid
// for the basic EntryID defintion.
type abeidHeader struct {
	ABFlags [4]byte
	GUID    [16]byte
	Version uint32
}

// abeidV1Data define further values as defined in provider/include/kcore.hpp
// for version 1 ABEID structs.
type abeidV1Data struct {
	Type MAPIType
	ID   uint32
	/* Rest is exID of arbitrary size */
}

// NewABEIDFromBytes takes a byte value and returns the ABEID represented by
// those bytes.
func NewABEIDFromBytes(value []byte) (ABEID, error) {
	reader := bytes.NewReader(value)

	// Parse header into header struct.
	var header abeidHeader
	err := binary.Read(reader, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}

	var abeid ABEID
	switch header.Version {
	case 1:
		// Parse fixed size V1 data into data struct.
		var data abeidV1Data
		err = binary.Read(reader, binary.LittleEndian, &data)
		if err != nil {
			break
		}
		// Read all the rest.
		exIDRaw, err := ioutil.ReadAll(reader)
		if err != nil {
			break
		}
		// Remove padding.
		exIDRaw = unpadBytesRightWithRune(exIDRaw, '\x00')
		// Decode.
		exID := make([]byte, base64.StdEncoding.DecodedLen(len(exIDRaw)))
		n, err := base64.StdEncoding.Decode(exID, exIDRaw)
		if err != nil {
			break
		}
		// Construct with all the data.
		abeid = &abeidV1{
			header: &header,
			dataV1: &data,
			exID:   exID[:n],
		}

	default:
		err = fmt.Errorf("ABEID unsupported version %d", header.Version)
	}

	if err != nil {
		return nil, err
	}
	return abeid, nil
}

// NewABEIDFromHex takes a hex encoded byte value and returns the ABEID
// represented by those bytes.
func NewABEIDFromHex(hexValue []byte) (ABEID, error) {
	value := make([]byte, hex.DecodedLen(len(hexValue)))

	if _, err := hex.Decode(value, hexValue); err != nil {
		return nil, err
	}

	return NewABEIDFromBytes(value)
}

// NewABEIDFromBase64 takes a base64Std encoded byte value and returns the ABEID
// represented by those bytes.
func NewABEIDFromBase64(base64Value []byte) (ABEID, error) {
	value := make([]byte, base64.StdEncoding.DecodedLen(len(base64Value)))

	if _, err := base64.StdEncoding.Decode(value, base64Value); err != nil {
		return nil, err
	}

	return NewABEIDFromBytes(value)
}

// NewABEIDV1 creates a new NewABEIDV1 from the provided values.
func NewABEIDV1(guid [16]byte, typE MAPIType, id uint32, exID []byte) (ABEID, error) {
	abeid := &abeidV1{
		header: &abeidHeader{
			GUID:    guid,
			Version: 1,
		},
		dataV1: &abeidV1Data{
			Type: typE,
			ID:   id,
		},
		exID: exID,
	}

	return abeid, nil
}

// ABEIDEqual returns true if the provided btwo ABEID refer to the same entry
// considering all relevant fields, ignoring the not relevant (like ID).
func ABEIDEqual(first, second ABEID) bool {
	switch a := first.(type) {
	case *abeidV1:
		b, ok := second.(*abeidV1)
		if !ok {
			return false
		}

		if a.header == nil || b.header == nil {
			return false
		}
		if a.dataV1 == nil || b.dataV1 == nil {
			return false
		}
		if a.header.Version != a.header.Version {
			return false
		}
		if a.header.GUID != b.header.GUID {
			return false
		}
		if a.dataV1.Type != b.dataV1.Type {
			return false
		}

		return bytes.Equal(a.exID, b.exID)
	}

	return false
}

func unpadBytesRightWithRune(value []byte, p rune) []byte {
	return bytes.TrimRightFunc(value, func(r rune) bool {
		return r == p
	})
}
