package main

import (
	"context"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

type modDownloadCfg struct {
	remote  string
	verbose bool
}

func newModCmd(io *commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "mod",
			ShortUsage: "mod <command>",
			ShortHelp:  "Manage gno.mod",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newModDownloadCmd(io),
		newModInitCmd(),
		newModTidy(io),
	)

	return cmd
}

func newModDownloadCmd(io *commands.IO) *commands.Command {
	cfg := &modDownloadCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "download",
			ShortUsage: "download [flags]",
			ShortHelp:  "Download modules to local cache",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execModDownload(cfg, args, io)
		},
	)
}

func newModInitCmd() *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init [module-path]",
			ShortHelp:  "Initialize gno.mod file in current directory",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execModInit(args)
		},
	)
}

func newModTidy(io *commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "tidy",
			ShortUsage: "tidy",
			ShortHelp:  "Add missing and remove unused modules",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execModTidy(args, io)
		},
	)
}

func (c *modDownloadCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		"test3.gno.land:36657",
		"remote for fetching gno modules",
	)

	fs.BoolVar(
		&c.verbose,
		"verbose",
		false,
		"verbose output when running",
	)
}

func execModDownload(cfg *modDownloadCfg, args []string, io *commands.IO) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	modPath := filepath.Join(path, "gno.mod")
	if !isFileExist(modPath) {
		return errors.New("gno.mod not found")
	}

	// read gno.mod
	data, err := os.ReadFile(modPath)
	if err != nil {
		return fmt.Errorf("readfile %q: %w", modPath, err)
	}

	// parse gno.mod
	gnoMod, err := gnomod.Parse(modPath, data)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	// sanitize gno.mod
	gnoMod.Sanitize()

	// validate gno.mod
	if err := gnoMod.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	// fetch dependencies
	if err := gnoMod.FetchDeps(gnomod.GetGnoModPath(), cfg.remote, cfg.verbose); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	gomod, err := gnomod.GnoToGoMod(*gnoMod)
	if err != nil {
		return fmt.Errorf("sanitize: %w", err)
	}

	// write go.mod file
	err = gomod.Write(filepath.Join(path, "go.mod"))
	if err != nil {
		return fmt.Errorf("write go.mod file: %w", err)
	}

	return nil
}

func execModInit(args []string) error {
	if len(args) > 1 {
		return flag.ErrHelp
	}
	var modPath string
	if len(args) == 1 {
		modPath = args[0]
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := gnomod.CreateGnoModFile(dir, modPath); err != nil {
		return fmt.Errorf("create gno.mod file: %w", err)
	}

	return nil
}

func execModTidy(args []string, io *commands.IO) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	rd, err := gnomod.FindRootDir(wd)
	if err != nil {
		return err
	}
	fname := filepath.Join(rd, "gno.mod")

	gm, err := gnomod.ParseGnoMod(fname)
	if err != nil {
		return err
	}

	imports, err := getGnoImports(rd)
	if err != nil {
		return err
	}

	// TODO: Use gm.AddRequire()
	// Blocked by PR#1077
	var requires []*modfile.Require
	for _, im := range imports {
		requires = append(requires, &modfile.Require{
			Mod: module.Version{
				Path:    im,
				Version: "v0.0.0-latest",
			},
		})
	}
	gm.Require = requires

	gm.WriteToPath(fname)
	return nil
}

// getGnoImports returns the list of gno imports from a given path recursively.
// TODO: move this to better location.
func getGnoImports(path string) ([]string, error) {
	paths, err := gnoFilesFromArgs([]string{path})
	if err != nil {
		return nil, err
	}

	allImports := make([]string, 0)
	seen := make(map[string]struct{})
	for _, path := range paths {
		filename := filepath.Base(path)
		if strings.HasSuffix(filename, "_filetest.gno") {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		fs := token.NewFileSet()
		f, err := parser.ParseFile(fs, filename, data, parser.ImportsOnly)
		if err != nil {
			return nil, err
		}
		for _, imp := range f.Imports {
			importPath := strings.TrimPrefix(strings.TrimSuffix(imp.Path.Value, `"`), `"`)
			if !strings.HasPrefix(importPath, "gno.land/") {
				continue
			}
			if _, ok := seen[importPath]; ok {
				continue
			}
			allImports = append(allImports, importPath)
			seen[importPath] = struct{}{}
		}
	}
	sort.Strings(allImports)

	return allImports, nil
}
