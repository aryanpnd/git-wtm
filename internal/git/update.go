package git

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type UpdateInfo struct {
	NewVersion  string
	UpgradeCmd  string
	InstallMethod string
}

func CheckForUpdate(currentVersion string) (*UpdateInfo, error) {
	if currentVersion == "dev" || currentVersion == "" {
		return nil, nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/aryanpnd/git-wtm/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	if latest == current || latest == "" {
		return nil, nil
	}

	method := detectInstallMethod()
	cmd := upgradeCommand(method)

	return &UpdateInfo{
		NewVersion:    latest,
		UpgradeCmd:    cmd,
		InstallMethod: method,
	}, nil
}

func detectInstallMethod() string {
	exe, err := os.Executable()
	if err != nil {
		return "unknown"
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "unknown"
	}
	exe = filepath.ToSlash(exe)

	switch runtime.GOOS {
	case "darwin", "linux":
		switch {
		case strings.Contains(exe, "homebrew") || strings.Contains(exe, "linuxbrew"):
			return "brew"
		case strings.Contains(exe, "/snap/"):
			return "snap"
		case strings.Contains(exe, "/nix/"):
			return "nix"
		case strings.Contains(exe, "/go/bin/"):
			return "go"
		default:
			return "binary"
		}
	case "windows":
		switch {
		case strings.Contains(exe, "\\scoop\\"):
			return "scoop"
		case strings.Contains(exe, "\\chocolatey\\") || strings.Contains(exe, "\\choco\\"):
			return "choco"
		case strings.Contains(exe, "\\winget\\"):
			return "winget"
		case strings.Contains(exe, "\\go\\bin\\"):
			return "go"
		default:
			return "binary"
		}
	}
	return "unknown"
}

func upgradeCommand(method string) string {
	switch method {
	case "brew":
		return "brew upgrade aryanpnd/tap/git-wtm"
	case "scoop":
		return "scoop update git-wtm"
	case "choco":
		return "choco upgrade git-wtm"
	case "winget":
		return "winget upgrade aryanpnd.git-wtm"
	case "snap":
		return "sudo snap refresh git-wtm"
	case "go":
		return "go install github.com/aryanpnd/git-wtm@latest"
	default:
		return fmt.Sprintf("download from https://github.com/aryanpnd/git-wtm/releases/latest")
	}
}
