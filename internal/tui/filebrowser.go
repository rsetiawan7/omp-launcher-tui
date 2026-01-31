package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) showFileBrowser(title string, onSelect func(path string), startPath string) {
	if startPath == "" {
		startPath = os.ExpandEnv("$HOME")
	}

	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitle(title + " (Esc to cancel)")

	var currentPath string
	var loadDir func(path string) error

	loadDir = func(path string) error {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		list.Clear()
		currentPath = path

		// Add parent directory option if not at root
		if path != "/" && path != "" {
			list.AddItem("..", "Parent Directory", 0, nil)
		}

		// Collect directories
		var dirs []os.DirEntry
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				dirs = append(dirs, entry)
			}
		}

		// Sort directories alphabetically
		sort.Slice(dirs, func(i, j int) bool {
			return dirs[i].Name() < dirs[j].Name()
		})

		// Add directories to list
		for _, dir := range dirs {
			list.AddItem(dir.Name(), "", 0, nil)
		}

		// Show current path in status
		a.layout.SetStatus("Current: " + currentPath)
		return nil
	}

	list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if mainText == ".." {
			// Go to parent directory
			parent := filepath.Dir(currentPath)
			_ = loadDir(parent)
		} else {
			// Select directory and close browser
			selected := filepath.Join(currentPath, mainText)
			onSelect(selected)
			a.app.SetRoot(a.layout.Root(), true)
			a.app.SetFocus(a.layout.Table())
			a.setKeybindings()
		}
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			a.app.SetRoot(a.layout.Root(), true)
			a.app.SetFocus(a.layout.Table())
			a.setKeybindings()
			return nil
		}
		return event
	})

	_ = loadDir(startPath)

	a.app.SetInputCapture(nil)
	a.app.SetRoot(list, true).SetFocus(list)
}
