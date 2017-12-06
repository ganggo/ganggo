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
  "net/http"
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
  federation "gopkg.in/ganggo/federation.v0"
)

type Webfinger struct {
  *revel.Controller
}

func (c Webfinger) Webfinger() revel.Result {
  var (
    resource string
    json federation.WebfingerJson
  )

  c.Params.Bind(&resource, "resource")

  revel.Config.SetSection("ganggo")
  proto, ok := revel.Config.String("proto")
  if !ok {
    c.Response.Status = http.StatusNotFound
    c.Log.Error("No proto config found")
    return c.RenderJSON(json)
  }
  address, ok := revel.Config.String("address")
  if !ok {
    c.Response.Status = http.StatusNotFound
    c.Log.Error("No address config found")
    return c.RenderJSON(json)
  }

  username, err := federation.ParseWebfingerHandle(resource)
  if err != nil {
    c.Response.Status = http.StatusNotFound
    c.Log.Error("Cannot parse webfinger handle", "error", err)
    return c.RenderJSON(json)
  }

  db, err := models.OpenDatabase()
  if err != nil {
    c.Log.Error("Cannot open database", "error", err)
    return c.RenderError(err)
  }
  defer db.Close()

  var person models.Person
  err = db.Where("author = ?", username + "@" + address).First(&person).Error
  if err != nil {
    c.Response.Status = http.StatusNotFound
    c.Log.Error("Cannot find person", "error", err)
    return c.RenderJSON(json)
  }

  json = federation.WebfingerJson{
    Subject: "acct:" + username + "@" + address,
    Aliases: []string{proto + address + "/people/" + person.Guid},
    Links: []federation.WebfingerJsonLink{
      federation.WebfingerJsonLink {
        Rel: "http://microformats.org/profile/hcard",
        Type: "text/html",
        Href: proto + address + "/hcard/users/" + person.Guid,
      },
      federation.WebfingerJsonLink {
        Rel: "http://joindiaspora.com/seed_location",
        Type: "text/html",
        Href: proto + address + "/",
      },
      federation.WebfingerJsonLink {
        Rel: "http://webfinger.net/rel/profile-page",
        Type: "text/html",
        Href: proto + address + "/u/" + username,
      },
      federation.WebfingerJsonLink {
        Rel: "http://schemas.google.com/g/2010#updates-from",
        Type: "application/atom+xml",
        Href: proto + address + "/public/" + username + ".atom",
      },
      federation.WebfingerJsonLink {
        Rel: "salmon",
        Href: proto + address + "/receive/users/" + person.Guid,
      },
      federation.WebfingerJsonLink {
        Rel: "http://ostatus.org/schema/1.0/subscribe",
        Template: proto + address + "/people?q={uri}",
      },
      federation.WebfingerJsonLink {
        Rel: "http://openid.net/specs/connect/1.0/issuer",
        Href: proto + address,
      },
    },
  }

  return c.RenderJSON(json)
}

func (c Webfinger) HostMeta() revel.Result {
  var m federation.WebfingerXml
  revel.Config.SetSection("ganggo")
  proto, ok := revel.Config.String("proto")
  if !ok {
    c.Response.Status = http.StatusNotFound
    c.Log.Error("No proto config found")
    return c.RenderXML(m)
  }
  address, ok := revel.Config.String("address")
  if !ok {
    c.Response.Status = http.StatusNotFound
    c.Log.Error("No address config found")
    return c.RenderXML(m)
  }

  m = federation.WebfingerXml{
    Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
    Links: []federation.WebfingerXmlLink{
      federation.WebfingerXmlLink{
        Rel: "lrdd",
        Type: "application/xrd+xml",
        Template: proto + address + "/webfinger?q={uri}",
      },
    },
  }

  return c.RenderXML(m)
}
