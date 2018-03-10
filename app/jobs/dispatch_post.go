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
  "github.com/ganggo/ganggo/app/models"
  diaspora "github.com/ganggo/federation/diaspora"
)

func (dispatcher *Dispatcher) Reshare(reshare diaspora.EntityReshare) {
  entityXml, err := xml.Marshal(reshare)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  post(dispatcher, entityXml)
}

func (dispatcher *Dispatcher) StatusMessage(message diaspora.EntityStatusMessage) {
  entityXml, err := xml.Marshal(message)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  post(dispatcher, entityXml)
}

func post(dispatcher *Dispatcher, entityXml []byte) {
  modelPost, ok := dispatcher.Model.(models.Post)
  if !ok {
    revel.AppLog.Error(
      "Submitted model is not a type of models.Post",
      "model", modelPost,
    )
    return
  }

  if modelPost.Exists(modelPost.ID) {
    if user, ok := modelPost.IsLocal(); ok {
      dispatcher.Send(modelPost, user, entityXml, 0)
    } else {
      // check if it is a reshare
      if modelPost.RootPersonID > 0 {
        dispatcher.Send(modelPost, models.User{}, entityXml, 0)
      } else {
        revel.AppLog.Error("Cannot relay remote posts", "post", modelPost)
        return;
      }
    }
  }
}
