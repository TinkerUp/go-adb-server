package adb

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

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

func (adbServerClient *client) Version(ctx context.Context) (string, error) {
	versionCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()

	out, _, err := adbServerClient.run(versionCtx, "", "version")

	if err != nil {
		return "", fmt.Errorf("adb version failed: %w", err)
	}

	return strings.TrimSpace(out), nil
}

func (adbServerClient *client) Devices(ctx context.Context) ([]Device, error) {
	devicesContext, cancel := context.WithTimeout(ctx, adbServerClient.readTimeout)
	defer cancel()

	out, errOut, err := adbServerClient.run(devicesContext, "", "devices", "-l")

	if err != nil {
		return nil, fmt.Errorf("failed to list devices connect to adb: %v: %s", err, errOut)
	}

	var devices []Device

	scanner := bufio.NewScanner(strings.NewReader(out))

	for scanner.Scan() {
		scannedLineText := strings.TrimSpace(scanner.Text())

		if scannedLineText == "" || strings.HasPrefix(scannedLineText, "List of devices") {
			continue
		}

		// Since standard output for the adb devices -l command is:
		// List of devices attached
		// 33011JEHN19347         device product:lynx_beta model:Pixel_7a device:lynx transport_id:2
		// We split the output where there are more than one consecutive whitespaces using the Fields function
		fields := strings.Fields(scannedLineText)

		// This is just filtering unwanted data and lines of text
		if len(fields) < 2 {
			continue
		}

		serial, state := fields[0], fields[1]

		connectedDevice := Device{Serial: serial, State: state, IsAuthorized: state == "device"}

		for _, deviceFields := range fields[2:] {
			if after, ok := strings.CutPrefix(deviceFields, "model:"); ok {
				connectedDevice.Model = after
			}
		}

		devices = append(devices, connectedDevice)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return devices, nil
}

func (adbServerClient *client) Install(ctx context.Context, serial string, apk io.Reader, name string) error {
	if serial == "" {
		return errors.New("serial is required")
	}

	if name == "" {
		name = fmt.Sprintf("tinkerup-%d.apk", time.Now().UnixNano())
	}

	// I dont want to be accessing the user's computer willy nilly due to safety concerns
	// This is why we first copy the file to a temporary path and then access it
	temporaryDirPath := filepath.Join(adbServerClient.tempDir, name)

	temporaryDir, err := os.Create(temporaryDirPath)

	if err != nil {
		return fmt.Errorf("error creating temporary directory for apk storage: %w", err)
	}

	defer os.Remove(temporaryDirPath)

	_, copyErr := io.Copy(temporaryDir, apk)

	closeErr := temporaryDir.Close()

	if copyErr != nil {
		return fmt.Errorf("error while tryig to copy apk to temporary dir: %w", copyErr)
	}

	if closeErr != nil {
		return fmt.Errorf("close temp apk: %w", closeErr)
	}

	unlock := adbServerClient.lock(serial)
	defer unlock()

	// Pushing the APK file to the device
	remoteAPK := "/data/local/tmp/" + name

	{ // Creating local scope to introduce local variables
		apkPushCtx, cancel := context.WithTimeout(ctx, adbServerClient.installTimeout)

		defer cancel()

		if _, errOut, err := adbServerClient.run(apkPushCtx, serial, "push", temporaryDirPath, remoteAPK); err != nil {
			return fmt.Errorf("adb push failed: %v: %s", err, errOut)
		}
	}

	// Installing the pushed apk

	// Deferring cleanup of file from user device
	defer adbServerClient.run(context.Background(), serial, "shell", "rm", "-f", remoteAPK)

	{
		apkInstallCtx, cancel := context.WithTimeout(ctx, adbServerClient.installTimeout)

		defer cancel()

		out, errOut, err := adbServerClient.run(apkInstallCtx, serial, "shell", "pm", "install", "-r", remoteAPK)

		if err != nil {
			return fmt.Errorf("pm install failed: %v: %s", err, errOut)
		}

		if !strings.Contains(out, "Success") {
			return fmt.Errorf("install error: %s %s", strings.TrimSpace(out), strings.TrimSpace(errOut))
		}
	}

	return nil
}

func (adbServerClient *client) Uninstall(ctx context.Context, serial, pkg string, keepData bool, user int) error {
	var pkgNameRe = regexp.MustCompile(`^[a-zA-Z0-9._]+$`)

	if serial == "" {
		return errors.New("serial is required")
	}

	if !pkgNameRe.MatchString(pkg) {
		return fmt.Errorf("invalid package name")
	}

	unlock := adbServerClient.lock(serial)

	defer unlock()

	// Default to user 0
	var args []string

	if user >= 0 {
		args = []string{"shell", "pm", "uninstall", "--user", fmt.Sprint(user)}
	} else {
		args = []string{"shell", "pm", "uninstall"}
	}

	if keepData {
		args = append(args, "-k")
	}

	args = append(args, pkg)

	uninstallCtx, cancel := context.WithTimeout(ctx, adbServerClient.readTimeout)

	defer cancel()

	out, errOut, err := adbServerClient.run(uninstallCtx, serial, args...)

	if err != nil {
		return fmt.Errorf("uninstall failed: %v: %s", err, errOut)
	}

	// pm prints "Success" or "Failure [REASON]"
	if !strings.Contains(out, "Success") {
		if strings.Contains(out, "Failure") || errOut != "" {
			return fmt.Errorf("uninstall error: %s %s", strings.TrimSpace(out), strings.TrimSpace(errOut))
		}
	}

	return nil
}

func (adbServerClient *client) Packages(ctx context.Context, serial string, opts ListPackageOptions) ([]Package, error) {
	if serial == "" {
		return nil, errors.New("serial is required")
	}

	args := []string{"shell", "pm", "list", "packages", "-f"}

	if opts.IncludeUninstalled {
		args = append(args, "-u")
	}
	if opts.IncludeSystem {
		args = append(args, "-s")
	} else {
		args = append(args, "-3")
	}

	listPackagesContext, cancel := context.WithTimeout(ctx, adbServerClient.readTimeout)

	defer cancel()

	out, errOut, err := adbServerClient.run(listPackagesContext, serial, args...)

	if err != nil {
		return nil, fmt.Errorf("pm list packages failed: %v: %s", err, errOut)
	}

	var packages []Package

	scanner := bufio.NewScanner(strings.NewReader(out))

	for scanner.Scan() {
		scannedLineText := strings.TrimSpace(scanner.Text())

		if scannedLineText == "" {
			continue
		}

		// Expected: package:/data/app/..../base.apk=com.example
		eq := strings.LastIndex(scannedLineText, "=")

		if !strings.HasPrefix(scannedLineText, "package:") || eq == -1 {
			continue
		}

		apk := strings.TrimPrefix(scannedLineText[:eq], "package:")

		name := scannedLineText[eq+1:]

		packages = append(packages, Package{
			Name:     name,
			ApkPath:  apk,
			IsSystem: opts.IncludeSystem,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return packages, nil
}

func (adbServerClient *client) lock(serial string) func() {
	value, _ := adbServerClient.perSerialMu.LoadOrStore(serial, &sync.Mutex{})

	mutex := value.(*sync.Mutex)

	mutex.Lock()
	return mutex.Unlock
}

func (adbServerClient *client) run(ctx context.Context, serial string, args ...string) (stdout string, stderr string, err error) {
	argumentsArray := make([]string, 0, len(args)+2) // Creating an array of strings for arguments

	if serial != "" {
		argumentsArray = append(argumentsArray, "-s", serial) // Adding in the serial number of the device using the "-s" flag for commands where we need it
	}

	argumentsArray = append(argumentsArray, args...)

	adbCommand := exec.CommandContext(ctx, adbServerClient.adbPath, argumentsArray...)

	var outBuf, errorBuf bytes.Buffer

	adbCommand.Stdout = &outBuf
	adbCommand.Stderr = &errorBuf

	err = adbCommand.Run()
	return outBuf.String(), errorBuf.String(), err
}
