package main
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
  "io/ioutil"
  "time"
  "fmt"
  "flag"
  "github.com/jpillora/overseer"
  "os"
  "os/exec"
  "github.com/revel/config"
  "runtime"
  "gopkg.in/AlecAivazis/survey.v1"
)

var (
  mode string = "prod"
  updateChannel string = "alpha"
  version string  // version will be defined on compile time
  packageName = "git.feneas.org/ganggo/ganggo"
  srcDir = "src/" + packageName
  configPath = srcDir + "/conf/app.conf"
  userConfigPath = "app.conf"
  interval = 10 * time.Minute
)

var answers = struct {
  Protocol string
  Tld string
  DatabaseDriver string
  DatabaseUsername string
  DatabasePassword string
  DatabaseHost string
  DatabaseName string
  DatabaseDSN string
  Relay bool
  RelayTags string
}{}

var questions =[]*survey.Question{
  {
    Name: "protocol",
    Prompt: &survey.Select{
      Message: "Which protocol will be used for the webpage?",
      Options: []string{"http://", "https://"},
      Default: "http://",
    },
    Validate: survey.Required,
  },
  {
    Name: "tld",
    Prompt: &survey.Input{
      Message: "Which top level domain will be used?",
      Help: "This could be example.com, pod.example.com or example.com:8080",
      Default: "localhost:9000",
    },
    Validate: survey.Required,
  },
  {
    Name: "databaseDriver",
    Prompt: &survey.Select{
      Message: "What database do you want to use?",
      Options: []string{"mssql", "mysql", "postgres", "sqlite"},
    },
    Validate: func (val interface{}) error {
      if str, ok := val.(string) ; ok {
        switch str {
        case "mysql":
          answers.DatabaseDSN = "%s:%s@tcp(%s)/%s?parseTime=true"
        case "postgres":
          answers.DatabaseDSN = "user=%s password=%s host=%s dbname=%s sslmode=disable"
        case "sqlite":
          prompt := &survey.Input{
            Message: "Please specify the database file path:",
            Default: "./ganggo.db",
          }
          survey.AskOne(prompt, &answers.DatabaseDSN, nil)
        default:
          prompt := &survey.Input{
            Message: "Please specify your database DSN (Data Source Name):",
            Help: "e.g. [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]",
          }
          survey.AskOne(prompt, &answers.DatabaseDSN, nil)
        }
      }
      return nil
    },
  },
  {
    Name: "databaseUsername",
    Prompt: &survey.Input{Message: "What is the database username?"},
    Validate: survey.Required,
  },
  {
    Name: "databasePassword",
    Prompt: &survey.Password{Message: "What is the database password?"},
    Validate: survey.Required,
  },
  {
    Name: "databaseHost",
    Prompt: &survey.Input{
      Message: "What is the database hostname?",
      Default: "localhost",
    },
    Validate: survey.Required,
  },
  {
    Name: "databaseName",
    Prompt: &survey.Input{
      Message: "What is the database name?",
      Default: "ganggo",
    },
    Validate: survey.Required,
  },
  {
    Name: "relay",
    Prompt: &survey.Confirm{
      Help: "The relay is disabled on default in respect of privacy matters!\n" +
      "ⓘ You should think about enabling it! The software will register at\n" +
      "ⓘ the-federation.info and start using the social-relay. After using it you\n" +
      "ⓘ will benefit from public posts and comments from the relay and fill your\n" +
      "ⓘ pod up more quicklier then just using the diaspora protocol!",
      Message: "Enable the relay for fetching public entities?",
      Default: false,
    },
    Validate: func (val interface{}) error {
      var relayType string
      if enabled, ok := val.(bool); ok && enabled {
        prompt := &survey.Select{
          Message: "Do you want to send/receive all public entities or only tags?",
          Options: []string{"tags", "all"},
          Default: "all",
        }
        survey.AskOne(prompt, &relayType, nil)
        if relayType == "tags" {
          prompt := &survey.Input{
            Message: "Which tags do you want to be relayed?",
            Help: "e.g. social,network,politics",
          }
          survey.AskOne(prompt, &answers.RelayTags, nil)
        }
      }
      return nil
    },
  },
}

