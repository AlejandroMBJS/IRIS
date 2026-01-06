/*
Package payroll - Payroll Configuration Loader

==============================================================================
FILE: internal/config/payroll/loader.go
==============================================================================

DESCRIPTION:
    Loads payroll configuration from JSON files in the config directory.
    Uses a master configuration file (main.json) that references other
    config files for modular configuration management.

USER PERSPECTIVE:
    - Loads all payroll rules from config/payroll/ directory
    - Master file controls which config files are loaded
    - Validates configuration after loading

DEVELOPER GUIDELINES:
    ✅  OK to modify: Add new config file loaders
    ⚠️  CAUTION: Changing file paths or structure
    ❌  DO NOT modify: Master file format without updating docs
    📝  Add new config types to loadAllConfigurations()

SYNTAX EXPLANATION:
    - PayrollConfigLoader: Manages config loading process
    - MasterConfig: Controls which files to load
    - Load(): Main entry point, returns complete config
    - loadConfigFile(): Generic loader for any config section

CONFIG DIRECTORY STRUCTURE:
    config/
    └── payroll/
        ├── main.json           (master control file)
        ├── official_values.json
        ├── regional.json
        ├── contribution_rates.json
        ├── labor_concepts.json
        └── calculation_tables.json

==============================================================================
*/
package payroll

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"


)

// PayrollConfigLoader handles loading payroll configuration
type PayrollConfigLoader struct {
    configDir string
    config    *PayrollConfig // Corrected to PayrollConfig
    master    *MasterConfig
}

// MasterConfig represents the main control JSON file
type MasterConfig struct {
    Version     string                 `json:"version"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    LastUpdated string                 `json:"last_updated"`
    ConfigFiles map[string]string      `json:"config_files"`
    Tables      map[string]interface{} `json:"tables"`
    Holidays    map[string]string      `json:"holidays"`
    Settings    map[string]interface{} `json:"settings"`
}

// NewPayrollConfigLoader creates a new configuration loader
func NewPayrollConfigLoader(configDir string) *PayrollConfigLoader {
    return &PayrollConfigLoader{
        configDir: configDir,
        config:    &PayrollConfig{},
    }
}

// Load loads all payroll configuration from JSON files
func (pcl *PayrollConfigLoader) Load() (*PayrollConfig, error) {
    // Load master configuration file
    masterPath := filepath.Join(pcl.configDir, "payroll", "main.json")
    if err := pcl.loadMasterConfig(masterPath); err != nil {
        return nil, fmt.Errorf("error loading master config: %w", err)
    }
    
    // Load all configuration files referenced in master
    if err := pcl.loadAllConfigurations(); err != nil {
        return nil, err
    }
    
    // Validate the complete configuration
    validator := NewValidator()
    if err := validator.Validate(pcl.config); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return pcl.config, nil
}

// loadMasterConfig loads the master control file
func (pcl *PayrollConfigLoader) loadMasterConfig(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("error reading master config file %s: %w", path, err)
    }
    
    var master MasterConfig
    if err := json.Unmarshal(data, &master); err != nil {
        return fmt.Errorf("error parsing master config JSON: %w", err)
    }
    
    pcl.master = &master
    return nil
}

// loadAllConfigurations loads all configuration files
func (pcl *PayrollConfigLoader) loadAllConfigurations() error {
    // Load official values
    if err := pcl.loadConfigFile("official_values", &pcl.config.OfficialValues); err != nil {
        return err
    }
    
    // Load regional configuration
    if err := pcl.loadConfigFile("regional", &pcl.config.Regional); err != nil {
        return err
    }
    
    // Load contribution rates
    if err := pcl.loadConfigFile("contribution_rates", &pcl.config.ContributionRates); err != nil {
        return err
    }
    
    // Load labor concepts
    if err := pcl.loadConfigFile("labor_concepts", &pcl.config.LaborConcepts); err != nil {
        return err
    }
    
    // Load calculation tables configuration
    if err := pcl.loadConfigFile("calculation_tables", &pcl.config.CalculationTables); err != nil {
        return err
    }
    
    return nil
}

// loadConfigFile loads a specific configuration file
func (pcl *PayrollConfigLoader) loadConfigFile(configType string, target interface{}) error {
    if pcl.master == nil || pcl.master.ConfigFiles == nil {
        return fmt.Errorf("master config not loaded")
    }
    
    filePath, exists := pcl.master.ConfigFiles[configType]
    if !exists {
        return fmt.Errorf("configuration file for %s not specified in master config", configType)
    }
    
    // Resolve relative path
    if !filepath.IsAbs(filePath) {
        filePath = filepath.Join(pcl.configDir, filePath)
    }
    
    data, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("error reading config file %s: %w", filePath, err)
    }
    
    if err := json.Unmarshal(data, target); err != nil {
        return fmt.Errorf("error parsing config file %s: %w", filePath, err)
    }
    
    return nil
}

// GetMasterConfig returns the loaded master configuration
func (pcl *PayrollConfigLoader) GetMasterConfig() *MasterConfig {
    return pcl.master
}
