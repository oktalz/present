package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/joho/godotenv"
	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/version"
)

//go:embed ui/static
var dist embed.FS

//go:embed ui/login.html
var loginPage []byte

func main() { //nolint:funlen
	_ = godotenv.Load("present.env")
	_ = godotenv.Overload(".env")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	globalENV := path.Join(homeDir, ".present", "present.env")
	_ = godotenv.Overload(globalENV)
	_ = version.Set()
	// f, err := os.Create("cpu.pprof")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer f.Close()
	// _ = pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// f, err := os.Create("mem.pprof")
	// if err != nil {
	// 	log.Fatal("could not create memory profile: ", err)
	// }
	// defer f.Close()
	// defer pprof.WriteHeapProfile(f)

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stderr) // why do packages feel the need to change this in init()? :(

	config := configuration.Get()

	if config.Version {
		fmt.Println("present", version.Version)
		fmt.Println("built-from", version.Repo)
		if version.CommitDate != "" {
			fmt.Println("commit-date", version.CommitDate)
		}
		os.Exit(0)
	}

	if config.Tag {
		fmt.Println(version.Tag)
		os.Exit(0)
	}

	if config.Update.Latest {
		if config.Update.UpdateToLatest() {
			os.Exit(0)
		}
		os.Exit(1)
	}

	fmt.Println("present", version.Version)

	if config.Compress != "" {
		config.CompressPresentation()
	}

	if config.File != "" {
		config.DecompressPresentation()
	}

	if config.GIT != "" {
		config.Git()
	}

	if config.Dir != "" { //nolint:nestif
		err := os.Chdir(config.Dir)
		if err != nil {
			panic(err)
		}
	}

	config.CheckPasswords()

	startServer(config)
}
