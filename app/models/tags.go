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

import "gopkg.in/ganggo/gorm.v2"

type Tag struct {
  ID uint `gorm:"primary_key"`

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Name string `gorm:"size:191"`

  ShareableTaggings ShareableTaggings
  UserTaggings UserTaggings
}

type Tags []Tag

type ShareableTagging struct {
  ID uint `gorm:"primary_key"`

  TagID uint
  Public bool
  ShareableID uint
  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  ShareableType string `gorm:"size:191"`

  Post Post `json:",omitempty"`
  Comment Comment `json:",omitempty"`
}

type ShareableTaggings []ShareableTagging

type UserTagging struct {
  ID uint `gorm:"primary_key"`

  TagID uint
  UserID uint
}

type UserTaggings []UserTagging

func (shareable * ShareableTagging) AfterFind(db *gorm.DB) error {
  if shareable.ShareableType == ShareablePost {
    return db.First(&shareable.Post, shareable.ShareableID).Error
  } else if shareable.ShareableType == ShareableComment {
    return db.First(&shareable.Comment, shareable.ShareableID).Error
  }
  return nil
}

func (tag *Tag) FindByName(name string, user User, offset uint) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  err = db.Where(`name = ?`, name).First(tag).Error
  if err != nil {
    return err
  }

  query := db.Offset(offset).Limit(10).Table("shareable_taggings").
    Joins(`left join shareables on
      shareables.shareable_id = shareable_taggings.shareable_id and
      shareables.shareable_type = shareable_taggings.shareable_type`).
    Where(`shareable_taggings.public = ? and shareable_taggings.tag_id = ?`, true, tag.ID)

  if user.SerializedPrivateKey != "" {
    query = query.Or(`shareables.user_id = ? and shareable_taggings.tag_id = ?`, user.ID, tag.ID)
  }
  return query.Find(&tag.ShareableTaggings).Error
}
