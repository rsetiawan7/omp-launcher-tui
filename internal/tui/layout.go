package tui

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/rsetiawan7/omp-launcher-tui/internal/server"
)

type Layout struct {
	root        *tview.Flex
	table       *tview.Table
	players     *tview.Table
	rules       *tview.Table
	pingChart   *tview.TextView
	status      *tview.TextView
	keys        *tview.TextView
	filterPanel *tview.TextView
	statusBar   *tview.Flex
	onSelect    func(row int)
}

func NewLayout() *Layout {
	table := tview.NewTable().SetSelectable(true, false)
	table.SetBorder(true).SetTitle("Servers")
	table.SetBordersColor(tcell.ColorWhite)
	table.SetSeparator(tview.Borders.Vertical)

	players := tview.NewTable().SetSelectable(true, false)
	players.SetBorder(true).SetTitle("Players")
	players.SetBordersColor(tcell.ColorWhite)
	rules := tview.NewTable().SetSelectable(false, false)
	rules.SetBorder(true).SetTitle("Server Rules")
	rules.SetBordersColor(tcell.ColorWhite)
	pingChart := tview.NewTextView().SetDynamicColors(false)
	pingChart.SetBorder(true).SetTitle("Ping History")
	status := tview.NewTextView().SetDynamicColors(true)
	status.SetText("Ready")
	keys := tview.NewTextView().SetDynamicColors(true)
	keys.SetText(StatusKeys)
	filterPanel := tview.NewTextView().SetDynamicColors(true)
	filterPanel.SetBorder(true).SetTitle("Filters (V to toggle)")
	filterPanel.SetText("No version filters active")

	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	rightPanel.AddItem(players, 0, 1, false)
	rightPanel.AddItem(rules, 0, 1, false)
	rightPanel.AddItem(pingChart, 8, 0, false)

	main := tview.NewFlex().SetDirection(tview.FlexColumn)
	main.AddItem(table, 0, 3, true)
	main.AddItem(rightPanel, 0, 2, false)

	statusBar := tview.NewFlex().SetDirection(tview.FlexColumn)
	statusBar.AddItem(status, 0, 1, false)
	statusBar.AddItem(keys, 0, 2, false)

	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.AddItem(main, 0, 1, true)
	root.AddItem(filterPanel, 3, 0, false)
	root.AddItem(statusBar, 1, 0, false)

	layout := &Layout{
		root:        root,
		table:       table,
		players:     players,
		rules:       rules,
		pingChart:   pingChart,
		status:      status,
		keys:        keys,
		filterPanel: filterPanel,
		statusBar:   statusBar,
	}
	layout.initTable()

	// Set selection change handler
	table.SetSelectionChangedFunc(func(row, col int) {
		if layout.onSelect != nil && row > 0 {
			layout.onSelect(row)
		}
	})

	return layout
}

func (l *Layout) Root() tview.Primitive {
	return l.root
}

func (l *Layout) Table() *tview.Table {
	return l.table
}

func (l *Layout) SetSelectionChangedFunc(f func(row int)) {
	l.onSelect = f
}

func (l *Layout) SetStatus(message string) {
	l.status.SetText(message)
}

func (l *Layout) SetKeysText(keysText string) {
	l.keys.SetText(keysText)
}

func (l *Layout) SetTableTitle(title string) {
	l.table.SetTitle(title)
}

func (l *Layout) SetPlayers(players []string, playerCount int) {
	// Clear existing rows
	l.players.Clear()

	if len(players) == 0 {
		if playerCount > 0 {
			l.players.SetCell(0, 0, tview.NewTableCell("Player list unavailable (SA-MP limitation)").SetSelectable(false))
		} else {
			l.players.SetCell(0, 0, tview.NewTableCell("No players online").SetSelectable(false))
		}
		return
	}

	// Add header
	l.players.SetCell(0, 0, tview.NewTableCell("ID").SetTextColor(tcell.ColorYellow).SetSelectable(false))
	l.players.SetCell(0, 1, tview.NewTableCell("Player Name").SetTextColor(tcell.ColorYellow).SetSelectable(false).SetExpansion(1))

	// Add players
	for i, name := range players {
		row := i + 1
		l.players.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("%d", i)).SetSelectable(false))
		l.players.SetCell(row, 1, tview.NewTableCell(name).SetSelectable(false).SetExpansion(1))
	}
}

