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
  "github.com/ganggo/ganggo/app/models"
  federation "github.com/ganggo/federation"
  helpers "github.com/ganggo/federation/helpers"
  "errors"
)

type Webfinger struct {
  *revel.Controller
}

func (c Webfinger) Webfinger() revel.Result {
  var (
    legacy bool
    resource string
    webfinger federation.WebfingerData
  )

  c.Params.Bind(&resource, "resource")
  if resource == "" {
    c.Params.Bind(&resource, "q")
    legacy = true
  }

  revel.Config.SetSection("ganggo")
  proto, ok := revel.Config.String("proto")
  if !ok {
    c.Response.Status = http.StatusNotFound
    c.Log.Error("No proto config found")
    return c.RenderError(errors.New("No proto config found"))
  }
  address, ok := revel.Config.String("address")
  if !ok {
    c.Response.Status = http.StatusNotFound
    c.Log.Error("No address config found")
    return c.RenderError(errors.New("No address config found"))
  }

  username, err := helpers.ParseWebfingerHandle(resource)
  if err != nil {
    c.Response.Status = http.StatusNotFound
    c.Log.Error("Cannot parse webfinger handle", "error", err)
    return c.RenderError(err)
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
    return c.RenderError(err)
  }

  webfinger = federation.WebfingerData{
    Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
    Subject: "acct:" + username + "@" + address,
    Aliases: []string{proto + address + "/people/" + person.Guid},
    Links: []federation.WebfingerDataLink{
      federation.WebfingerDataLink {
        Rel: "http://microformats.org/profile/hcard",
        Type: "text/html",
        Href: proto + address + "/hcard/users/" + person.Guid,
      },
      federation.WebfingerDataLink {
        Rel: "http://joindiaspora.com/seed_location",
        Type: "text/html",
        Href: proto + address + "/",
      },
      federation.WebfingerDataLink {
        Rel: "http://webfinger.net/rel/profile-page",
        Type: "text/html",
        Href: proto + address + "/profiles/" + person.Guid,
      },
      federation.WebfingerDataLink {
        Rel: "self",
        Type: "application/activity+json",
        Href: proto + address + "/api/v0/activity/" + username + "/actor",
      },
      federation.WebfingerDataLink {
        Rel: "salmon",
        Href: proto + address + "/receive/users/" + person.Guid,
      },
    },
  }

  if legacy {
    return c.RenderXML(webfinger)
  }
  return c.RenderJSON(webfinger)
}

func (c Webfinger) HostMeta() revel.Result {
  var m federation.WebfingerData
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

  m = federation.WebfingerData{
    Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
    Links: []federation.WebfingerDataLink{
      federation.WebfingerDataLink{
        Rel: "lrdd",
        Type: "application/xrd+xml",
        Template: proto + address + "/.well-known/webfinger?q={uri}",
      },
    },
  }

  return c.RenderXML(m)
}
