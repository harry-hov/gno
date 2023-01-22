package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/gnomod"
)

type modFlags struct {
	Verbose bool `flag:"verbose" help:"verbose"`
}

var defaultModFlags = modFlags{
	Verbose: false,
}

func modApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(modFlags)

	if len(args) != 1 || args[0] != "download" {
		cmd.ErrPrintfln("Usage: mod download [flags]")
		return errors.New("invalid command")
	}

	if err := runModDownload(&opts); err != nil {
		return fmt.Errorf("mod download: %s", err)
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
		return err
	}

	if err := gnomod.FetchModPackages(gnoMod); err != nil {
		return err
	}

	return nil
}
