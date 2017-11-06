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
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

const (
	soapUserAgent = "kcc-go-fakesoap"
	soapHeader    = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xop="http://www.w3.org/2004/08/xop/include" xmlns:xmlmime="http://www.w3.org/2004/11/xmlmime" xmlns:ns="urn:zarafa"><SOAP-ENV:Body SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">`
	soapFooter = `</SOAP-ENV:Body></SOAP-ENV:Envelope>`
)

func soapEnvelope(payload *string) io.Reader {
	var b bytes.Buffer
	b.WriteString(soapHeader)
	b.WriteString(*payload)
	b.WriteString(soapFooter)
	return &b
}

func newSOAPRequest(ctx context.Context, url string, payload *string) (*http.Request, error) {
	body := soapEnvelope(payload)

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("User-Agent", soapUserAgent+"/"+Version)

	return req, nil
}

func parseSOAPResponse(data io.Reader, v interface{}) error {
	decoder := xml.NewDecoder(data)

	match := false
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if match {
				return decoder.DecodeElement(v, &se)
			}

			if se.Name.Local == "Body" {
				match = true
			}
		}
	}

	return fmt.Errorf("failed to unmarshal SOAP response body")
}
