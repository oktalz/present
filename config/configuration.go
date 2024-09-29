package configuration

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/oklog/ulid/v2"
	"github.com/oktalz/present/archive"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

type Config struct {
	Version  bool   `ff:"short: v, long: version,  usage: 'show version'"`
	Tag      bool   `ff:"short: t, long: tag,      usage: 'show tag'"`
	Compress string `ff:"short: c, long: compress, usage: 'compress current folder'"`
	File     string `ff:"short: f, long: file,     usage: 'file to open (.tar.gz format)'"`
	GIT      string `ff:"short: g, long: git,      usage: 'git repository URL'"`
	Dir      string `ff:"short: d, long: dir,      usage: 'directory to open'"`
	Help     bool   `ff:"          long: help,     usage: 'help'"`

	Address string `ff:"short: h, long:host, default: 127.0.0.1, usage: 'address that present will listen on'"`
	Port    int    `ff:"short: p, long:port, default: 8080, usage: 'port that present will listen on'"`

	Security Security
	Controls Controls
	Update   Update
}

type Security struct {
	AdminPwd        string `ff:"long: admin-pwd,         usage: 'admin password'"`
	UserPwd         string `ff:"long: user-pwd,          usage: 'user password'"`
	AdminPwdDisable bool   `ff:"long: admin-pwd-disable, usage: 'disable admin password'"`
}

type Controls struct {
	Menu          string `ff:"long: menu,              usage: 'keys that opens menu'"`
	NextPage      string `ff:"long: next-page,         usage: 'keys that go to next page'"`
	PreviousPage  string `ff:"long: previous-page,     usage: 'keys that go to previous page'"`
	TerminalCast  string `ff:"long: terminal-cast,     usage: 'keys that run commands'"`
	TerminalClose string `ff:"long: terminal-close,    usage: 'keys that closes terminal'"`
}

func Get() Config {
	configuration := Config{}
	security := Security{}
	controls := Controls{}
	update := Update{}
	osArgsFF := ff.NewFlagSet("present")
	err := osArgsFF.AddStruct(&configuration)
	if err != nil {
		panic(err)
	}

	err = osArgsFF.AddStruct(&security)
	if err != nil {
		panic(err)
	}

	err = osArgsFF.AddStruct(&controls)
	if err != nil {
		panic(err)
	}

	err = osArgsFF.AddStruct(&update)
	if err != nil {
		panic(err)
	}

	err = ff.Parse(osArgsFF, os.Args[1:], ff.WithEnvVars())
	if err != nil {
		fmt.Println(ffhelp.Flags(osArgsFF))
		os.Exit(1)
	}
	if configuration.Help {
		fmt.Println(ffhelp.Flags(osArgsFF))
		os.Exit(0)
	}

	configuration.Security = security
	configuration.Controls = controls
	configuration.Update = update

	host, setAsEmpty := os.LookupEnv("HOST")
	if setAsEmpty {
		configuration.Address = host
	}
	return configuration
}

func (c *Config) DecompressPresentation() {
	if fileInfo, err := os.Stat(c.File); err == nil {
		if !fileInfo.IsDir() {
			if err := archive.UnGzip(c.File); err != nil {
				panic(err)
			}
		}
	}
}

func (c *Config) CompressPresentation() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fileName := c.Compress
	if !strings.HasSuffix(fileName, ".tar.gz") {
		fileName += ".tar.gz"
	}
	err = archive.Gzip(wd, fileName)
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}

func (c *Config) Git() {
	// https://github.com/oktalz/present.git
	if !strings.HasPrefix(c.GIT, "https://") {
		c.GIT = "https://" + c.GIT
	}
	if !strings.HasSuffix(c.GIT, ".git") {
		c.GIT += ".git"
	}
	// Create a temporary directory to extract to
	tmpDir, err := os.MkdirTemp("", "present_git_")
	if err != nil {
		panic(err)
	}
	log.Println("Created temporary directory:", tmpDir)
	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:      c.GIT,
		Progress: os.Stdout,
	})
	if err != nil {
		panic(err)
	}
	if c.Dir != "" {
		c.Dir = path.Join(tmpDir, c.Dir)
	} else {
		c.Dir = tmpDir
	}
}

func (c *Config) CheckPasswords() {
	if c.Security.AdminPwdDisable {
		log.Println("admin token is disabled")
		c.Security.AdminPwd = ""
	} else {
		if c.Security.AdminPwd != "" {
			log.Println("admin token is set ☢☢☢ 🙈🙉🙊")
		} else {
			c.Security.AdminPwd = ulid.Make().String()
			log.Println("admin token is not set, created one is:", c.Security.AdminPwd)
		}
	}
	if c.Security.UserPwd != "" {
		log.Println("user password is set")
	}
}
