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
  "git.feneas.org/ganggo/ganggo/app"
  "git.feneas.org/ganggo/ganggo/app/models"
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
  Version string `json:"version"`
  Software SchemaSoftwareJson `json:"software"`
  Protocols interface{} `json:"protocols"`
  Services SchemaInOutJson `json:"services"`
  OpenRegistrations bool `json:"openRegistrations"`
  Usage SchemaUsageJson `json:"usage"`
  MetaData SchemaMetaDataJson `json:"metadata"`
}

type SchemaInOutJson struct {
  Inbound []string `json:"inbound"`
  Outbound []string `json:"outbound"`
}

type SchemaSoftwareJson struct {
  Name string `json:"name"`
  Version string `json:"version"`
}

type SchemaUsageJson struct {
  Users SchemaUsersJson `json:"users"`
  LocalPosts int `json:"localPosts"`
  LocalComments int `json:"localComments"`
}

type SchemaUsersJson struct {
  Total int `json:"total"`
  ActiveHalfyear int `json:"activeHalfyear"`
  ActiveMonth int `json:"activeMonth"`
}

type SchemaMetaDataJson struct {
  NodeName string `json:"nodeName"`
  XmppChat bool `json:"xmppChat"`
  AdminAccount string `json:"adminAccount"`
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
  revel.Config.SetSection("DEFAULT")
  appName, found := revel.Config.String("app.name")
  if !found {
    revel.AppLog.Error("app.name configuration value not found!")
    return SchemaJson{}
  }

  if version == "1.0" {
    protocols = SchemaInOutJson{
      Inbound: []string{"diaspora"},
      Outbound: []string{"diaspora"},
    }
  } else if version == "2.0" {
    protocols = []string{"diaspora"}
  }

  var (
    users models.Users
    comments models.Comments
    posts models.Posts
  )

  return SchemaJson{
    Version: version,
    Software: SchemaSoftwareJson{
      Name: "ganggo",
      Version: app.AppVersion,
    },
    Protocols: protocols,
    Services: SchemaInOutJson{
      Inbound: []string{},
      Outbound: []string{},
    },
    OpenRegistrations: true,
    Usage: SchemaUsageJson{
      Users: SchemaUsersJson{
        Total: users.Count(),
        ActiveHalfyear: users.ActiveHalfyear(),
        ActiveMonth: users.ActiveMonth(),
      },
      LocalPosts: posts.CountLocal(),
      LocalComments: comments.CountLocal(),
    },
    MetaData: SchemaMetaDataJson{
      NodeName: appName,
      XmppChat: false,
      AdminAccount: "hq",
    },
  }
}

func (n NodeInfo) Schema(version string) revel.Result {
  return n.RenderJSON(generateSchema(version))
}
