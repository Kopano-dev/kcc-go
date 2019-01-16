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
	"testing"
)

func compareV1(t *testing.T, idx int, abeid *abeidV1, typE MAPIType, ID uint32, exID []byte) {
	if abeid.header.ABFlags != [4]byte{0, 0, 0, 0} {
		t.Errorf("ABEID(%d) unexpected ABFlags header value: %v", idx, abeid.header.ABFlags)
	}
	if abeid.header.GUID != MUIDECSAB {
		t.Errorf("ABEID(%d) unexpected GUID header value: %v", idx, abeid.header.GUID)
	}
	if abeid.header.Version != 1 {
		t.Errorf("ABEID(%d) unexpected Version header value: %v", idx, abeid.header.Version)
	}
	if abeid.dataV1.Type != typE {
		t.Errorf("ABEID(%d) unexpected Type data value: %v", idx, abeid.dataV1.Type)
	}
	if abeid.dataV1.ID != ID {
		t.Errorf("ABEID(%d) unexpected ID data value: %v", idx, abeid.dataV1.ID)
	}

	if abeid.ABFlags() != abeid.header.ABFlags[0] {
		t.Errorf("ABEID(%d) ABFlags value mismatch", idx)
	}
	if abeid.GUID() != abeid.header.GUID {
		t.Errorf("ABEID(%d) GUID value mismatch", idx)
	}
	if abeid.Type() != abeid.dataV1.Type {
		t.Errorf("ABEID(%d) Type value mismatch", idx)
	}
	if abeid.ID() != abeid.dataV1.ID {
		t.Errorf("ABEID(%d) ID value mismatch", idx)
	}

	exIDBytesWanted := make([]byte, base64.StdEncoding.DecodedLen(len(exID)))
	n, err := base64.StdEncoding.Decode(exIDBytesWanted, exID[:])
	if err != nil {
		t.Errorf("ABEID(%d) invalid ExID value in test: %v", idx, err)
		return
	}
	exIDBytes := abeid.ExID()
	if err != nil {
		t.Errorf("ABEID(%d error decoding ExID value: %v", idx, err)
		return
	}
	if !bytes.Equal(exIDBytesWanted[:n], exIDBytes) {
		t.Errorf("ABEID(%d) ExID value mismatch: %v, wanted %v", idx, exIDBytes, exIDBytesWanted[:n])
	}
}

func TestABEIDFromHex(t *testing.T) {
	values := [][]byte{
		[]byte("00000000ac21a95040d3ee48b319fba7533044250100000006000000040000004d673d3d00000000"),
		[]byte("00000000AC21A95040D3EE48B319FBA7533044250100000006000000450000004F4441774D673D3D00000000"),
		[]byte("00000000AC21A95040D3EE48B319FBA7533044250100000006000000450000004F4441774D673D3D"),
	}
	types := []MAPIType{
		MAPI_MAILUSER,
		MAPI_MAILUSER,
		MAPI_MAILUSER,
	}
	ids := []uint32{
		4,
		69,
		69,
	}
	exIDs := [][]byte{
		[]byte{77, 103, 61, 61},
		[]byte{79, 68, 65, 119, 77, 103, 61, 61},
		[]byte{79, 68, 65, 119, 77, 103, 61, 61},
	}

	for idx, value := range values {
		abeid, err := NewABEIDFromHex(value)
		if err != nil {
			t.Error(err)
			continue
		}
		compareV1(t, idx, abeid.(*abeidV1), types[idx], ids[idx], exIDs[idx])
	}
}

func TestABEIDFromBase64(t *testing.T) {
	values := [][]byte{
		[]byte("AAAAAKwhqVBA0+5Isxn7p1MwRCUBAAAABgAAAAMAAABNZz09AAAAAA=="),
		[]byte("AAAAAKwhqVBA0+5Isxn7p1MwRCUBAAAABgAAAAMAAABNZz09"), // No padding.
	}
	types := []MAPIType{
		MAPI_MAILUSER,
		MAPI_MAILUSER,
	}
	ids := []uint32{
		3,
		3,
	}
	exIDs := [][]byte{
		[]byte{77, 103, 61, 61},
		[]byte{77, 103, 61, 61},
	}

	for idx, value := range values {
		abeid, err := NewABEIDFromBase64(value)
		if err != nil {
			t.Error(err)
			continue
		}
		compareV1(t, idx, abeid.(*abeidV1), types[idx], ids[idx], exIDs[idx])
	}
}

func TestDefineGUID(t *testing.T) {
	muidecsab := DEFINE_GUID(0x50a921ac, 0xd340, 0x48ee, [8]byte{0xb3, 0x19, 0xfb, 0xa7, 0x53, 0x30, 0x44, 0x25})

	if muidecsab != [16]byte{172, 33, 169, 80, 64, 211, 238, 72, 179, 25, 251, 167, 83, 48, 68, 37} {
		t.Errorf("DEFINE_GUID created unexpected value: %v", muidecsab)
	}
}

func TestABEIDEqual(t *testing.T) {
	a, _ := NewABEIDFromHex([]byte("00000000ac21a95040d3ee48b319fba7533044250100000006000000040000004d673d3d00000000"))
	b, _ := NewABEIDFromHex([]byte("00000000ac21a95040d3ee48b319fba7533044250100000006000000040000004d673d3d00000000"))
	c, _ := NewABEIDFromHex([]byte("00000000ac21a95040d3ee48b319fba7533044250100000006000000050000004d673d3d00000000"))
	d, _ := NewABEIDFromHex([]byte("00000000ac21a95040d3ee48b319fba7533044250100000006000000040000004d674d3d00000000"))
	e, _ := NewABEIDFromHex([]byte("00000000ac21a95040d3ee48b319fba7533044250100000006000000040000004d673d3d")) // No padding.

	if !ABEIDEqual(a, b) {
		t.Error("ABEID compare mismatch a and b")
	}
	if !ABEIDEqual(a, c) {
		t.Error("ABEID compare mismatch a and c")
	}
	if ABEIDEqual(a, d) {
		t.Error("ABEID compare match a and d while it should not match")
	}
	if !ABEIDEqual(a, e) {
		t.Error("ABEID compare mismatch a and e")
	}
}

func TestABEIDV1String(t *testing.T) {
	value := "AAAAAKwhqVBA0+5Isxn7p1MwRCUBAAAABgAAAAMAAABNZz09"

	a, _ := NewABEIDFromBase64([]byte(value))
	s := a.String()

	if s != value {
		t.Errorf("ABEID string value mismatch got %v, wanted %v", s, value)
	}
}

func TestABEIDV1Hex(t *testing.T) {
	value := "00000000ac21a95040d3ee48b319fba7533044250100000006000000040000004d673d3d"

	a, _ := NewABEIDFromHex([]byte(value))
	h := a.Hex()

	if h != value {
		t.Errorf("ABEID string value mismatch got %v, wanted %v", h, value)
	}
}

func TestNewABEIDV1(t *testing.T) {
	a, err := NewABEIDV1(MUIDECSAB, MAPI_MAILUSER, 0, []byte{1, 2, 3, 4})
	if err != nil {
		t.Fatalf("NewABEIDV1 failed with error: %v", err)
	}
	b, _ := NewABEIDFromBase64([]byte("AAAAAKwhqVBA0+5Isxn7p1MwRCUBAAAABgAAAAAAAABBUUlEQkE9PQ=="))
	if !ABEIDEqual(a, b) {
		t.Error("ABEID compare mismatch a and b", a.String(), b.String())
	}
}
