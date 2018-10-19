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

import (
  "strings"
  "git.feneas.org/ganggo/gorm"
  "sort"
  "errors"
  "time"
  "github.com/revel/revel"
  "crypto/rand"
  "crypto/rsa"
  "crypto/x509"
  "encoding/pem"
  "golang.org/x/crypto/bcrypt"
  "git.feneas.org/ganggo/ganggo/app/helpers"
  "git.feneas.org/ganggo/ganggo/app/jobs/notifier"
  run "github.com/revel/modules/jobs/app/jobs"
  "github.com/patrickmn/go-cache"
)

type User struct {
  gorm.Model

  // size should be max 191 with mysql innodb
  // cause asumming we use utf8mb 4*191 = 764 < 767
  Username string `gorm:"size:191"`
  Email string `gorm:"size:191"`
  SerializedPrivateKey string `gorm:"type:text" json:"-" xml:"-"`
  Password string `gorm:"-" json:"-" xml:"-"`
  EncryptedPassword string `json:"-" xml:"-"`
  LastSeen time.Time

  PersonID uint
  Person Person `gorm:"ForeignKey:PersonID"`

  Settings UserSettings `gorm:"AssociationForeignKey:UserID"`
  Aspects []Aspect `gorm:"AssociationForeignKey:UserID"`
}

type Users []User

type UserStream struct {
  ID uint `gorm:"primary_key"`
  CreatedAt time.Time
  UpdatedAt time.Time

  UserID uint
  Name string `gorm:"size:191"`

  Tags string
  People string
  Expression string

  User User `json:"-"`
}

type UserStreams []UserStream

func (user *User) BeforeCreate() error {
  // generate priv/pub key
  privKey, err := rsa.GenerateKey(rand.Reader, 2048)
  if err != nil {
    return err
  }

  // private key
  key := x509.MarshalPKCS1PrivateKey(privKey)
  block := pem.Block{Type: "PRIVATE KEY", Bytes: key}
  keyEncoded := pem.EncodeToMemory(&block)

  // public key
  pub, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
  if err != nil {
    return err
  }

  pubBlock := pem.Block{Type: "PUBLIC KEY", Bytes: pub}
  pubEncoded := pem.EncodeToMemory(&pubBlock)

  guid, err := helpers.Uuid(); if err != nil {
    return err
  }

  passwordEncoded, err := bcrypt.GenerateFromPassword(
    []byte(user.Password), -1); if err != nil {
    return err
  }

  revel.Config.SetSection("ganggo")
  host, found := revel.Config.String("address")
  if !found {
    return errors.New("No server address configured")
  }

  // set priv/pub keys and encrypted password
  user.SerializedPrivateKey = string(keyEncoded)
  user.EncryptedPassword = string(passwordEncoded)
  user.Person.Guid = guid
  user.Person.Author = user.Username + "@" + host
  user.Person.Profile.Author = user.Username + "@" + host
  user.Person.SerializedPublicKey = string(pubEncoded)

  return nil
}

func (user *User) AfterCreate(tx *gorm.DB) error {
  token, err := helpers.Token(); if err != nil {
    return err
  }

  // set initial telegram token
  setting := UserSetting{
    UserID: user.ID,
    Key: UserSettingTelegramVerified,
    Value: token,
  }
  if err = setting.Update(); err != nil {
    return err
  }

  return tx.Model(&user.Person).Update("user_id", user.ID).Error
}

func (user *User) AfterFind(db *gorm.DB) error {
  if structLoaded(user.Person.CreatedAt) {
    return nil
  }

  err := db.Model(user).Related(&user.Person).Error
  if err != nil {
    return err
  }

  // ignore errors its possible that it
  // doesn't exist on newly created users
  db.Model(user).Related(&user.Settings)

  return db.Model(user).Related(&user.Aspects).Error
}

func (user *User) FindByID(id uint) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Find(user, id).Error
}

func (user *User) FindByUsername(name string) (err error) {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("username = ?", name).Find(user).Error
}

func (users *Users) Count() (count int) {
  db, err := OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  defer db.Close()

  if countInt, found := databaseCache.Get("users.count.all"); found {
    return countInt.(int)
  }

  db.Table("users").Count(&count)
  databaseCache.Set("users.count.all", count, cache.DefaultExpiration)
  return
}

