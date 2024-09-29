package version

import (
	"errors"
	"runtime/debug"
	"strings"
)

var (
	Repo       = ""
	Version    = "dev"
	Tag        = "dev"
	CommitDate = ""
)

var ErrBuildDataNotReadable = errors.New("not able to read build data")

func Set() error {
	buildinfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ErrBuildDataNotReadable
	}
	Repo = buildinfo.Main.Path
	CommitDate = get(buildinfo, "vcs.time")

	extra := ""
	commit := get(buildinfo, "vcs.revision")
	if len(commit) > 8 {
		commit = commit[:8]
		extra = "-" + commit
	}

	if get(buildinfo, "vcs.modified") == "true" {
		extra += "-dirty"
	}

	Version = strings.Replace(buildinfo.Main.Version, "(devel)", "dev", 1) + extra
	Tag = strings.Replace(buildinfo.Main.Version, "(devel)", "dev", 1)

	return nil
}

func get(buildInfo *debug.BuildInfo, key string) string {
	for _, setting := range buildInfo.Settings {
		if setting.Key == key {
			return setting.Value
		}
	}
	return ""
}
