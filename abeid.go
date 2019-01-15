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
)

// An ABEID defines an AB EntryID. See
type ABEID struct {
	header *abeidHeader
	dataV1 *abeidV1Data
}

// ABFlags returns the first byte of the associated ABEIDs abflag data.
func (abeid *ABEID) ABFlags() byte {
	return abeid.header.ABFlags[0]
}

// GUID returns the associated ABEID GUID value.
func (abeid *ABEID) GUID() [16]byte {
	return abeid.header.GUID
}

// Type returns the associated ABEID Type.
func (abeid *ABEID) Type() MAPIType {
	return abeid.dataV1.Type
}

// ID returns the associated ABEID ID numeric field value.
func (abeid *ABEID) ID() uint32 {
	return abeid.dataV1.ID
}

// ExID returns the associated ABEID external ID field byte value.
func (abeid *ABEID) ExID() ([]byte, error) {
	value := unpadExID(abeid.dataV1.ExID[:])
	extIDBytes := make([]byte, base64.StdEncoding.DecodedLen(len(value)))
	_, err := base64.StdEncoding.Decode(extIDBytes, value)
	if err != nil {
		return nil, fmt.Errorf("error decoding ExID: %v", err)
	}

	return extIDBytes, nil
}

func unpadExID(value []byte) []byte {
	return bytes.TrimRightFunc(value, func(r rune) bool {
		return r == '\x00'
	})
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
	ExID [8]byte
}

// NewABEIDFromBytes takes a byte value and returns the ABEID represented by
// those bytes.
func NewABEIDFromBytes(value []byte) (*ABEID, error) {
	reader := bytes.NewReader(value)

	var header abeidHeader
	err := binary.Read(reader, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}
	if header.Version != 1 {
		return nil, fmt.Errorf("ABEID unsupported version %d", header.Version)
	}

	var data abeidV1Data
	err = binary.Read(reader, binary.LittleEndian, &data)
	if err != nil {
		return nil, err
	}

	return &ABEID{&header, &data}, nil
}

// NewABEIDFromHex takes a hex encoded byte value and returns the ABEID
// represented by those bytes.
func NewABEIDFromHex(hexValue []byte) (*ABEID, error) {
	value := make([]byte, hex.DecodedLen(len(hexValue)))

	if _, err := hex.Decode(value, hexValue); err != nil {
		return nil, err
	}

	return NewABEIDFromBytes(value)
}

// NewABEIDFromBase64 takes a base64Std encoded byte value and returns the ABEID
// represented by those bytes.
func NewABEIDFromBase64(base64Value []byte) (*ABEID, error) {
	value := make([]byte, base64.StdEncoding.DecodedLen(len(base64Value)))

	if _, err := base64.StdEncoding.Decode(value, base64Value); err != nil {
		return nil, err
	}

	return NewABEIDFromBytes(value)
}
