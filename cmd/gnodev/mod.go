package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/gnolang/gnomod"
)

type modFlags struct {
	Verbose bool `flag:"verbose" help:"verbose"`
}

var defaultModFlags = modFlags{
	Verbose: false,
}

func modApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(modFlags)

	if len(args) != 1 {
		cmd.ErrPrintfln("Usage: mod [flags] <command>")
		return errors.New("invalid command")
	}

	switch args[0] {
	case "download":
		if err := runModDownload(&opts); err != nil {
			return fmt.Errorf("mod download: %w", err)
		}
	default:
		return fmt.Errorf("invalid command: %s", args[0])
	}

	return nil
}

func runModDownload(opts *modFlags) error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	modPath := filepath.Join(path, "gno.mod")
	if !gnomod.IsModFileExist(modPath) {
		return errors.New("gno.mod not found")
	}

	gnoMod, err := gnomod.ReadModFile(modPath)
	if err != nil {
		return fmt.Errorf("read mod file: %w", err)
	}

	if err := gnoMod.FetchDeps(); err != nil {
		return fmt.Errorf("read mod file: %w", err)
	}

	if err := gnomod.Sanitize(gnoMod); err != nil {
		return fmt.Errorf("read mod file: %w", err)
	}
	err = gnoMod.WriteToPath(path)
	if err != nil {
		return fmt.Errorf("read mod file: %w", err)
	}

	return nil
}
