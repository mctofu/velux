package cli

import (
	"context"
	"fmt"

	"github.com/mctofu/velux"
	"github.com/spf13/cobra"
)

func setCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set target position of particular windows and/or blinds",
	}

	positionParam := cmd.Flags().Uint8P("position", "p", 0, "Position to set window to")

	markFlagRequired(cmd, "position")

	cmd.RunE = selectCommandRunner(cmd,
		func(ctx context.Context, clientCtx *clientContext, client *velux.Client, selection *velux.WindowSelection) error {
			return setPosition(ctx, clientCtx, client, selection, *positionParam)
		},
	)

	return cmd
}

func setPosition(
	ctx context.Context,
	clientCtx *clientContext,
	client *velux.Client,
	selection *velux.WindowSelection,
	position byte,
) error {
	if position > 100 {
		return fmt.Errorf("out of range position: %d", position)
	}

	status, err := client.SetPosition(ctx, *selection, position)
	if err != nil {
		return fmt.Errorf("SetPosition: %v", err)
	}

	if status.Total() == 0 {
		fmt.Println("No windows selected")
	} else if len(status.Modified) == 0 {
		fmt.Println("Selected windows already in position")
	} else {
		fmt.Println("Updated windows:")
		for _, window := range status.Modified {
			fmt.Printf("%s (%s): %d (%d)\n", window.FriendlyName(), window.Code(), window.CurrentPosition, position)
		}
	}

	return nil
}
