package cli

import (
	"context"
	"fmt"
	"os"
	"path"

	homekitcfg "github.com/mctofu/homekit/cmd/homekit/cli/config"
	"github.com/mctofu/velux/cmd/velux/cli/config"
	"github.com/spf13/cobra"
)

func importCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import a homekit cli configuration",
	}

	var defaultImportConfigPath string
	configDir, configDirErr := os.UserConfigDir()
	if configDirErr == nil {
		defaultImportConfigPath = path.Join(configDir, "mctofu", "homekit")
	}

	importPathParam := cmd.Flags().String("importPath", defaultImportConfigPath, "homekit config dir")
	importControllerParam := cmd.Flags().String("importController", "default", "homekit controller to import")

	cmd.RunE = configCommandRunner(cmd,
		func(ctx context.Context, configPath, controllerName string) error {
			if *importPathParam == "" {
				if configDirErr != nil {
					return fmt.Errorf("resolve config dir: %v", configDirErr)
				}
				return fmt.Errorf("importPath must be specified")
			}

			if *importControllerParam == "" {
				return fmt.Errorf("importController must be specified")
			}

			return importConfig(ctx, configPath, controllerName, *importPathParam, *importControllerParam)
		},
	)

	return cmd
}

func importConfig(ctx context.Context, configPath, controllerName string, importPath string, importController string) error {
	cfg, err := homekitcfg.ReadControllerConfig(importPath, importController)
	if err != nil {
		return fmt.Errorf("read homekit config: %v", err)
	}

	var importedPairings []*config.AccessoryPairing
	for _, pairing := range cfg.AccessoryPairings {
		if pairing.Model == "VELUX" {
			importPairing := config.AccessoryPairing(*pairing)
			importedPairings = append(importedPairings, &importPairing)
		}
	}

	imported := config.ControllerConfig{
		Name:              controllerName,
		DeviceID:          cfg.DeviceID,
		AccessoryPairings: importedPairings,
		PublicKey:         cfg.PublicKey,
		PrivateKey:        cfg.PrivateKey,
	}
	if err := config.SaveControllerConfig(configPath, &imported, false); err != nil {
		return fmt.Errorf("save velux config: %v", err)
	}

	fmt.Printf("Imported controller and %d pairings\n", len(importedPairings))

	return nil
}
