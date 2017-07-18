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
  federation "gopkg.in/ganggo/federation.v0"
)

func (d *Dispatcher) Post(post *federation.EntityStatusMessage) {
  entity := federation.Entity{
    Post: federation.EntityPost{
      StatusMessage: post,
    },
  }

  entityXml, err := xml.Marshal(entity)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  if d.AspectID > 0 {
    sendToAspect(
      d.AspectID, (*d).User.SerializedPrivateKey,
      post.Author, entityXml,
    )
  } else {
    payload, err := federation.MagicEnvelope(
      (*d).User.SerializedPrivateKey,
      post.Author, entityXml,
    ); if err != nil {
      revel.ERROR.Println(err)
      return
    }
    sendPublic(payload)
  }
}