func (users *Users) ActiveHalfyear() (count int) {
  db, err := OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  defer db.Close()

  if countInt, found := databaseCache.Get("users.count.halfyear"); found {
    return countInt.(int)
  }

  halfYear := time.Now().AddDate(0, -6, 0)
  db.Table("users").Where("last_seen >= ?", halfYear).Count(&count)
  databaseCache.Set("users.count.halfyear", count, cache.DefaultExpiration)
  return
}

func (users *Users) ActiveMonth() (count int) {
  db, err := OpenDatabase()
  if err != nil {
    revel.AppLog.Error(err.Error())
    return
  }
  defer db.Close()

  if countInt, found := databaseCache.Get("users.count.month"); found {
    return countInt.(int)
  }

  month := time.Now().AddDate(0, -1, 0)
  db.Table("users").Where("last_seen >= ?", month).Count(&count)
  databaseCache.Set("users.count.month", count, cache.DefaultExpiration)
  return
}

func (user *User) Notify(model Model) error {
  // do not send notification for your own activity
  if user.Person.ID == model.FetchPersonID() {
    return nil
  }

  notify := Notification{
    ShareableType: model.FetchType(),
    ShareableGuid: model.FetchGuid(),
    UserID: user.ID,
    PersonID: model.FetchPersonID(),
    Unread: true,
  }
  err := notify.Create()
  if err != nil {
    return notify.Update()
  }

  // send notification via mail/telegram/webhook
  run.Now(notifier.Notifier{Message: []interface{}{
    notifier.Telegram{
      ID: user.Settings.GetValue(UserSettingTelegramID),
      Text: model.FetchText(),
    },
    notifier.Mail{
      To: user.Settings.GetValue(UserSettingMailAddress),
      Subject: "Reply:", // XXX
      Body: model.FetchText(),
      Lang: user.Settings.GetValue(UserSettingLanguage),
    },
  }})
  return err
}

func (stream *UserStream) Create() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Create(stream).Error
}

func (stream *UserStream) FindByName(name string) error { BACKEND_ONLY()
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("name = ?", name).Find(stream).Error
}

func (stream *UserStream) FetchPosts(posts *Posts, offset uint) error { BACKEND_ONLY()
  tagNames := strings.Split(stream.Tags, ",")
  people := strings.Split(stream.People, ",")

  for _, name := range tagNames {
    var tag Tag
    err := tag.FindByName(name, stream.User, offset)
    if err != nil {
      if err == gorm.ErrRecordNotFound {
        continue
      }
      return err
    }
    for _, tagging := range tag.ShareableTaggings {
      *posts = append(*posts, tagging.Post)
    }
  }

  for _, author := range people {
    var person Person
    err := person.FindByAuthor(author)
    if err != nil {
      if err == gorm.ErrRecordNotFound {
        continue
      }
      return err
    }
    var authorPosts Posts
    err = authorPosts.FindAllByUserAndPersonID(
      stream.User, person.ID, offset)
    if err != nil && err != gorm.ErrRecordNotFound {
      return err
    }
    *posts = append(*posts, authorPosts...)
  }

  var expressionPosts Posts
  err := expressionPosts.FindAllByUserAndText(
    stream.User, stream.Expression, offset)
  if err != nil && err != gorm.ErrRecordNotFound {
    return err
  }
  *posts = append(*posts, expressionPosts...)

  sort.Sort(*posts)

  if uint(len(*posts)) > offset && offset > 0 {
    *posts = (*posts)[:offset-1]
  }
  return nil
}

func (streams *UserStreams) FindByUser(user User) error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Where("user_id = ?", user.ID).Find(streams).Error
}

func (stream *UserStream) Delete() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  if stream.ID == 0 || stream.UserID == 0 {
    // NOTE ID being zero will delete ALL entries
    return errors.New("Cannot delete user stream without ID and UserID")
  }
  return db.Delete(stream).Error
}

func (user *User) ActiveLastDay() bool {
  oneDayAgo := time.Now().AddDate(0, 0, -1)
  return user.LastSeen.After(oneDayAgo)
}

func (user *User) UpdateLastSeen() error {
  db, err := OpenDatabase()
  if err != nil {
    return err
  }
  defer db.Close()

  return db.Model(user).Update("last_seen", time.Now()).Error
}
