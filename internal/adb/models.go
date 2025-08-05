package adb

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"
)

type Client interface {
	Version(ctx context.Context) (string, error)
	Devices(ctx context.Context) ([]Device, error)
	Packages(ctx context.Context, serial string, opts ListPackageOptions) ([]Package, error)
	Uninstall(ctx context.Context, serial, pkg string, keepData bool, user int) error
	Install(ctx context.Context, serial string, apk io.Reader, name string) error
}

type Device struct {
	Serial       string
	State        string // device, unauthorized, offline
	Model        string // optional, may be empty in MVP
	Manufacturer string // optional
	IsAuthorized bool
}

type Package struct {
	Name     string
	ApkPath  string
	IsSystem bool
}

type ListPackageOptions struct {
	IncludeSystem      bool // false => user apps only
	IncludeUninstalled bool // adds -u
}

type Config struct {
	ADBPath        string
	ReadTimeout    time.Duration // generic list timeout
	InstallTimeout time.Duration
	TempDir        string
}

type client struct {
	adbPath        string
	readTimeout    time.Duration
	installTimeout time.Duration
	tempDir        string

	perSerialMu sync.Map // map[string]*sync.Mutex, serialize installs/uninstalls per device
}

func New(cfg Config) (Client, error) {
	if cfg.ADBPath == "" {
		return nil, errors.New("adb path is required")
	}

	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 15 * time.Second
	}

	if cfg.InstallTimeout == 0 {
		cfg.InstallTimeout = 8 * time.Minute
	}

	if cfg.TempDir == "" {
		cfg.TempDir = os.TempDir()
	}

	return &client{
		adbPath:        cfg.ADBPath,
		readTimeout:    cfg.ReadTimeout,
		installTimeout: cfg.InstallTimeout,
		tempDir:        cfg.TempDir,
	}, nil
}

func (adbServerClient *client) Devices(ctx context.Context) ([]Device, error) {
	panic("unimplemented")
}

func (adbServerClient *client) Install(ctx context.Context, serial string, apk io.Reader, name string) error {
	panic("unimplemented")
}

func (adbServerClient *client) Packages(ctx context.Context, serial string, opts ListPackageOptions) ([]Package, error) {
	panic("unimplemented")
}

func (adbServerClient *client) Uninstall(ctx context.Context, serial string, pkg string, keepData bool, user int) error {
	panic("unimplemented")
}

func (adbServerClient *client) Version(ctx context.Context) (string, error) {
	panic("unimplemented")
}
