package controllers
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
  "git.feneas.org/ganggo/ganggo/app/helpers"
  "git.feneas.org/ganggo/ganggo/app/models"
  "git.feneas.org/ganggo/ganggo/app/jobs/notifier"
  run "github.com/revel/modules/jobs/app/jobs"
  "regexp"
)

type Setting struct {
  *revel.Controller
}

func (s Setting) Index() revel.Result {
  user, err := models.CurrentUser(s.Controller)
  if err != nil {
    s.Log.Error("Cannot fetch current user", "error", err)
    return s.RenderError(err)
  }
  s.ViewArgs["currentUser"] = user

  var tokens models.OAuthTokens
  err = tokens.FindByUserID(user.ID)
  if err != nil {
    s.Log.Error("Cannot fetch user tokens", "error", err)
    return s.RenderError(err)
  }
  s.ViewArgs["tokens"] = tokens

  return s.RenderTemplate("user/settings.html")
}

func (s Setting) Update() revel.Result {
  var settings models.UserSettings

  var email, lang string
  s.Params.Bind(&email, "email")
  s.Params.Bind(&lang, "lang")

  // verification tokens
  var emailToken string
  s.Params.Bind(&emailToken, "emailtoken")

  user, err := models.CurrentUser(s.Controller)
  if err != nil {
    s.Log.Error("Cannot fetch current user", "error", err)
    return s.RenderError(err)
  }

  emRegex := regexp.MustCompile(`^[\w\d\._%\+-]+@[\w\d\.-]+\.\w{2,}$`)
  emailChanged := user.Settings.GetValue(models.UserSettingMailVerified) == "" ||
    user.Settings.GetValue(models.UserSettingMailAddress) != email
  if emailChanged && (emRegex.MatchString(email) || email == "") {
    token, err := helpers.Token()
    if err != nil {
      s.Log.Error("Cannot generate token", "error", err)
      return s.RenderError(err)
    }

    settings = append(settings, models.UserSettings{
      models.UserSetting{
        UserID: user.ID,
        Key: models.UserSettingMailAddress,
        Value: email,
      },
      models.UserSetting{
        UserID: user.ID,
        Key: models.UserSettingMailVerified,
        Value: token,
      },
    }...)

    if email != "" {
      run.Now(notifier.Notifier{Message: []interface{}{
        notifier.Mail{
          To: email,
          Subject: "Reply:", // XXX
          Body: token,
          Lang: user.Settings.GetValue(models.UserSettingLanguage),
        },
      }})
    }
  }

  // mail verification token
  if emailToken != "" &&
    user.Settings.GetValue(models.UserSettingMailVerified) == emailToken {
    settings = append(settings, models.UserSetting{
      UserID: user.ID,
      Key: models.UserSettingMailVerified,
      Value: "true",
    })
  }

  lgRegex := regexp.MustCompile(`^[\w-_]{1,}$`)
  if lgRegex.MatchString(lang) || lang == "" {
    settings = append(settings, models.UserSetting{
      UserID: user.ID,
      Key: models.UserSettingLanguage,
      Value: lang,
    })
  }

  var errors bool = false
  errs := settings.Update()
  for _, err := range errs {
    if err != nil {
      errors = true
      s.Log.Error("Cannot update settings", "errors", err)
    }
  }

  if errors || len(settings) == 0 {
    s.Flash.Error(s.Message("flash.errors.settings"))
  } else {
    s.Flash.Success(s.Message("flash.success.settings"))
  }
  return s.Redirect(Setting.Index)
}
