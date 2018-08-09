package main
//
// GangGo Application Server
// Copyright (C) 2018 Lukas Matt <lukas@zauberstuhl.de>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

import (
  "fmt"
  "io"
  "net/http"
  "time"
  "runtime"
  "encoding/json"
)

const (
  BINTRAY_API = "https://api.bintray.com/packages/%s/%s/%s/versions/_latest"
  BINTRAY_DL = "https://dl.bintray.com/%s/%s/updater.%s-%s.%s.bin"
)

type Bintray struct {
  User, Repo, Pkg string
  Interval time.Duration
  // internal
  delay bool
  last time.Time
}

type ApiResponse struct {
  Name string
  Created time.Time
}

// Init validates the provided config
func (h *Bintray) Init() error {
  //apply defaults
  if h.User == "" || h.Repo == "" || h.Pkg == "" {
    return fmt.Errorf("User, Repo and Pkg required")
  }
  if h.Interval == 0 {
    h.Interval = 5 * time.Minute
  }
  return nil
}

// Fetch the binary from the provided URL
func (h *Bintray) Fetch() (io.Reader, error) {
  // delay fetches after first
  if h.delay {
    time.Sleep(h.Interval)
  }
  h.delay = true

  // check version and return if no update is necessary
  // https://api.bintray.com/packages/:user/:repo/:pkg/versions/_latest
  resp, err := http.Get(fmt.Sprintf(BINTRAY_API, h.User, h.Repo, h.Pkg))
  if err != nil {
    return nil, fmt.Errorf("GET request failed (%s)", err)
  }
  if resp.StatusCode != http.StatusOK {
    return nil, fmt.Errorf("GET request failed (status code %d)", resp.StatusCode)
  }
  apiResp := ApiResponse{}
  decoder := json.NewDecoder(resp.Body)
  err = decoder.Decode(&apiResp)
  if err != nil {
    return nil, fmt.Errorf("DECODER failed (%s)", err)
  }
  if apiResp.Created.Before(h.last) {
    return nil, nil
  }
  h.last = apiResp.Created

  // binary fetch using GET
  resp, err = http.Get(fmt.Sprintf(BINTRAY_DL, h.User, h.Repo,
    runtime.GOOS, runtime.GOARCH, apiResp.Name))
  if err != nil {
    return nil, fmt.Errorf("GET request failed (%s)", err)
  }
  if resp.StatusCode != http.StatusOK {
    return nil, fmt.Errorf("GET request failed (status code %d)", resp.StatusCode)
  }
  return resp.Body, nil
}
