package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/izumin5210/gex/pkg/manifest"
	"github.com/spf13/afero"
)

const (
	manifestName = "tools.go"
	binDirName   = "bin"
)

var (
	pkgToBeAdded = flag.String("add", "", "Add new tools")
)

func main() {
	var exitCode int

	if err := run(); err != nil {
		exitCode = 1
	}

	os.Exit(exitCode)
}

type config struct {
	outW, errW   io.Writer
	inR          io.Reader
	fs           afero.Fs
	workingDir   string
	manifestName string
	binDirName   string
}

func (c *config) ManifestPath() string {
	return filepath.Join(c.workingDir, c.manifestName)
}

func (c *config) BinDir() string {
	return filepath.Join(c.workingDir, c.binDirName)
}

func run() error {
	var (
		cfg = &config{
			outW:         os.Stdout,
			errW:         os.Stderr,
			inR:          os.Stdin,
			fs:           afero.NewOsFs(),
			manifestName: manifestName,
			binDirName:   binDirName,
		}
		ctx = context.TODO()
		err error
	)

	cfg.workingDir, err = os.Getwd()
	if err != nil {
		return err
	}

	flag.Parse()
	args := flag.Args()

	switch {
	case *pkgToBeAdded != "":
		// TODO: not yet implemented.
	case len(args) > 0:
		runTool(ctx, args[0], args[1:], cfg)
	default:
		err = errors.New("invalid arguments")
	}

	return err
}

func runTool(ctx context.Context, name string, args []string, cfg *config) error {
	p := manifest.NewParser(cfg.fs)
	m, err := p.Parse(cfg.ManifestPath())
	if err != nil {
		return err
	}

	t, ok := m.FindTool(name)
	if !ok {
		return fmt.Errorf("failed to find the tool %q", name)
	}

	bin := filepath.Join(cfg.BinDir(), name)

	if st, err := cfg.fs.Stat(bin); err != nil {
		// build
		cmd := exec.CommandContext(ctx, "go", "build", "-v", "-o", bin, string(t))
		cmd.Stdout = cfg.outW
		cmd.Stderr = cfg.errW
		cmd.Stdin = cfg.inR
		err = cmd.Run()
		if err != nil {
			return err
		}
	} else if st.IsDir() {
		return fmt.Errorf("%q is a directory", bin)
	}

	// exec
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = cfg.outW
	cmd.Stderr = cfg.errW
	cmd.Stdin = cfg.inR
	return cmd.Run()
}