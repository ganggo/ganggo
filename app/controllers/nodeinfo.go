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
)

type NodeInfo struct {
  *revel.Controller
}

type IndexJson struct {
  Links []IndexLinkJson `json:"links"`
}

type IndexLinkJson struct {
  Rel string `json:"rel"`
  Href string `json:"href"`
}

type SchemaJson struct {
  Version string
  Software SchemaSoftwareJson
  Protocols interface{}
  Services SchemaServiceJson
  Usage SchemaUsageJson
  MetaData SchemaMetaDataJson
}

type SchemaProtocolsOne struct {
  Inbound []string
  Outbound []string
}

type SchemaSoftwareJson struct {
  Name string
  Version string
}

type SchemaServiceJson struct {
  Inbound []string
  Outbound []string
  OpenRegistrations bool
}

type SchemaUsageJson struct {
  Users SchemaUsersJson
  LocalPosts int
  LocalComments int
}

type SchemaUsersJson struct {
  Total int
  ActiveHalfyear int
  ActiveMonth int
}

type SchemaMetaDataJson struct {
  NodeName string
  XmppChat bool
  AdminAccount string
}

func (n NodeInfo) Index() revel.Result {
  revel.Config.SetSection("ganggo")
  address, found := revel.Config.String("address")
  if !found {
    return n.RenderJSON(struct{}{})
  }
  proto, found := revel.Config.String("proto")
  if !found {
    return n.RenderJSON(struct{}{})
  }

  return n.RenderJSON(IndexJson{
    Links: []IndexLinkJson{
      IndexLinkJson{
        Rel: "http://nodeinfo.diaspora.software/ns/schema/1.0",
        Href: proto + address + "/nodeinfo/1.0",
      },
      IndexLinkJson{
        Rel: "http://nodeinfo.diaspora.software/ns/schema/2.0",
        Href: proto + address + "/nodeinfo/2.0",
      },
    },
  })
}

func generateSchema(version string) SchemaJson {
  var protocols interface{}
  var softwareName string = "ganggo"
  var softwareVersion string = "6.6.6"

  revel.Config.SetSection("DEFAULT")
  name, found := revel.Config.String("app.name")
  if !found {
    revel.ERROR.Println("app.name configuration value not found!")
    return SchemaJson{}
  }
  appVersion, found := revel.Config.String("app.version")
  if !found {
    revel.ERROR.Println("app.version configuration value not found!")
    return SchemaJson{}
  }

  if version == "1.0" {
    softwareName = "diaspora"
    protocols = SchemaProtocolsOne{
      Outbound: []string{"diaspora"},
    }
  } else if version == "2.0" {
    softwareVersion = appVersion
    protocols = []string{"diaspora"}
  }

  return SchemaJson{
    Version: version,
    Software: SchemaSoftwareJson{
      Name: softwareName,
      Version: softwareVersion,
    },
    Protocols: protocols,
    Services: SchemaServiceJson{
      Inbound: []string{},
      Outbound: []string{},
      OpenRegistrations: true,
    },
    Usage: SchemaUsageJson{
      Users: SchemaUsersJson{
        Total: 0,
        ActiveHalfyear: 0,
        ActiveMonth: 0,
      },
      LocalPosts: 0,
      LocalComments: 0,
    },
    MetaData: SchemaMetaDataJson{
      NodeName: name,
      XmppChat: false,
      AdminAccount: "hq",
    },
  }
}

func (n NodeInfo) Schema(version string) revel.Result {
  return n.RenderJSON(generateSchema(version))
}
