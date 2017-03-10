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
  "gopkg.in/ganggo/ganggo.v0/app/models"
  federation "gopkg.in/ganggo/federation.v0"
  "bytes"
  "sync"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
  MAX_ASYNC_JOBS int
)

type Dispatcher struct {
  User models.User
  Message interface{}
}

func (d *Dispatcher) Run() {
  revel.Config.SetSection("ganggo")
  MAX_ASYNC_JOBS = revel.Config.IntDefault("max_async_jobs", 20)

  switch entity := d.Message.(type) {
  case federation.EntityStatusMessage:
    d.Post(&entity)
  case federation.EntityComment:
    d.Comment(&entity)
  default:
    revel.ERROR.Println("Unknown entity type in dispatcher!")
  }
}

func send(payload []byte) {
  revel.TRACE.Println("Sending payload", string(payload))

  //pods, err := models.DB.FindPods()
  //if err != nil {
  //  revel.ERROR.Println(err)
  //  return
  //}
  pods := []models.Pod{
    {Host: "192.168.0.173:3000"},
  }

  var wg sync.WaitGroup
  // lets do it \m/ and publish the post
  for i, pod := range pods {
    wg.Add(1)
    // do everything asynchronous
    go func() {
      err := federation.PushXmlToPublic(
        pod.Host, bytes.NewBuffer(payload), true,
      )
      if err != nil {
        revel.ERROR.Println(err, pod.Host)
      }
    }()
    // do a maximum of e.g. 20 jobs async
    if i >= MAX_ASYNC_JOBS {
      wg.Wait()
    }
  }
}
