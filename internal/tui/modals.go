package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rsetiawan7/omp-launcher-tui/internal/config"
	"github.com/rsetiawan7/omp-launcher-tui/internal/server"
)

func (a *App) promptSearch() {
	input := tview.NewInputField().SetLabel("Search: ").SetText(a.searchQuery)
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			a.searchQuery = input.GetText()
			if a.viewMode == ViewFavorites {
				a.applyFavoritesFilterAndSort()
				a.layout.UpdateTable(a.filteredFavorites)
			} else {
				a.applyFilterAndSort()
			}
		}
		a.setKeybindings()
		a.app.SetRoot(a.layout.Root(), true)
	})

	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(input, 3, 0, true)
	modal.SetBorder(true).SetTitle("Search (Enter to search, Esc to cancel)")

	// Clear global keybindings for modal
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.setKeybindings()
			a.app.SetRoot(a.layout.Root(), true)
			return nil
		}
		return event
	})

	a.app.SetRoot(modal, true).SetFocus(input)
}

func (a *App) promptPassword() {
	srv, ok := a.selectedServer()
	if !ok {
		return
	}
	input := tview.NewInputField().SetLabel("Password: ").SetMaskCharacter('*')
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			a.passwords[srv.Addr()] = input.GetText()
			a.launchServer(srv)
		} else {
			a.setKeybindings()
			a.app.SetRoot(a.layout.Root(), true)
		}
	})

	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(input, 3, 0, true)
	modal.SetBorder(true).SetTitle("Password (Enter to confirm, Esc to cancel)")

	// Clear global keybindings for modal
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.setKeybindings()
			a.app.SetRoot(a.layout.Root(), true)
			return nil
		}
		return event
	})

	a.app.SetRoot(modal, true).SetFocus(input)
}

func (a *App) showAddFavoriteDialog() {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Add Favorite Server (Enter to save, Esc to cancel)")

	var hostInput, portInput *tview.InputField

	hostInput = tview.NewInputField().
		SetLabel("Host (IP:Port or IP): ").
		SetFieldWidth(30).
		SetAcceptanceFunc(nil)

	portInput = tview.NewInputField().
		SetLabel("Port: ").
		SetFieldWidth(10).
		SetText("7777").
		SetAcceptanceFunc(tview.InputFieldInteger)

	form.AddFormItem(hostInput)
	form.AddFormItem(portInput)

	form.AddButton("Add", func() {
		hostText := strings.TrimSpace(hostInput.GetText())
		portText := strings.TrimSpace(portInput.GetText())

		if hostText == "" {
			a.layout.SetStatus("Host cannot be empty")
			return
		}

		// Parse host:port format if provided
		host := hostText
		port := 7777

		if strings.Contains(hostText, ":") {
			parts := strings.Split(hostText, ":")
			host = parts[0]
			if len(parts) > 1 {
				if p, err := strconv.Atoi(parts[1]); err == nil {
					port = p
				}
			}
		} else if portText != "" {
			if p, err := strconv.Atoi(portText); err == nil {
				port = p
			}
		}

		// Add to favorites
		if err := config.AddFavorite("", host, port); err != nil {
			a.layout.SetStatus(fmt.Sprintf("Failed to add favorite: %v", err))
			a.setKeybindings()
			a.app.SetRoot(a.layout.Root(), true)
			return
		}

		// Add to local favorites list
		a.favorites = append(a.favorites, server.Server{
			Name:    "",
			Host:    host,
			Port:    port,
			Loading: true,
		})

		a.layout.SetStatus(fmt.Sprintf("Added %s:%d to favorites", host, port))
		a.setKeybindings()
		a.app.SetRoot(a.layout.Root(), true)

		// Refresh favorites view
		if a.viewMode == ViewFavorites {
			a.applyFavoritesFilterAndSort()
			a.layout.UpdateTable(a.filteredFavorites)
			go a.refreshFavorites()
		}
	})

	form.AddButton("Cancel", func() {
		a.setKeybindings()
		a.app.SetRoot(a.layout.Root(), true)
	})

	// Handle escape key
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.setKeybindings()
			a.app.SetRoot(a.layout.Root(), true)
			return nil
		}
		return event
	})

	a.app.SetRoot(form, true).SetFocus(hostInput)
}
