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
  "git.feneas.org/ganggo/ganggo/app/models"
  "git.feneas.org/ganggo/federation"
  "time"
)

func (receiver *Receiver) Profile(entity federation.MessageProfile) {
  var person models.Person
  err := person.FindByAuthor(entity.Author())
  if err != nil {
    revel.AppLog.Error("Receiver Profile", err.Error(), err)
    return
  }

  birthday, timeErr := time.Parse("2006-02-01", entity.Birthday())
  profile := models.Profile{
    Author: entity.Author(),
    FirstName: entity.FirstName(),
    LastName: entity.LastName(),
    ImageUrl: entity.ImageUrl(),
    Gender: entity.Gender(),
    Bio: entity.Bio(),
    PersonID: person.ID,
    Searchable: entity.Public(),
    Location: entity.Location(),
    FullName: entity.FirstName() + " " + entity.LastName(),
    Nsfw: entity.Nsfw(),
  }
  if timeErr == nil {
    profile.Birthday = birthday
  }

  err = profile.CreateOrUpdate()
  if err != nil {
    revel.AppLog.Error("Receiver Profile", err.Error(), err)
    return
  }
}
