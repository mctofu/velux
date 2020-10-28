package cli

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/mctofu/homekit/client"
	"github.com/mctofu/velux"
	"github.com/mctofu/velux/cmd/velux/cli/config"
	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use: "velux",
}

func init() {
	rootCommand.AddCommand(importCmd())
	rootCommand.AddCommand(statusCmd())
	rootCommand.AddCommand(setupCmd())
	rootCommand.AddCommand(setCmd())
}

// Execute the command line interface
func Execute() error {
	return rootCommand.ExecuteContext(context.Background())
}

func markFlagRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

// runner is cobra.Command.RunE
type runner func(cmd *cobra.Command, args []string) error

type configCommand func(ctx context.Context, configPath, controllerName string) error

func configCommandRunner(cmd *cobra.Command, cfgCmd configCommand) runner {
	var defaultConfigPath string
	configDir, configDirErr := os.UserConfigDir()
	if configDirErr == nil {
		defaultConfigPath = path.Join(configDir, "mctofu", "velux")
	}

	configPath := cmd.Flags().String("configPath", defaultConfigPath, "Directory to store controller configs.")
	controllerName := cmd.Flags().String("controller", "default", "Name of controller profile to use.")

	return func(cmd *cobra.Command, args []string) error {
		if *configPath == "" && configDirErr != nil {
			return fmt.Errorf("resolve config dir: %v", configDirErr)
		}

		return cfgCmd(cmd.Context(), *configPath, *controllerName)
	}
}

type clientContext struct {
	AccessoryName string
	Config        *config.ControllerConfig
	ConfigPath    string
}

type clientCommand func(ctx context.Context, clientCtx *clientContext, client *velux.Client) error

func clientCommandRunner(cmd *cobra.Command, clientCmd clientCommand) runner {
	name := cmd.Flags().StringP("name", "n", "", "Name of accessory to act on")

	cfgCmd := func(ctx context.Context, configPath, controllerName string) (rErr error) {
		cfg, err := config.ReadControllerConfig(configPath, controllerName)
		if err != nil {
			return fmt.Errorf("read controller config: %v", err)
		}

		client, err := createClient(cfg, *name)
		if err != nil {
			return err
		}
		defer func() {
			if cErr := client.Close(); cErr != nil {
				rErr = multierror.Append(rErr, cErr)
			}
		}()

		clientCtx := clientContext{
			AccessoryName: *name,
			Config:        cfg,
			ConfigPath:    configPath,
		}

		return clientCmd(cmd.Context(), &clientCtx, client)
	}

	return configCommandRunner(cmd, cfgCmd)
}

type selectCommand func(ctx context.Context, clientCtx *clientContext, client *velux.Client, selection *velux.WindowSelection) error

func selectCommandRunner(cmd *cobra.Command, selectCmd selectCommand) runner {
	codeParams := cmd.Flags().StringSliceP("code", "c", nil, "Select windows with matching codes")
	serialParams := cmd.Flags().StringSliceP("serial", "s", nil, "Select windows with matching serials")
	typeParams := cmd.Flags().StringSliceP("type", "t", nil, "Select windows with matching type [(w)indow or (b)lind")

	clientCmd := func(ctx context.Context, clientCtx *clientContext, client *velux.Client) error {
		windowTypes, err := parseWindowTypes(*typeParams)
		if err != nil {
			return err
		}

		selection := velux.WindowSelection{
			Codes:   *codeParams,
			Serials: *serialParams,
			Types:   windowTypes,
		}

		return selectCmd(ctx, clientCtx, client, &selection)
	}

	return clientCommandRunner(cmd, clientCmd)
}

func parseWindowTypes(windowTypes []string) ([]velux.WindowType, error) {
	if len(windowTypes) == 0 {
		return nil, nil
	}

	var resolvedTypes []velux.WindowType

	for _, windowType := range windowTypes {
		var resolvedType velux.WindowType
		prefix := strings.ToLower(windowType[:1])
		switch prefix {
		case "w":
			resolvedType = velux.WindowTypeWindow
		case "b":
			resolvedType = velux.WindowTypeBlind
		default:
			return nil, fmt.Errorf("unknown type: %s", windowType)
		}
		resolvedTypes = append(resolvedTypes, resolvedType)
	}

	return resolvedTypes, nil
}

func createClient(cfg *config.ControllerConfig, name string) (*velux.Client, error) {
	var accPairing *config.AccessoryPairing

	if len(cfg.AccessoryPairings) == 0 {
		return nil, fmt.Errorf("no paired accessories")
	}

	for _, pair := range cfg.AccessoryPairings {
		if name == "" || pair.Name == name {
			accPairing = pair
			break
		}
	}

	if accPairing == nil {
		return nil, fmt.Errorf("accessory %s not found", name)
	}

	accClient := client.NewAccessoryClient(
		client.NewIPDialer(),
		&client.ControllerIdentity{
			DeviceID:   cfg.DeviceID,
			PrivateKey: cfg.PrivateKey,
			PublicKey:  cfg.PublicKey,
		},
		&client.AccessoryConnectionConfig{
			DeviceID:         accPairing.DeviceID,
			PublicKey:        accPairing.PublicKey,
			IPConnectionInfo: accPairing.IPConnectionInfo,
		},
	)

	return &velux.Client{
		HomekitClient:  accClient,
		WindowMappings: cfg.WindowMappings,
	}, nil
}
