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
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	soapUserAgent = "kcc-go-fakesoap"
	soapHeader    = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xop="http://www.w3.org/2004/08/xop/include" xmlns:xmlmime="http://www.w3.org/2004/11/xmlmime" xmlns:ns="urn:zarafa"><SOAP-ENV:Body SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">`
	soapFooter = `</SOAP-ENV:Body></SOAP-ENV:Envelope>`
)

func soapEnvelope(payload *string) *bytes.Buffer {
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

// A SoapClient is a network client which sends SOAP requests.
type SoapClient interface {
	DoRequest(ctx context.Context, payload *string, v interface{}) error
}

// A SoapHTTPClient implements a SOAP client using the HTTP protocol.
type SoapHTTPClient struct {
	Client *http.Client
	uri    string
}

// A SoapSocketClient implements a SOAP client connecting to a unix socket.
type SoapSocketClient struct {
	Dialer *net.Dialer
	path   string
}

// NewSOAPClient creates a new SOAP client for the protocol matching the
// provided URL. If the protocol is unsupported, an error is returned.
func NewSOAPClient(uri *url.URL) (SoapClient, error) {
	var err error

	if uri == nil {
		uri, err = uri.Parse(DefaultURI)
		if err != nil {
			return nil, err
		}
	}

	switch uri.Scheme {
	case "https":
		fallthrough
	case "http":
		c := &SoapHTTPClient{
			Client: DefaultHTTPClient,
			uri:    uri.String(),
		}
		return c, nil
	case "file":
		c := &SoapSocketClient{
			Dialer: DefaultUnixDialer,
			path:   uri.Path,
		}
		return c, nil

	default:
		return nil, fmt.Errorf("invalid scheme '%v' for SOAP client", uri.Scheme)
	}
}

// DoRequest sends the provided payload data as SOAP through the means of the
// accociated client.
func (sc *SoapHTTPClient) DoRequest(ctx context.Context, payload *string, v interface{}) error {
	body := soapEnvelope(payload)

	req, err := http.NewRequest(http.MethodPost, sc.uri, body)
	if err != nil {
		return err
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("User-Agent", soapUserAgent+"/"+Version)

	resp, err := sc.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http response status: %v", resp.StatusCode)
	}

	return parseSOAPResponse(resp.Body, v)
}

// DoRequest sends the provided payload data as SOAP through the means of the
// accociated client.
func (sc *SoapSocketClient) DoRequest(ctx context.Context, payload *string, v interface{}) error {
	c, err := sc.Dialer.DialContext(ctx, "unix", sc.path)
	if err != nil {
		return fmt.Errorf("failed to open unix socket: %v", err)
	}
	defer c.Close()

	body := soapEnvelope(payload)

	r := bufio.NewReader(c)

	c.SetWriteDeadline(time.Now().Add(sc.Dialer.Timeout))
	_, err = body.WriteTo(c)
	if err != nil {
		return fmt.Errorf("unexcepted unix socket write error: %v", err)
	}

	// NOTE(longsleep): Kopano SOAP socket return HTTP protocol data.
	c.SetReadDeadline(time.Now().Add(sc.Dialer.Timeout))
	resp, err := http.ReadResponse(r, nil)
	if err != nil {
		return fmt.Errorf("failed to read from unix socket: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http response status: %v", resp.StatusCode)
	}

	return parseSOAPResponse(resp.Body, v)
}
