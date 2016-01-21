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
	"bytes"
	"github.com/neuneu2k/hyenad"
	"io"
	"testing"
)

func TestReqSerDes(t *testing.T) {
	request := HttpRequest{}
	request.Headers = make(map[byte]string)
	request.Headers[Authorization.Code] = "bearer: GotThePower"
	request.Headers[UserAgent.Code] = "Dummy user agent, testing yeah"
	request.Method = GET
	content := bytes.NewBufferString("This is my content, there are many like it, but this one is mine")
	request.ContentStream = content
	buf := bytes.Buffer{}
	t.Logf("Initial request : %v %v", request.Method, request.Headers)
	request.WriteHttpRequest(NopCloser(&buf))
	out := buf.Bytes()
	t.Logf("Encoded request (%v bytes)", len(out))
	r := bytes.NewReader(out)
	req2, err := ReadHttpRequest(r)
	if err != nil {
		t.Errorf("Error decoding request: %v", err)
	} else {
		t.Logf("Decoded request : %v", req2)
		content := make([]byte, 200)
		n, _ := req2.ContentStream.Read(content)
		t.Logf("Decoded content : %v", string(content[0:n]))
	}
}

func TestResSerDes(t *testing.T) {
	response := HttpResponse{}
	response.Headers = make(map[byte]string)
	response.Headers[ContentType.Code] = "text/html"
	response.Headers[Age.Code] = "3600"
	response.StatusCode = 200
	response.CorrelationId = hyenad.CreateMid(0, 1, 20)
	content := bytes.NewBufferString("This is my content, there are many like it, but this one is mine")
	response.ContentStream = content
	buf := bytes.Buffer{}
	t.Logf("Initial response : %v:%v:%v", response.CorrelationId, response.StatusCode, response.Headers)
	response.WriteHttpResponse(NopCloser(&buf))
	out := buf.Bytes()
	t.Logf("Encoded response  (%v bytes)", len(out))
	r := bytes.NewReader(out)
	res2, err := ReadHttpResponse(r)
	if err != nil {
		t.Errorf("Error decoding response: %v", err)
	} else {
		t.Logf("Decoded response : %v", res2)
		content := make([]byte, 200)
		n, _ := res2.ContentStream.Read(content)
		t.Logf("Decoded content : %v", string(content[0:n]))
	}
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error {
	return nil
}

func NopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}
