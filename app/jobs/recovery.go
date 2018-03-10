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
  "github.com/revel/revel"
  "github.com/ganggo/ganggo/app/models"
  federation "github.com/ganggo/federation"
  diaspora "github.com/ganggo/federation/diaspora"
  "strings"
  "fmt"
)

type Recovery struct {
  Shareable string
  Guid string
}

func (recovery Recovery) Run() {
  revel.AppLog.Debug("Running recovery",
    "shareable", recovery.Shareable, "guid", recovery.Guid)

  var pods models.Pods
  err := pods.FindRandom(10)
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  for _, pod := range pods {
    var message diaspora.Message
    err = federation.FetchXml(
      "GET", fmt.Sprintf("%s/fetch/%s/%s", pod.Host,
        strings.ToLower(recovery.Shareable), recovery.Guid,
      ), nil, &message)
    if err != nil {
      revel.AppLog.Error("Fetching from host failed",
        "host", pod.Host, "err", err)
      continue
    }
    _, err := message.Parse()
    if err != nil {
      revel.AppLog.Error("Parsing message from host failed",
        "host", pod.Host, "err", err)
      continue
    }
    // XXX
    //receiver := Receiver{Message: message, Entity: entity}
    //receiver.Run()
    break
  }
}

// XXX This could be the way of restoring all entities rather than one
// unfortunately diaspora only supports fetching post entities
// (see https://github.com/diaspora/diaspora_diaspora/issues/31#issue-142060252)
//
// Interactions will try to fetch all available
// child entities for a parent one
//func (recovery *Recovery) Interactions(host string) {
//  revel.AppLog.Debug("Running interaction recovery",
//    "shareable", recovery.Shareable, "guid", recovery.Guid)
//
//  switch recovery.Shareable {
//  case models.ShareablePost:
//    var inta Interactions
//    err := diaspora.FetchJson("GET", fmt.Sprintf(
//        "%s/posts/%s.json", host, recovery.Guid,
//      ), nil, &inta)
//    if err != nil {
//      revel.AppLog.Error("Fetching from host failed",
//        "host", host, "err", err)
//      return
//    }
//    for _, like := range inta.Interactions.Likes {
//      run.Now(Recovery{
//        Guid: like.Guid,
//        Shareable: models.ShareableLike,
//      })
//    }
//    for _, comment := range inta.Interactions.Comments {
//      run.Now(Recovery{
//        Guid: comment.Guid,
//        Shareable: models.ShareableComment,
//      })
//    }
//    for _, reshare := range inta.Interactions.Reshares {
//      run.Now(Recovery{
//        Guid: reshare.Guid,
//        Shareable: models.Reshare,
//      })
//    }
//  }
//}
