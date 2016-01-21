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
	log "github.com/Sirupsen/logrus"
	"github.com/neuneu2k/hyenad"
	"io"
	"net/http"
	"sync"
	"time"
)

type GateHandler struct {
	client   hyenad.HyenaClient
	pendings map[hyenad.MsgId]PendingResponse
	lock     sync.Mutex
}

const Timeout = 40
const GcTimer = 5 * time.Second

func (g *GateHandler) gc() {
	ticks := time.NewTicker(GcTimer).C
	for tick := range ticks {
		g.lock.Lock()
		toDelete := make([]hyenad.MsgId, 0, 16)
		for id, pending := range g.pendings {
			age := tick.Sub(pending.Start)
			if age.Seconds() > Timeout {
				log.WithField("PendingRequest", pending).Debug("Timeout")
				toDelete = append(toDelete, id)
				close(pending.Response)
			}
		}
		for _, id := range toDelete {
			delete(g.pendings, id)
		}
		g.lock.Unlock()
	}
}
func (g *GateHandler) OnStream(stream hyenad.ReadStream) {
	response, err := ReadHttpResponse(&stream)
	if err == nil {
		id := response.CorrelationId
		g.lock.Lock()
		defer g.lock.Unlock()
		log.WithField("Id", id).Debug("Searching for pending request")
		pending, ok := g.pendings[id]
		if ok {
			log.WithField("Id", id).Debug("Found pending request, transferring response stream")
			pending.Response <- &response
			delete(g.pendings, id)
		} else {
			log.WithField("Id", id).Error("Pending request not found")
			// Read stream to null
			io.Copy(&NullWriter{}, &stream)
		}
	} else {
		log.WithError(err).Error("Decoding response")
	}
}

type NullWriter struct{}

func (n *NullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func CreateGateHandler() (*GateHandler, error) {
	res := GateHandler{}
	var err error
	res.client, err = hyenad.NewHyenaClient(1, &res)
	res.pendings = make(map[hyenad.MsgId]PendingResponse)
	go res.gc()
	return &res, err
}

func (g *GateHandler) CreateResponse(id hyenad.MsgId) PendingResponse {
	res := PendingResponse{
		Id:       id,
		Start:    time.Now(),
		Response: make(chan *HttpResponse),
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	log.WithField("Id", id).Debug("Creating pending request")
	g.pendings[id] = res
	return res
}

type PendingResponse struct {
	Id       hyenad.MsgId
	Start    time.Time
	Response chan *HttpResponse
}

func (g *GateHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	logger := log.WithField("Url", req.URL)
	hyreq := ToHyenaRequest(req)
	uri := req.RequestURI
	if uri == "" || uri[0] != '/' {
		uri = "/" + uri
	}
	stream := g.client.CreateStream("h:" + uri)
	hyPendingRes := g.CreateResponse(stream.Id)
	go hyreq.WriteHttpRequest(stream)
	hyres := <-hyPendingRes.Response
	logger = logger.WithField("Response", hyres)
	if hyres == nil {
		//TODO: Have a nice gateway timeout responder
		http.Error(res, "Gateway Timeout", 504)
	} else {
		WriteResponse(hyres, res)
	}
}

func WriteResponse(h *HttpResponse, res http.ResponseWriter) {
	res.WriteHeader(int(h.StatusCode))
	for code, value := range h.Headers {
		res.Header().Add(Headers[code].Name, value)
	}
	io.Copy(res, h.ContentStream)
}

func ToHyenaRequest(req *http.Request) HttpRequest {
	res := HttpRequest{}
	switch req.Method {
	case "GET":
		{
			res.Method = GET
		}
	case "HEAD":
		{
			res.Method = HEAD
		}
	case "POST":
		{
			res.Method = POST
		}
	case "PUT":
		{
			res.Method = PUT
		}
	case "DELETE":
		{
			res.Method = DELETE
		}
	default:
		{
			res.Method = GET
		}
	}
	res.Headers = make(map[byte]string)
	for name, values := range req.Header {
		for code, header := range Headers {
			if header.Name == name {
				res.Headers[code] = values[0]
			}
		}
	}
	res.Headers[XForwardedHost.Code] = req.RemoteAddr
	if req.TLS != nil {
		res.Headers[XForwardedProto.Code] = "https"
	} else {
		res.Headers[XForwardedProto.Code] = "http"
	}
	res.ContentStream = req.Body
	return res
}
