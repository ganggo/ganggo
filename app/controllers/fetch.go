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
  "encoding/xml"
  "github.com/revel/revel"
  "git.feneas.org/ganggo/ganggo/app/models"
  "git.feneas.org/ganggo/ganggo/app/helpers"
  federation "git.feneas.org/ganggo/federation"
  "strings"
)

type Fetch struct {
  *revel.Controller
}

type RenderBytes []byte

func (b RenderBytes) Apply(req *revel.Request, resp *revel.Response) {
  resp.WriteHeader(http.StatusOK, "application/magic-envelope+xml")
  resp.GetWriter().Write(b)
}

func (f Fetch) Index(shareable, guid string) revel.Result {
  var payload []byte
  switch {
  case strings.EqualFold(shareable, models.ShareablePost):
    fallthrough
  // XXX hard-coded status message type
  case strings.EqualFold(shareable, "status_message"):
    var post models.Post
    err := post.FindByGuid(guid)
    if err != nil {
      f.Log.Debug("Fetch post error", "err", err)
      return f.NotFound(err.Error())
    }

    if !post.Public {
      // only public entities should be fetchable
      return f.NotFound("record not found")
    }

    if user, ok := post.IsLocal(); ok {
      privKey, err := federation.ParseRSAPrivateKey(
        []byte(user.SerializedPrivateKey))
      if err != nil {
        f.Log.Error("Fetch parse key error", "err", err)
        return f.NotFound("record not found")
      }

      entity := federation.EntityStatusMessage{
        Text: post.Text,
        Author: post.Person.Author,
        Guid: post.Guid,
        ProviderName: post.ProviderName,
        Public: post.Public,
      }
      entity.CreatedAt.New(post.CreatedAt)

      entityXml, err := xml.Marshal(entity)
      if err != nil {
        f.Log.Error("Fetch marshal error", "err", err)
        return f.RenderError(err)
      }

      payload, err = federation.MagicEnvelope(
        privKey, user.Person.Author, entityXml,
      )
      if err != nil {
        f.Log.Error("Fetch magic envelope error", "err", err)
        return f.RenderError(err)
      }
    } else {
      _, host, err := helpers.ParseAuthor(post.Person.Author)
      if err != nil {
        f.Log.Error("Fetch parse author error", "err", err)
        return f.RenderError(err)
      }
      return f.Redirect("http://%s/fetch/%s/%s", host, shareable, guid)
    }
  default:
    return f.NotFound("not supported type")
  }
  return RenderBytes(payload)
}
