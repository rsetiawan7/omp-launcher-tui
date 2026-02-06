package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rsetiawan7/omp-launcher-tui/internal/config"
)

// ExportData represents all exportable configuration data
type ExportData struct {
	Version     string             `json:"version"`
	ExportedAt  string             `json:"exported_at"`
	Config      config.Config      `json:"config"`
	Favorites   config.Favorites   `json:"favorites"`
	MasterLists config.MasterLists `json:"master_lists"`
}

// Export exports configuration, favorites, and master lists to a single file
func Export(outputPath string) error {
	// Load all data
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	favorites, err := config.LoadFavorites()
	if err != nil {
		return fmt.Errorf("failed to load favorites: %w", err)
	}

	masterLists, err := config.LoadMasterLists()
	if err != nil {
		return fmt.Errorf("failed to load master lists: %w", err)
	}

	// Create export data structure
	exportData := ExportData{
		Version:     "1.0",
		ExportedAt:  time.Now().Format(time.RFC3339),
		Config:      cfg,
		Favorites:   favorites,
		MasterLists: masterLists,
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal export data: %w", err)
	}

	// Ensure the output directory exists
	outputDir := filepath.Dir(outputPath)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	// Get absolute path for output
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath
	}

	fmt.Println("✓ Export completed successfully!")
	fmt.Printf("✓ Config, favorites, and master lists exported to: %s\n", absPath)
	fmt.Printf("✓ Total favorites: %d\n", len(favorites.Servers))
	fmt.Printf("✓ Total master lists: %d\n", len(masterLists.Lists))

	return nil
}

// Import imports configuration, favorites, and master lists from a file
func Import(inputPath string) error {
	// Check if file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("import file not found: %s", inputPath)
	}

	// Read the file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	// Parse the export data
	var exportData ExportData
	if err := json.Unmarshal(data, &exportData); err != nil {
		return fmt.Errorf("failed to parse import file: %w", err)
	}

	// Validate version (for future compatibility checks)
	if exportData.Version == "" {
		return fmt.Errorf("invalid export file: missing version")
	}

	// Prompt for confirmation
	fmt.Println("Import will overwrite your current configuration.")
	fmt.Printf("Importing data exported at: %s\n", exportData.ExportedAt)
	fmt.Printf("- Config settings\n")
	fmt.Printf("- %d favorite server(s)\n", len(exportData.Favorites.Servers))
	fmt.Printf("- %d master list(s)\n", len(exportData.MasterLists.Lists))
	fmt.Print("\nDo you want to continue? (y/N): ")

	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Import cancelled.")
		return nil
	}

	// Import config
	if err := config.Save(exportData.Config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Import favorites
	if err := config.SaveFavorites(exportData.Favorites); err != nil {
		return fmt.Errorf("failed to save favorites: %w", err)
	}

	// Import master lists
	if err := config.SaveMasterLists(exportData.MasterLists); err != nil {
		return fmt.Errorf("failed to save master lists: %w", err)
	}

	// Get absolute path for input
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		absPath = inputPath
	}

	fmt.Println("✓ Import completed successfully!")
	fmt.Printf("✓ Data imported from: %s\n", absPath)
	fmt.Printf("✓ Imported %d favorite(s)\n", len(exportData.Favorites.Servers))
	fmt.Printf("✓ Imported %d master list(s)\n", len(exportData.MasterLists.Lists))

	return nil
}
