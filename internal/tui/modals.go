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
			a.updateFilterPanel()
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

func (a *App) promptAlias(srv server.Server) {
	// Validation function for alias: alphanumeric, dash, underscore only
	aliasValidator := func(text string, ch rune) bool {
		return (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' ||
			ch == '_'
	}

	input := tview.NewInputField().
		SetLabel("Alias (a-z, 0-9, -, _): ").
		SetFieldWidth(40).
		SetAcceptanceFunc(aliasValidator)

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			alias := strings.TrimSpace(input.GetText())

			// Check alias uniqueness if provided
			if alias != "" && !config.IsAliasUnique(alias, srv.Host, srv.Port) {
				a.layout.SetStatus("Error: Alias already exists")
				a.setKeybindings()
				a.app.SetRoot(a.layout.Root(), true)
				return
			}

			// Add to favorites with alias
			if err := config.AddFavorite(srv.Name, alias, srv.Host, srv.Port); err != nil {
				a.layout.SetStatus(fmt.Sprintf("Failed to add favorite: %v", err))
				a.setKeybindings()
				a.app.SetRoot(a.layout.Root(), true)
				return
			}

			displayName := srv.Name
			if alias != "" {
				displayName = alias
			}
			a.layout.SetStatus(fmt.Sprintf("Added %s to favorites", displayName))

			// Update local favorites list
			a.favorites = append(a.favorites, srv)
			a.applyFavoritesFilterAndSort()
		}

		a.setKeybindings()
		a.app.SetRoot(a.layout.Root(), true)
	})

	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(input, 3, 0, true)
	modal.SetBorder(true).SetTitle(fmt.Sprintf("Add '%s' to Favorites", srv.Name))

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

	var aliasInput, hostInput, portInput *tview.InputField

	// Validation function for alias: alphanumeric, dash, underscore only
	aliasValidator := func(text string, ch rune) bool {
		return (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' ||
			ch == '_'
	}

	aliasInput = tview.NewInputField().
		SetLabel("Alias (a-z, 0-9, -, _): ").
		SetFieldWidth(30).
		SetAcceptanceFunc(aliasValidator)

	hostInput = tview.NewInputField().
		SetLabel("Host (IP:Port or IP): ").
		SetFieldWidth(30).
		SetAcceptanceFunc(nil)

	portInput = tview.NewInputField().
		SetLabel("Port: ").
		SetFieldWidth(10).
		SetText("7777").
		SetAcceptanceFunc(tview.InputFieldInteger)

	form.AddFormItem(aliasInput)
	form.AddFormItem(hostInput)
	form.AddFormItem(portInput)

	form.AddButton("Add", func() {
		aliasText := strings.TrimSpace(aliasInput.GetText())
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

		// Check alias uniqueness if provided
		if aliasText != "" && !config.IsAliasUnique(aliasText, host, port) {
			a.layout.SetStatus("Error: Alias already exists")
			return
		}

		// Add to favorites
		if err := config.AddFavorite("", aliasText, host, port); err != nil {
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

		displayName := aliasText
		if displayName == "" {
			displayName = fmt.Sprintf("%s:%d", host, port)
		}
		a.layout.SetStatus(fmt.Sprintf("Added %s to favorites", displayName))
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

func (a *App) showVersionFilterDialog() {
	versions := a.collectAvailableVersions()

	if len(versions) == 0 {
		a.layout.SetStatus("No server versions available (servers may not have rules loaded yet)")
		return
	}

	// Sort versions for consistent display
	// Simple alphabetical sort
	for i := 0; i < len(versions); i++ {
		for j := i + 1; j < len(versions); j++ {
			if versions[i] > versions[j] {
				versions[i], versions[j] = versions[j], versions[i]
			}
		}
	}

	list := tview.NewList()
	list.SetBorder(true).SetTitle("Version Filter (Space to toggle, Enter to apply, Esc to cancel)")

	// Track temporary filter state
	tempFilters := make(map[string]bool)
	for k, v := range a.versionFilters {
		tempFilters[k] = v
	}

	// Add all versions to the list
	for _, version := range versions {
		ver := version // capture for closure
		selected := tempFilters[ver]
		text := version
		if selected {
			text = "[X] " + text
		} else {
			text = "[ ] " + text
		}

		list.AddItem(text, "", 0, func() {
			// Toggle selection
			tempFilters[ver] = !tempFilters[ver]

			// Update list item text
			for i := 0; i < list.GetItemCount(); i++ {
				itemText, _ := list.GetItemText(i)
				itemVersion := strings.TrimPrefix(strings.TrimPrefix(itemText, "[X] "), "[ ] ")
				if itemVersion == ver {
					if tempFilters[ver] {
						list.SetItemText(i, "[X] "+itemVersion, "")
					} else {
						list.SetItemText(i, "[ ] "+itemVersion, "")
					}
					break
				}
			}
		})
	}

	// Add buttons at the bottom
	buttons := tview.NewFlex().SetDirection(tview.FlexColumn)

	applyBtn := tview.NewButton("Apply (Enter)")
	applyBtn.SetSelectedFunc(func() {
		// Apply filters
		a.versionFilters = make(map[string]bool)
		for k, v := range tempFilters {
			if v {
				a.versionFilters[k] = v
			}
		}

		// Update filter panel
		a.updateFilterPanel()

		// Reapply filters
		if a.viewMode == ViewFavorites {
			a.applyFavoritesFilterAndSort()
			a.layout.UpdateTable(a.filteredFavorites)
		} else {
			a.applyFilterAndSort()
		}

		a.setKeybindings()
		a.app.SetRoot(a.layout.Root(), true)

		activeCount := 0
		for _, v := range a.versionFilters {
			if v {
				activeCount++
			}
		}
		a.layout.SetStatus(fmt.Sprintf("Applied %d version filter(s)", activeCount))
	})

	clearBtn := tview.NewButton("Clear All")
	clearBtn.SetSelectedFunc(func() {
		// Clear all filters
		tempFilters = make(map[string]bool)

		// Update list
		for i := 0; i < list.GetItemCount(); i++ {
			itemText, _ := list.GetItemText(i)
			itemVersion := strings.TrimPrefix(strings.TrimPrefix(itemText, "[X] "), "[ ] ")
			list.SetItemText(i, "[ ] "+itemVersion, "")
		}
	})

	cancelBtn := tview.NewButton("Cancel (Esc)")
	cancelBtn.SetSelectedFunc(func() {
		a.setKeybindings()
		a.app.SetRoot(a.layout.Root(), true)
	})

	buttons.AddItem(applyBtn, 0, 1, false)
	buttons.AddItem(clearBtn, 0, 1, false)
	buttons.AddItem(cancelBtn, 0, 1, false)

	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.AddItem(list, 0, 1, true)
	modal.AddItem(buttons, 3, 0, false)

	// Handle keyboard input
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.setKeybindings()
			a.app.SetRoot(a.layout.Root(), true)
			return nil
		} else if event.Key() == tcell.KeyEnter {
			// Apply filters
			a.versionFilters = make(map[string]bool)
			for k, v := range tempFilters {
				if v {
					a.versionFilters[k] = v
				}
			}

			// Update filter panel
			a.updateFilterPanel()

			// Reapply filters
			if a.viewMode == ViewFavorites {
				a.applyFavoritesFilterAndSort()
				a.layout.UpdateTable(a.filteredFavorites)
			} else {
				a.applyFilterAndSort()
			}

			a.setKeybindings()
			a.app.SetRoot(a.layout.Root(), true)

			activeCount := 0
			for _, v := range a.versionFilters {
				if v {
					activeCount++
				}
			}
			a.layout.SetStatus(fmt.Sprintf("Applied %d version filter(s)", activeCount))
			return nil
		} else if event.Rune() == ' ' {
			// Toggle current item
			currentItem := list.GetCurrentItem()
			if currentItem >= 0 && currentItem < list.GetItemCount() {
				itemText, _ := list.GetItemText(currentItem)
				itemVersion := strings.TrimPrefix(strings.TrimPrefix(itemText, "[X] "), "[ ] ")
				tempFilters[itemVersion] = !tempFilters[itemVersion]

				if tempFilters[itemVersion] {
					list.SetItemText(currentItem, "[X] "+itemVersion, "")
				} else {
					list.SetItemText(currentItem, "[ ] "+itemVersion, "")
				}
			}
			return nil
		} else if event.Rune() == 'c' || event.Rune() == 'C' {
			// Clear all filters
			tempFilters = make(map[string]bool)

			// Update list
			for i := 0; i < list.GetItemCount(); i++ {
				itemText, _ := list.GetItemText(i)
				itemVersion := strings.TrimPrefix(strings.TrimPrefix(itemText, "[X] "), "[ ] ")
				list.SetItemText(i, "[ ] "+itemVersion, "")
			}
			return nil
		}
		return event
	})

	a.app.SetRoot(modal, true).SetFocus(list)
}
