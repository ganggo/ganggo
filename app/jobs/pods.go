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
  "github.com/ganggo/federation"
  "github.com/jinzhu/gorm"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

type PodsJob struct {}

type PodsJson struct {
  Pods []models.Pod `json:"pods"`
}

func (p PodsJob) Run() {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    return
  }
  defer db.Close()

  revel.Config.SetSection("ganggo")
  endpoint, found := revel.Config.String("pods.discovery")
  if !found {
    revel.ERROR.Println("pods.discovery configuration missing")
    return
  }

  var fetched PodsJson
  err = federation.FetchJson("GET", endpoint, nil, &fetched)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  var dbPods []models.Pod
  err = db.Find(&dbPods).Error
  if err != nil {
    revel.ERROR.Println(err)
    return
  }

  // delete old hosts from database
  var offlineCount int
  for _, dbPod := range dbPods {
    if !contains(fetched.Pods, dbPod.Host) {
      db.Delete(&dbPod)
      offlineCount += 1
    }
  }
  revel.INFO.Println("PodsJob deleted", offlineCount, "old hosts")

  // insert new hosts only
  var newCount int
  for _, fetchedPod := range fetched.Pods {
    if !contains(dbPods, fetchedPod.Host) {
      db.Create(&fetchedPod)
      newCount += 1
    }
  }
  revel.INFO.Println("PodsJob discovered", newCount, "new hosts")
}

func contains(a []models.Pod, host string) bool {
  for _, pod := range a {
    if pod.Host == host {
      return true
    }
  }
  return false
}
