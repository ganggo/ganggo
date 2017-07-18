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
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Webfinger struct {
  *revel.Controller
}

func (c Webfinger) Webfinger() revel.Result {
  var(
    q string
    m federation.WebfingerXml
  )

  c.Params.Bind(&q, "q")
  revel.Config.SetSection("ganggo")
  proto, ok := revel.Config.String("proto")
  if !ok {
    c.Response.Status = http.StatusNotFound
    revel.TRACE.Println("no proto config found")
    return c.RenderXML(m)
  }
  address, ok := revel.Config.String("address")
  if !ok {
    c.Response.Status = http.StatusNotFound
    revel.TRACE.Println("no address config found")
    return c.RenderXML(m)
  }

  username, err := federation.ParseWebfingerHandle(q)
  if err != nil {
    c.Response.Status = http.StatusNotFound
    revel.TRACE.Println(err)
    return c.RenderXML(m)
  }

  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.WARN.Println(err)
    return c.Render()
  }
  defer db.Close()

  var person models.Person
  err = db.Where("author = ?", username + "@" + address).First(&person).Error
  if err != nil {
    c.Response.Status = http.StatusNotFound
    revel.TRACE.Println(err)
    return c.RenderXML(m)
  }

  m = federation.WebfingerXml{
    Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
    Subject: "acct:" + username + "@" + address,
    Links: []federation.WebfingerXmlLink{
      federation.WebfingerXmlLink {
        Rel: "http://microformats.org/profile/hcard",
        Type: "text/html",
        Href: proto + address + "/hcard/users/" + person.Guid,
      },
      federation.WebfingerXmlLink {
        Rel: "http://joindiaspora.com/seed_location",
        Type: "text/html",
        Href: proto + address + "/",
      },
    },
  }

  return c.RenderXML(m)
}

func (c Webfinger) HostMeta() revel.Result {
  var m federation.WebfingerXml
  revel.Config.SetSection("ganggo")
  proto, ok := revel.Config.String("proto")
  if !ok {
    c.Response.Status = http.StatusNotFound
    revel.TRACE.Println("no proto config found")
    return c.RenderXML(m)
  }
  address, ok := revel.Config.String("address")
  if !ok {
    c.Response.Status = http.StatusNotFound
    revel.TRACE.Println("no address config found")
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
