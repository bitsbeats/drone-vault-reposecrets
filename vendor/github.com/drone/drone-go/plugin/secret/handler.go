// Copyright 2018 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package secret

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/drone/drone-go/plugin/internal/aesgcm"
	"github.com/drone/drone-go/plugin/logger"

	"github.com/99designs/httpsignatures-go"
)

// Handler returns a http.Handler that accepts JSON-encoded
// HTTP requests for a secret, invokes the underlying secret
// plugin, and writes the JSON-encoded secret to the HTTP response.
//
// The handler verifies the authenticity of the HTTP request
// using the http-signature, and returns a 400 Bad Request if
// the signature is missing or invalid.
//
// The handler can optionally encrypt the response body using
// aesgcm if the HTTP request includes the Accept-Encoding header
// set to aesgcm.
func Handler(secret string, plugin Plugin, logs logger.Logger) http.Handler {
	handler := &handler{
		secret: secret,
		plugin: plugin,
		logger: logs,
	}
	if handler.logger == nil {
		handler.logger = logger.Discard()
	}
	return handler
}

type handler struct {
	secret string
	plugin Plugin
	logger logger.Logger
}

func (p *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.logger.Debugf("received request")
	signature, err := httpsignatures.FromRequest(r)
	if err != nil {
		p.logger.Debugf("secrets: invalid or missing signature in http.Request")
		http.Error(w, "Invalid or Missing Signature", http.StatusBadRequest)
		return
	}
	if !signature.IsValid(p.secret, r) {
		p.logger.Debugf("secrets: invalid signature in http.Request")
		http.Error(w, "Invalid Signature", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		p.logger.Debugf("secrets: cannot read http.Request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := &Request{}
	err = json.Unmarshal(body, req)
	if err != nil {
		p.logger.Debugf("secrets: cannot unmarshal http.Request body")
		http.Error(w, "Invalid Input", http.StatusBadRequest)
		return
	}

	p.logger.Debugf("fetching secrets for %q", req.Name)

	secret, err := p.plugin.Find(r.Context(), req)
	if err != nil {
		p.logger.Debugf("secrets: cannot find secret %s: %s", req.Name, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	out, _ := json.Marshal(secret)

	// If the client can optionally accept an encrypted
	// response, we encrypt the payload body using secretbox.
	if r.Header.Get("Accept-Encoding") == "aesgcm" {
		key, err := aesgcm.Key(p.secret)
		if err != nil {
			p.logger.Errorf("secrets: invalid encryption key: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out, err = aesgcm.Encrypt(out, key)
		if err != nil {
			p.logger.Errorf("secrets: cannot encrypt message: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Encoding", "aesgcm")
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	w.Write(out)
}
