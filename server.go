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
	"crypto/tls"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	httpServer  http.Server
	httpsServer http.Server
	HttpPort    int
	HttpsPort   int
	CertsDir    string
	handler     http.Handler
}

func NewServer(httpPort, httpsPort int, certs string, handler http.Handler) *Server {
	s := &Server{
		HttpPort:  httpPort,
		HttpsPort: httpsPort,
		CertsDir:  certs,
		handler:   handler,
	}
	if httpPort != 0 {
		s.httpServer = http.Server{
			Addr:           ":" + strconv.Itoa(httpPort),
			Handler:        handler,
			ReadTimeout:    60 * time.Second,
			WriteTimeout:   60 * time.Second,
			MaxHeaderBytes: 4096,
		}
	}
	if httpsPort != 0 {
		s.httpServer = http.Server{
			Addr:           ":" + strconv.Itoa(httpPort),
			Handler:        handler,
			ReadTimeout:    60 * time.Second,
			WriteTimeout:   60 * time.Second,
			MaxHeaderBytes: 4096,
		}
	}
	return s
}

func (s *Server) Serve() {
	if s.HttpPort == 0 && s.HttpsPort == 0 {
		log.Error("Both ports are disabled, cowardly refusing to serve nothing forever")
	}
	if s.HttpPort != 0 {
		go func() {
			err := s.serveHttp()
			if err != nil {
				log.WithError(err).Error("Starting http listener")
			}
		}()
	}
	if s.HttpsPort != 0 {
		go func() {
			err := s.serveHttps()
			if err != nil {
				log.WithError(err).Error("Starting https listener")
			}
		}()
	}
}

func (s *Server) serveHttp() error {
	return s.httpServer.ListenAndServe()

}

func (s *Server) serveHttps() error {
	// Load SNI Certificated
	file, err := os.OpenFile(s.CertsDir, os.O_RDONLY, os.ModeDir)
	if err != nil {
		return err
	}
	config := &tls.Config{}
	if s.httpsServer.TLSConfig != nil {
		config = s.httpServer.TLSConfig
	} else {
		s.httpServer.TLSConfig = config
	}
	files, err := file.Readdir(-1)
	if err != nil {
		return err
	}
	var i int = 0
	config.Certificates = make([]tls.Certificate, 0, 20)
	for i < len(files) {
		fileInfo := files[i]
		if !fileInfo.IsDir() {
			if strings.HasSuffix(fileInfo.Name(), "key") {
				// We have the key, find the corresponding cert !
				certFile := path.Join(s.CertsDir, strings.TrimSuffix(fileInfo.Name(), "key")+"crt")
				keyFile := path.Join(s.CertsDir, fileInfo.Name())
				cert, err := tls.LoadX509KeyPair(certFile, keyFile)
				if err != nil {
					log.WithFields(log.Fields{"error": err.Error(), "certFile": certFile, "keyFile": keyFile}).Error("Reading certificate")
				} else {
					config.Certificates = append(config.Certificates, cert)
				}
			}
		}
		i++
	}
	config.BuildNameToCertificate()
	return s.httpsServer.ListenAndServe()
}