func (l *Layout) SetRules(rules map[string]string) {
	// Clear existing rows
	l.rules.Clear()

	if len(rules) == 0 {
		l.rules.SetCell(0, 0, tview.NewTableCell("No rules available").SetSelectable(false))
		return
	}

	// Sort rule names alphabetically
	keys := make([]string, 0, len(rules))
	for key := range rules {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Add header
	l.rules.SetCell(0, 0, tview.NewTableCell("Rule").SetTextColor(tcell.ColorYellow).SetSelectable(false).SetExpansion(1))
	l.rules.SetCell(0, 1, tview.NewTableCell("Value").SetTextColor(tcell.ColorYellow).SetSelectable(false).SetExpansion(2))

	// Add rules as table rows
	for i, key := range keys {
		row := i + 1
		l.rules.SetCell(row, 0, tview.NewTableCell(key).SetSelectable(false).SetExpansion(1))
		l.rules.SetCell(row, 1, tview.NewTableCell(rules[key]).SetSelectable(false).SetExpansion(2))
	}
}

func (l *Layout) SetPingChart(pings []int64) {
	if len(pings) == 0 {
		l.pingChart.SetText("No ping data")
		return
	}

	// Find max ping for scaling
	maxPing := int64(0)
	for _, p := range pings {
		if p > maxPing {
			maxPing = p
		}
	}
	if maxPing == 0 {
		maxPing = 1
	}

	// Create chart (6 lines height)
	height := 6
	width := len(pings)
	if width > 50 {
		pings = pings[len(pings)-50:]
		width = 50
	}

	chart := ""
	// Draw from top to bottom
	for row := height - 1; row >= 0; row-- {
		threshold := maxPing * int64(row+1) / int64(height)
		for _, ping := range pings {
			if ping >= threshold {
				chart += "█"
			} else {
				chart += " "
			}
		}
		if row == height-1 {
			chart += fmt.Sprintf(" %dms", maxPing)
		} else if row == 0 {
			chart += " 0ms"
		}
		chart += "\n"
	}

	// Add labels
	chart += fmt.Sprintf("Latest: %dms | Avg: %dms | Max: %dms",
		pings[len(pings)-1],
		average(pings),
		maxPing)

	l.pingChart.SetText(chart)
}

func average(nums []int64) int64 {
	if len(nums) == 0 {
		return 0
	}
	sum := int64(0)
	for _, n := range nums {
		sum += n
	}
	return sum / int64(len(nums))
}

func (l *Layout) initTable() {
	headers := []string{"Name", "Host", "Ping", "Players"}
	for i, h := range headers {
		cell := tview.NewTableCell(fmt.Sprintf("[::b]%s", h)).
			SetSelectable(false).
			SetExpansion(1)
		l.table.SetCell(0, i, cell)
	}
	l.table.SetFixed(1, 0)
}

func (l *Layout) UpdateTable(servers []server.Server) {
	// Save current selection
	row, col := l.table.GetSelection()

	// Clear all data rows (keep header row 0)
	currentRowCount := l.table.GetRowCount()
	for i := currentRowCount - 1; i > 0; i-- {
		l.table.RemoveRow(i)
	}

	// Handle empty state
	if len(servers) == 0 {
		l.table.SetCell(1, 0, tview.NewTableCell("No servers found").SetSelectable(false))
		l.table.Select(1, 0)
		return
	}

	// Add all server rows
	for i, srv := range servers {
		tableRow := i + 1
		ping := "-"
		players := "-"
		name := srv.Name
		if name == "" {
			name = "(unknown)"
		}
		if !srv.Loading {
			ping = fmt.Sprintf("%d ms", srv.Ping.Milliseconds())
			players = fmt.Sprintf("%d/%d", srv.Players, srv.MaxPlayers)
		} else if srv.LastUpdated.IsZero() {
			ping = "..."
			players = "..."
		}
		if srv.Passworded {
			name = fmt.Sprintf("%s [locked]", name)
		}
		l.table.SetCell(tableRow, 0, tview.NewTableCell(name).SetExpansion(2))
		l.table.SetCell(tableRow, 1, tview.NewTableCell(srv.Addr()).SetExpansion(1))
		l.table.SetCell(tableRow, 2, tview.NewTableCell(ping).SetExpansion(1))
		l.table.SetCell(tableRow, 3, tview.NewTableCell(players).SetExpansion(1))
	}

	// Restore selection if still valid
	newRowCount := len(servers) + 1
	if row < 1 {
		row = 1
	}
	if row >= newRowCount {
		row = newRowCount - 1
	}
	if col < 0 {
		col = 0
	}
	l.table.Select(row, col)
}

func (l *Layout) UpdateTableRow(index int, srv server.Server) {
	tableRow := index + 1
	ping := "-"
	players := "-"
	name := srv.Name
	if name == "" {
		name = "(unknown)"
	}
	if !srv.Loading {
		ping = fmt.Sprintf("%d ms", srv.Ping.Milliseconds())
		players = fmt.Sprintf("%d/%d", srv.Players, srv.MaxPlayers)
	} else if srv.LastUpdated.IsZero() {
		ping = "..."
		players = "..."
	}
	if srv.Passworded {
		name = fmt.Sprintf("%s [locked]", name)
	}
	l.table.SetCell(tableRow, 0, tview.NewTableCell(name).SetExpansion(2))
	l.table.SetCell(tableRow, 1, tview.NewTableCell(srv.Addr()).SetExpansion(1))
	l.table.SetCell(tableRow, 2, tview.NewTableCell(ping).SetExpansion(1))
	l.table.SetCell(tableRow, 3, tview.NewTableCell(players).SetExpansion(1))
}

func (l *Layout) UpdateFilterPanel(text string) {
	l.filterPanel.SetText(text)
}
