package controllers
//
// GangGo Application Server
// Copyright (C) 2017 Lukas Matt <lukas@zauberstuhl.de>
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
  "github.com/revel/revel"
  "strings"
)

type SocialRelaySchema struct {
  Subscribe bool `json:"subscribe"`
  Scope string `json:"scope"`
  Tags []string `json:"tags"`
}

type SocialRelay struct {
  *revel.Controller
}

func (s SocialRelay) Index() revel.Result {
  revel.Config.SetSection("ganggo")
  subscribe := revel.Config.BoolDefault("relay.subscribe", false)
  scope := revel.Config.StringDefault("relay.scope", "all")
  tags := strings.Split(revel.Config.StringDefault("relay.tags", ""), ",")

  return s.RenderJSON(SocialRelaySchema{
    Subscribe: subscribe,
    Scope: scope,
    Tags: tags,
  })
}
