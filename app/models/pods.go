package models
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

import "time"

type Pod struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Host string `gorm:"size:191" json:"host"`
  Helo bool `gorm:"default:false"`
}

type Pods []Pod

func (pod *Pod) Save() error { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Save(pod).Error
}

func (pod *Pod) CreateOrFindHost() (err error) { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return
  }
  defer db.Close()

  if db.Where("host = ?", pod.Host).Find(pod).RecordNotFound() {
    return db.Create(pod).Error
  }
  return
}

func (pods *Pods) FindRandom(limit uint) (err error) { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Find(pods).Order(randomOrder()).Limit(limit).Error
}

func (pods *Pods) FindAll() (err error) { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Find(pods).Error
}

func (pods *Pods) FindByHelo(helo bool) error { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("helo = ?", helo).Find(pods).Error
}
