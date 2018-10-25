package jobs
//
// GangGo Application Server
// Copyright (C) 2018 Lukas Matt <lukas@zauberstuhl.de>
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
  "git.feneas.org/ganggo/ganggo/app/models"
  run "github.com/revel/modules/jobs/app/jobs"
  "time"
)

type Retry struct {
  Pod *models.Pod
  Send func() error
  wait time.Duration
}

func (retry Retry) Run() {
  err := retry.Send()
  if err != nil {
    if retry.wait == 0 {
      retry.wait = time.Minute
    } else if retry.wait == time.Minute {
      retry.wait = time.Hour
    } else if retry.wait == time.Hour {
      retry.wait = 24 * time.Hour
    } else {
      // this server is probably down. skip it..
      revel.AppLog.Error("Jobs Retry", "error", err)
      if retry.Pod != nil {
        retry.Pod.Alive = false
        err = retry.Pod.Save()
        if err != nil {
          revel.AppLog.Error("Jobs Retry", "error", err)
        }
      }
      return
    }
    revel.AppLog.Warn("Jobs Retry", "waitfor", retry.wait, "error", err)

    // repeat until timeout
    run.In(retry.wait, retry)
  }
}
