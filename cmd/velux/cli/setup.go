package cli

import (
	"context"
	"fmt"

	"github.com/mctofu/velux"
	"github.com/mctofu/velux/cmd/velux/cli/config"
	"github.com/spf13/cobra"
)

func setupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Set a friendly name and code for a window",
	}

	serial := cmd.Flags().String("serial", "", "serial number of window to setup")
	markFlagRequired(cmd, "serial")
	name := cmd.Flags().String("desc", "", "friendly name to assign window")
	markFlagRequired(cmd, "desc")
	code := cmd.Flags().String("code", "", "short code to assign window")
	markFlagRequired(cmd, "code")

	cmd.RunE = clientCommandRunner(cmd,
		func(ctx context.Context, clientCtx *clientContext, client *velux.Client) error {
			return setupWindow(ctx, clientCtx, client, *serial, *name, *code)
		},
	)

	return cmd
}

func setupWindow(ctx context.Context, clientCtx *clientContext, client *velux.Client, serial, name, code string) error {
	status, err := client.ReadStatus(ctx, velux.WindowSelection{})
	if err != nil {
		return fmt.Errorf("ReadSensors: %v", err)
	}

	var setupWindow *velux.WindowInfo

	for _, window := range status.Windows {
		if window.SerialNumber == serial {
			w := window
			setupWindow = &w
			break
		}
	}

	if setupWindow == nil {
		return fmt.Errorf("no window with serial %s", serial)
	}

	var setupMapping *velux.WindowMapping

	for _, mapping := range clientCtx.Config.WindowMappings {
		if mapping.SerialNumber == serial {
			setupMapping = mapping
			break
		}
	}

	if setupMapping == nil {
		setupMapping = &velux.WindowMapping{
			SerialNumber: serial,
		}
		clientCtx.Config.WindowMappings = append(clientCtx.Config.WindowMappings, setupMapping)
	}

	setupMapping.Name = name
	setupMapping.Code = code

	if err := config.SaveControllerConfig(clientCtx.ConfigPath, clientCtx.Config, true); err != nil {
		return fmt.Errorf("SaveControllerConfig: %v", err)
	}

	return nil
}
