package adb

import (
	"context"
	"sync"
	"time"
)

type Client interface {
	Version(ctx context.Context) (string, error)
	Devices(ctx context.Context) ([]Device, error)
	Packages(ctx context.Context, serial string, opts ListPackageOptions) ([]Package, error)
	Uninstall(ctx context.Context, serial, pkg string, keepData bool, user int) error
	Install(ctx context.Context, serial string, apkPath string) error // Updated signature
}

type Device struct {
	Serial       string `json:"device_id"`
	State        string `json:"status"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	IsAuthorized bool   `json:"authorized"`
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
