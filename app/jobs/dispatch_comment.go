package jobs
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
  "encoding/xml"
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
  federation "gopkg.in/ganggo/federation.v0"
)

func (dispatcher *Dispatcher) Comment(comment federation.EntityComment) {
  modelComment, ok := dispatcher.Model.(models.Comment)
  if !ok {
    revel.AppLog.Error(
      "Submitted model is not a type of models.Comment",
      "model", modelComment,
    )
    return
  }

  if !dispatcher.Relay {
    err := comment.AppendSignature([]byte(
        dispatcher.User.SerializedPrivateKey,
      ), comment.SignatureOrder(),
    )
    if err != nil {
      revel.AppLog.Error(err.Error())
      return
    }
  }

  entityXml, err := xml.Marshal(comment)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }

  parentPost, parentUser, _ := modelComment.ParentPostUser()
  dispatcher.Send(
    parentPost, parentUser, entityXml,
    modelComment.Signature.SignatureOrderID,
  )
}