func init() {
  flag.StringVar(&mode, "mode", mode, "Start revel server in production or development mode")
  flag.StringVar(&updateChannel, "channel", updateChannel, "Specify an update channel e.g. stable, beta or alpha.\n" +
  "\tYou can disable auto-updates with '-channel disable'")
  flag.StringVar(&userConfigPath, "config", userConfigPath, "Set a different config path")
}

func main() {
  // Parse command-line flags
  flag.Parse()

  if updateChannel != "alpha" {
    interval = 1 * time.Hour
  }

  // Delete old files
  os.RemoveAll("src")
  // Extract assets and golang libraries
  err := RestoreAssets("", "src")
  if err != nil {
    panic(err)
  }

  // if app.conf is not in the current
  // directory start the wizard and create it
  if _, err := os.Stat(userConfigPath); os.IsNotExist(err) {
    if err := survey.Ask(questions, &answers); err != nil {
      panic(err)
    }
    // generate new configuration
    var section = "ganggo"
    config, err := config.ReadDefault(configPath + ".example")
    if err != nil { panic(err) }
    config.AddOption(section, "proto", answers.Protocol)
    config.AddOption(section, "address", answers.Tld)
    config.AddOption(section, "db.driver", answers.DatabaseDriver)
    config.AddOption(section, "db.dsn", answers.DatabaseDSN)
    config.AddOption(section, "db.user", answers.DatabaseUsername)
    config.AddOption(section, "db.password", answers.DatabasePassword)
    config.AddOption(section, "db.host", answers.DatabaseHost)
    config.AddOption(section, "db.database", answers.DatabaseName)
    if answers.Relay {
      config.AddOption(section, "relay.subscribe", "true")
    } else {
      config.AddOption(section, "relay.subscribe", "false")
    }
    if answers.RelayTags == "" {
      config.AddOption(section, "relay.scope", "all")
    } else {
      config.AddOption(section, "relay.scope", "tags")
      config.AddOption(section, "relay.tags", answers.RelayTags)
    }
    err = config.WriteFile(userConfigPath, 0644, "AUTO-GENERATED BY GANGGO WIZARD")
    if err != nil { panic(err) }
  }

  overseer.Run(overseer.Config{
    Program: prog,
    Fetcher: &Bintray{
      User: "ganggo",
      Repo: "ganggo",
      Pkg: updateChannel,
      Interval: interval,
    },
    Debug: true,
  })
}

func prog(state overseer.State) {
  buildID := int32(time.Now().Unix())
  binFile := fmt.Sprintf("./ganggo.%s.%d", version, buildID)
  if runtime.GOOS == "windows" {
    binFile = fmt.Sprintf("ganggo.%s.%d.exe", version, buildID)
  }
  // clean-up binaries
  defer func() {
    os.Remove(binFile)
  }()

  // restore app binary
  asset, err := ganggoBytes()
  if err != nil {
    panic(err)
  }
  err = ioutil.WriteFile(binFile, asset, 0755)
  if err != nil {
    panic(err)
  }

  // read config file from wizard and copy it to src folder
  config, err := config.ReadDefault(userConfigPath)
  if err != nil {
    panic(err)
  }
  err = config.WriteFile(configPath, 0644, "AUTO-GENERATED BY GANGGO WIZARD")
  if err != nil {
    panic(err)
  }

  cmd := exec.Command(binFile, "-importPath", packageName,
    "-srcPath", "src", "-runMode", mode)

  if mode != "prod" {
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
  }

  fmt.Printf("app#%s %s listening...\n", version, state.ID)
  if err := cmd.Start(); err != nil {
    panic(err)
  }

  // wait for overseer to restart or terminate
  <-state.GracefulShutdown

  // send kill signal to old process
  cmd.Process.Kill()

  fmt.Printf("app#%s %s exiting...\n", version, state.ID)
}
