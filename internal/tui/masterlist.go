package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rsetiawan7/omp-launcher-tui/internal/config"
	"github.com/rsetiawan7/omp-launcher-tui/internal/server"
)

func (a *App) showMasterListManager() {
	lists, err := config.LoadMasterLists()
	if err != nil {
		a.app.QueueUpdateDraw(func() {
			a.layout.SetStatus(fmt.Sprintf("Failed to load master lists: %v", err))
		})
		return
	}

	table := tview.NewTable().SetSelectable(true, false)
	table.SetBorder(true).SetTitle("Manage Master Server Lists (Enter: Edit | A: Add | D: Delete | S: Set Active | Esc: Back)")
	table.SetBordersColor(tcell.ColorWhite)

	updateTable := func() {
		table.Clear()
		// Header
		headers := []string{"Active", "Name", "Host", "Description"}
		for i, h := range headers {
			cell := tview.NewTableCell(fmt.Sprintf("[::b]%s", h)).
				SetSelectable(false).
				SetExpansion(1)
			table.SetCell(0, i, cell)
		}
		table.SetFixed(1, 0)

		// Data rows
		for i, list := range lists.Lists {
			row := i + 1
			active := " "
			if list.Active {
				active = "✓"
			}
			table.SetCell(row, 0, tview.NewTableCell(active).SetExpansion(1))
			table.SetCell(row, 1, tview.NewTableCell(list.Name).SetExpansion(2))
			table.SetCell(row, 2, tview.NewTableCell(list.Host).SetExpansion(3))
			table.SetCell(row, 3, tview.NewTableCell(list.Description).SetExpansion(2))
		}

		if len(lists.Lists) == 0 {
			table.SetCell(1, 0, tview.NewTableCell("No master lists. Press 'A' to add one.").SetSelectable(false))
		}
	}

	updateTable()

	// Clear app-level keybindings while in master list manager
	a.app.SetInputCapture(nil)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()
		idx := row - 1

		switch event.Key() {
		case tcell.KeyEscape:
			a.setKeybindings()
			a.app.SetRoot(a.layout.Root(), true)
			return nil

		case tcell.KeyEnter:
			if idx >= 0 && idx < len(lists.Lists) {
				a.editMasterList(&lists, idx, updateTable)
			}
			return nil

		case tcell.KeyRune:
			switch event.Rune() {
			case 'a', 'A':
				a.addMasterList(&lists, updateTable)
				return nil

			case 'd', 'D':
				if idx >= 0 && idx < len(lists.Lists) {
					// Remove the item
					lists.Lists = append(lists.Lists[:idx], lists.Lists[idx+1:]...)
					if err := config.SaveMasterLists(lists); err != nil {
						a.layout.SetStatus(fmt.Sprintf("Failed to save: %v", err))
					}
					updateTable()
					a.layout.SetStatus("Master list deleted")
				}
				return nil

			case 's', 'S':
				if idx >= 0 && idx < len(lists.Lists) {
					// Set active
					for i := range lists.Lists {
						lists.Lists[i].Active = (i == idx)
					}
					if err := config.SaveMasterLists(lists); err != nil {
						a.layout.SetStatus(fmt.Sprintf("Failed to save: %v", err))
					}
					// Update config
					a.cfg.MasterServer = lists.Lists[idx].Host
					config.Save(a.cfg)
					updateTable()
					a.layout.SetStatus(fmt.Sprintf("Active master list: %s", lists.Lists[idx].Name))
				}
				return nil
			}
		}
		return event
	})

	a.app.SetRoot(table, true).SetFocus(table)
}

func (a *App) addMasterList(lists *config.MasterLists, updateTable func()) {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Add Master Server List")

	var name, host, description string
	statusText := tview.NewTextView().SetDynamicColors(true)
	statusText.SetText("")

	form.AddInputField("Name:", "", 40, nil, func(text string) {
		name = text
	})
	form.AddInputField("Host:", "https://", 60, nil, func(text string) {
		host = text
	})
	form.AddTextArea("Description:", "", 60, 3, 0, func(text string) {
		description = text
	})

	form.AddButton("Test", func() {
		if host == "" {
			statusText.SetText("[red]Host is required")
			return
		}

		statusText.SetText("[yellow]Testing connection...")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.TestMasterServer(ctx, host); err != nil {
				a.app.QueueUpdateDraw(func() {
					statusText.SetText(fmt.Sprintf("[red]Test failed: %v", err))
				})
			} else {
				a.app.QueueUpdateDraw(func() {
					statusText.SetText("[green]✓ Connection successful! Valid server list found.")
				})
			}
		}()
	})

	form.AddButton("Save", func() {
		if name == "" || host == "" {
			statusText.SetText("[red]Name and Host are required")
			return
		}

		lists.Lists = append(lists.Lists, config.MasterList{
			Name:        name,
			Host:        host,
			Description: description,
			Active:      false,
		})

		if err := config.SaveMasterLists(*lists); err != nil {
			a.layout.SetStatus(fmt.Sprintf("Failed to save: %v", err))
			return
		}

		updateTable()
		a.showMasterListManager()
	})

	form.AddButton("Cancel", func() {
		a.showMasterListManager()
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.showMasterListManager()
			return nil
		}
		return event
	})

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(statusText, 1, 0, false)

	// Clear app-level keybindings to prevent interference
	a.app.SetInputCapture(nil)
	a.app.SetRoot(layout, true).SetFocus(form)
}

func (a *App) editMasterList(lists *config.MasterLists, idx int, updateTable func()) {
	if idx < 0 || idx >= len(lists.Lists) {
		return
	}

	list := lists.Lists[idx]
	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Edit Master Server List")

	statusText := tview.NewTextView().SetDynamicColors(true)
	statusText.SetText("")

	form.AddInputField("Name:", list.Name, 40, nil, func(text string) {
		list.Name = text
	})
	form.AddInputField("Host:", list.Host, 60, nil, func(text string) {
		list.Host = text
	})
	form.AddTextArea("Description:", list.Description, 60, 3, 0, func(text string) {
		list.Description = text
	})

	form.AddButton("Test", func() {
		if list.Host == "" {
			statusText.SetText("[red]Host is required")
			return
		}

		statusText.SetText("[yellow]Testing connection...")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.TestMasterServer(ctx, list.Host); err != nil {
				a.app.QueueUpdateDraw(func() {
					statusText.SetText(fmt.Sprintf("[red]Test failed: %v", err))
				})
			} else {
				a.app.QueueUpdateDraw(func() {
					statusText.SetText("[green]✓ Connection successful! Valid server list found.")
				})
			}
		}()
	})

	form.AddButton("Save", func() {
		if list.Name == "" || list.Host == "" {
			statusText.SetText("[red]Name and Host are required")
			return
		}

		lists.Lists[idx] = list

		if err := config.SaveMasterLists(*lists); err != nil {
			a.layout.SetStatus(fmt.Sprintf("Failed to save: %v", err))
			return
		}

		// Update config if this is the active one
		if list.Active {
			a.cfg.MasterServer = list.Host
			config.Save(a.cfg)
		}

		updateTable()
		a.showMasterListManager()
	})

	form.AddButton("Cancel", func() {
		a.showMasterListManager()
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.showMasterListManager()
			return nil
		}
		return event
	})

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 0, 1, true).
		AddItem(statusText, 1, 0, false)

	// Clear app-level keybindings to prevent interference
	a.app.SetInputCapture(nil)
	a.app.SetRoot(layout, true).SetFocus(form)
}
