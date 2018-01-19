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

type SignatureOrder struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Order string `gorm:"size:191"`
}

type SignatureOrders []SignatureOrder

func (o *SignatureOrder) FindByID(id uint) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Find(o, id).Error
}

func (o *SignatureOrder) CreateOrFind() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.FirstOrCreate(
    o, SignatureOrder{Order: o.Order},
  ).Error
}
