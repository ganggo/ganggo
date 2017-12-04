package views
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
  "regexp"
  "path/filepath"
  "os"
  "github.com/shaoshing/train"
  "github.com/revel/revel"
  "gopkg.in/ganggo/ganggo.v0/app/models"
  "gopkg.in/ganggo/ganggo.v0/app/helpers"
  "html/template"
  "github.com/jinzhu/gorm"
  "github.com/dchest/captcha"
  _ "github.com/jinzhu/gorm/dialects/postgres"
  _ "github.com/jinzhu/gorm/dialects/mssql"
  _ "github.com/jinzhu/gorm/dialects/mysql"
  _ "github.com/jinzhu/gorm/dialects/sqlite"
)

var TemplateFuncs = map[string]interface{}{
  // database types
  "IsReshare": func(a string) bool {
    return (a == models.Reshare)
  },
  "IsStatusMessage": func(a string) bool {
    return (a == models.StatusMessage)
  },
  "IsShareablePost": func(a string) bool {
    return (a == models.ShareablePost)
  },
  "IsShareableComment": func(a string) bool {
    return (a == models.ShareableComment)
  },
  "LikesByTargetID": func(id uint) []models.Like {
    return likes(id, true)
  },
  "DislikesByTargetID": func(id uint) []models.Like {
    return likes(id, false)
  },
  "PostByGuid": func(guid string) (post models.Post) {
    db, err := gorm.Open(models.DB.Driver, models.DB.Url)
    if err != nil {
      revel.ERROR.Println(err)
      return
    }
    defer db.Close()

    err = db.Where("guid = ?", guid).First(&post).Error
    if err != nil {
      revel.ERROR.Println(err, guid)
      return
    }
    return
  },
  "PersonByID": func(id uint) (person models.Person) {
    db, err := gorm.Open(models.DB.Driver, models.DB.Url)
    if err != nil {
      revel.ERROR.Println(err)
      return
    }
    defer db.Close()

    err = db.First(&person, id).Error
    if err != nil {
      revel.ERROR.Println(err, id)
      return
    }

    err = db.Where("person_id = ?", person.ID).First(&person.Profile).Error
    if err != nil {
      revel.ERROR.Println(err, person)
      return
    }
    return
  },
  // string parse helper
  "HostFromHandle": func(handle string) (host string) {
    _, host, err := helpers.ParseAuthor(handle)
    if err != nil {
      revel.ERROR.Println(err)
      return
    }
    return
  },
  // captcha generator
  "CaptchaNew": func() string { return captcha.New() },
  "FindAvailableLocales": func () (list []string) {
    directory := filepath.Join(revel.BasePath, "messages")
    re := regexp.MustCompile(`ganggo\.([\w-_]{1,})$`)
    err := filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
      result := re.FindAllStringSubmatch(path, 1)
      if len(result) > 0 && len(result[0]) > 0 {
        list = append(list, result[0][1])
      }
      return nil
    })
    if err != nil {
      revel.ERROR.Println(err)
      return
    }
    return
  },
  "FindUnreadNotifications": func(user models.User) (notify models.Notifications) {
    err := notify.FindUnreadByUserID(user.ID)
    if err != nil {
      revel.ERROR.Println(err)
      return
    }
    return
  },
  // custom train script/stylesheet include functions
  "javascript_include_tag": func(name string) template.HTML {
    path := "/assets/javascripts/" + name + ".js"
    src := loadAndFetchManifestEntry(path)
    tmpl := `<script src="/public` + src + `"></script>`
    if src == "" {
      tmpl = "Cannot load asset '" + name + "'.js"
      revel.ERROR.Println(tmpl)
    }
    return template.HTML(tmpl)
  },
  "stylesheet_link_tag": func(name string) template.HTML {
    path := "/assets/stylesheets/" + name + ".css"
    src := loadAndFetchManifestEntry(path)
    tmpl := `<link type="text/css" rel="stylesheet" href="/public` + src + `">`
    if src == "" {
      tmpl = "Cannot load asset '" + name + "'.css"
      revel.ERROR.Println(tmpl)
    }
    return template.HTML(tmpl)
  },
  "eq": func(a, b interface {}) bool {
    return a == b
  },
  "ne": func(a, b interface {}) bool {
    return a != b
  },
  "add": func(a, b int) int {
    return a + b
  },
  "sub": func(a, b int) int {
    return a - b
  },
  "concat": func(a, b string) string {
    return a + b
  },
}

func loadAndFetchManifestEntry(path string) (src string) {
  if len(train.ManifestInfo) <= 0 {
    train.Config.PublicPath = "src/" + revel.ImportPath + "/public"
    err := train.LoadManifestInfo()
    if err != nil {
      revel.ERROR.Println(err)
      return
    }
  }
  src = train.ManifestInfo[path]
  return
}

func likes(id uint, like bool) (likes []models.Like) {
  db, err := gorm.Open(models.DB.Driver, models.DB.Url)
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  defer db.Close()

  err = db.Where(
    `target_type = ?
      and target_id = ?
      and positive = ?`,
    models.ShareablePost, id, like,
  ).Find(&likes).Error
  if err != nil {
    revel.ERROR.Println(err)
    return
  }
  return
}
