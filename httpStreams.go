/*
Copyright 2016 Assoba S.A.S.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package weir2

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/neuneu2k/hyenad"
	"io"
)

type Method byte

const (
	GET Method = iota
	HEAD
	POST
	PUT
	DELETE
)

type Header struct {
	Code byte
	Name string
}

var (
	UserAgent               = Header{0, "User-Agent"}
	IfModifiedSince         = Header{1, "If-Modified-Since"}
	IfNoneMatch             = Header{2, "If-None-Match"}
	IfMatch                 = Header{3, "If-Match"}
	Range                   = Header{4, "Range"}
	Referer                 = Header{5, "Referer"}
	Via                     = Header{6, "Via"}
	XForwardedHost          = Header{7, "X-Forwarded-Host"}
	XForwardedProto         = Header{8, "X-Forwarded-Proto"}
	Authorization           = Header{9, "Authorization"}
	CacheControl            = Header{10, "Cache-Control"}
	Age                     = Header{11, "Age"}
	ContentType             = Header{12, "Content-Type"}
	ContentLength           = Header{13, "Content-Length"}
	ContentEncoding         = Header{14, "Content-Encoding"}
	ContentRange            = Header{15, "Content-Range"}
	ETag                    = Header{16, "ETag"}
	Expires                 = Header{17, "Expires"}
	LastModified            = Header{18, "Last-Modified"}
	Location                = Header{19, "Location"}
	PublicKeyPins           = Header{20, "Public-Key-Pins"}
	StrictTransportSecurity = Header{21, "Strict-Transport-Security"}
	XFrameOptions           = Header{22, "X-Frame-Options"}
	XContentSecurityPolicy  = Header{23, "X-Content-Security-Policy"}
)

var headersList = []Header{UserAgent, IfModifiedSince, IfNoneMatch, IfMatch, Range,
	Referer, Via, XForwardedHost, XForwardedProto, Authorization, CacheControl,
	Age, ContentType, ContentLength, ContentEncoding, ContentRange, ETag, Expires,
	LastModified, Location, PublicKeyPins, StrictTransportSecurity, XFrameOptions, XContentSecurityPolicy}

var Headers map[byte]Header

func init() {
	Headers = make(map[byte]Header)
	for _, h := range headersList {
		Headers[h.Code] = h
	}
}

type HttpRequest struct {
	Method        Method          `json:"m"`
	Headers       map[byte]string `json:"h"`
	ContentStream io.Reader       `json:"-"`
}

type HttpResponse struct {
	CorrelationId hyenad.MsgId    `json:"c"`
	StatusCode    uint16          `json:"s"`
	Headers       map[byte]string `json:"h"`
	ContentStream io.Reader       `json:"-"`
}

func ReadHttpRequest(r io.Reader) (HttpRequest, error) {
	res := HttpRequest{}
	res.Headers = make(map[byte]string)
	b := []byte{0}
	r.Read(b)
	res.Method = Method(b[0])
	r.Read(b)
	numHeaders := b[0]
	for i := byte(0); i < numHeaders; i++ {
		r.Read(b)
		code := b[0]
		r.Read(b)
		strLen := b[0]
		str := make([]byte, int(strLen))
		r.Read(str)
		value := string(str)
		res.Headers[code] = value
	}
	res.ContentStream = r
	return res, nil
}

func (h *HttpRequest) WriteHttpRequest(w io.WriteCloser) error {
	w.Write([]byte{byte(h.Method)})
	numHeaders := len(h.Headers)
	if numHeaders > 255 {
		return fmt.Errorf("Too many headers: %v (max 255 headers)", numHeaders)
	}
	w.Write([]byte{byte(numHeaders)})
	for k, v := range h.Headers {
		w.Write([]byte{k})
		strLen := len(v)
		w.Write([]byte{byte(strLen)})
		w.Write([]byte(v))
	}
	_, err := io.Copy(w, h.ContentStream)
	w.Close()
	return err
}

func ReadHttpResponse(r io.Reader) (HttpResponse, error) {
	res := HttpResponse{}
	res.Headers = make(map[byte]string)
	status := make([]byte, 2)
	_, err := r.Read(status)
	if err != nil {
		return res, err
	}
	res.StatusCode = uint16((int(status[0]) << 8) + int(status[1]))
	log.WithField("Status", res.StatusCode).Debug("Read status")
	msgId := make([]byte, 16)
	_, err = r.Read(msgId)
	if err != nil {
		return res, err
	}
	res.CorrelationId.ReadFrom(msgId)
	log.WithField("CorrelationId", res.CorrelationId).Debug("Read CID")
	b := []byte{0}
	_, err = r.Read(b)
	if err != nil {
		return res, err
	}
	numHeaders := b[0]
	log.WithField("HeadersSize", numHeaders).Debug("Read Size of Headers")
	for i := byte(0); i < numHeaders; i++ {
		_, err = r.Read(b)
		if err != nil {
			return res, err
		}
		code := b[0]
		_, err = r.Read(b)
		if err != nil {
			return res, err
		}
		strLen := b[0]
		str := make([]byte, int(strLen))
		_, err = r.Read(str)
		if err != nil {
			return res, err
		}
		value := string(str)
		res.Headers[code] = value
	}
	res.ContentStream = r
	return res, nil
}

func (h *HttpResponse) WriteHttpResponse(w io.WriteCloser) error {
	status := []byte{byte(h.StatusCode >> 8 & 0xFF), byte(h.StatusCode & 0xFF)}
	w.Write(status)
	w.Write(h.CorrelationId.Bytes())
	numHeaders := len(h.Headers)
	if numHeaders > 255 {
		return fmt.Errorf("Too many headers: %v (max 255 headers)", numHeaders)
	}
	w.Write([]byte{byte(numHeaders)})
	for k, v := range h.Headers {
		w.Write([]byte{k})
		strLen := len(v)
		w.Write([]byte{byte(strLen)})
		w.Write([]byte(v))
	}
	_, err := io.Copy(w, h.ContentStream)
	w.Close()
	return err
}
