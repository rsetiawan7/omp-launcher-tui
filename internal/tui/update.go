package tui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rivo/tview"
	"golang.org/x/mod/semver"
)

type UpdateChecker interface {
	Check(ctx context.Context, current string) (*Release, error)
	Download(ctx context.Context, asset ReleaseAsset) (string, error)
}

type GitHubChecker struct {
	Owner string
	Repo  string
}

type Release struct {
	Tag     string
	Name    string
	Asset   ReleaseAsset
	URL     string
	Summary string
}

type ReleaseAsset struct {
	Name string
	URL  string
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (g GitHubChecker) Check(ctx context.Context, current string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", g.Owner, g.Repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release check failed: %s", resp.Status)
	}
	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	latest := normalizeVersion(release.TagName)
	if latest == "" || !semver.IsValid(latest) {
		return nil, errors.New("invalid release version")
	}
	current = normalizeVersion(current)
	if current != "" && semver.IsValid(current) && semver.Compare(latest, current) <= 0 {
		return nil, nil
	}

	asset, err := selectAsset(release.Assets)
	if err != nil {
		return nil, err
	}

	return &Release{
		Tag:     release.TagName,
		Name:    release.Name,
		Asset:   asset,
		URL:     release.HTMLURL,
		Summary: release.Body,
	}, nil
}

func (g GitHubChecker) Download(ctx context.Context, asset ReleaseAsset) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}
	file, err := os.CreateTemp("", "omp-tui-*")
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", err
	}
	return file.Name(), nil
}

func selectAsset(assets []struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}) (ReleaseAsset, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, osName) && strings.Contains(name, arch) {
			return ReleaseAsset{Name: asset.Name, URL: asset.BrowserDownloadURL}, nil
		}
	}
	if len(assets) > 0 {
		return ReleaseAsset{Name: assets[0].Name, URL: assets[0].BrowserDownloadURL}, nil
	}
	return ReleaseAsset{}, errors.New("no release assets available")
}

func normalizeVersion(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return ""
	}
	if !strings.HasPrefix(tag, "v") {
		return "v" + tag
	}
	return tag
}

func (a *App) checkForUpdates() {
	if a.updateChecker == nil {
		a.layout.SetStatus("Update checker not configured")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	go func() {
		release, err := a.updateChecker.Check(ctx, a.version)
		if err != nil {
			a.app.QueueUpdateDraw(func() {
				a.layout.SetStatus(fmt.Sprintf("Update check failed: %v", err))
			})
			return
		}
		if release == nil {
			a.app.QueueUpdateDraw(func() {
				a.layout.SetStatus("Already up to date")
			})
			return
		}
		prompt := fmt.Sprintf("Update available: %s. Download now?", release.Tag)
		a.app.QueueUpdateDraw(func() {
			a.showUpdatePrompt(prompt, *release)
		})
	}()
}

func (a *App) showUpdatePrompt(message string, release Release) {
	modal := tview.NewModal().SetText(message).AddButtons([]string{"Download", "Cancel"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Download" {
			a.downloadAndReplace(release)
			return
		}
		a.app.SetRoot(a.layout.Root(), true)
	})
	a.app.SetRoot(modal, true)
}

func (a *App) downloadAndReplace(release Release) {
	a.layout.SetStatus("Downloading update...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		path, err := a.updateChecker.Download(ctx, release.Asset)
		if err != nil {
			a.app.QueueUpdateDraw(func() {
				a.layout.SetStatus(fmt.Sprintf("Download failed: %v", err))
				a.app.SetRoot(a.layout.Root(), true)
			})
			return
		}
		exe, err := os.Executable()
		if err != nil {
			a.app.QueueUpdateDraw(func() {
				a.layout.SetStatus(fmt.Sprintf("Update failed: %v", err))
				a.app.SetRoot(a.layout.Root(), true)
			})
			return
		}
		target := exe
		if runtime.GOOS == "windows" {
			target = exe + ".new"
		}
		if err := os.Rename(path, target); err != nil {
			if copyErr := copyFile(path, target); copyErr != nil {
				a.app.QueueUpdateDraw(func() {
					a.layout.SetStatus(fmt.Sprintf("Update replace failed: %v", err))
					a.app.SetRoot(a.layout.Root(), true)
				})
				return
			}
			_ = os.Remove(path)
		}
		if runtime.GOOS != "windows" {
			_ = os.Chmod(target, 0o755)
		}
		message := "Update downloaded. Please restart."
		if runtime.GOOS == "windows" {
			message = "Update downloaded as .new. Please replace and restart."
		}
		a.app.QueueUpdateDraw(func() {
			a.layout.SetStatus(message)
			a.app.SetRoot(a.layout.Root(), true)
		})
	}()
}

func (a *App) showUpdateInfo(release Release) {
	text := fmt.Sprintf("%s\n\n%s", release.Tag, release.URL)
	modal := tview.NewModal().SetText(text).AddButtons([]string{"OK"})
	modal.SetDoneFunc(func(_ int, _ string) {
		a.app.SetRoot(a.layout.Root(), true)
	})
	a.app.SetRoot(modal, true)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
