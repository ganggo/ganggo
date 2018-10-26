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

type RetryOnFail struct {
  Pod *models.Pod
  Send func() error
  After []time.Duration

  firstRun bool
}

func (retry RetryOnFail) Run() {
  // set default values on first run
  if !retry.firstRun {
    retry.firstRun = true
    if len(retry.After) == 0 {
      retry.After = append(retry.After, []time.Duration{
        time.Minute, time.Hour, 24 * time.Hour,
      }...)
    }
  }

  // execute job and check on errors
  err := retry.Send()
  if err != nil {
    if len(retry.After) > 0 {
      // repeat until timeout (empty array)
      duration := retry.After[0]
      retry.After = retry.After[1:]
      // repeat until timeout (empty array)
      revel.AppLog.Warn("Jobs Retry", "waitfor", duration, "error", err)
      run.In(duration, retry)
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
    }
  }
}
