package cli

import (
	"context"
	"fmt"

	"github.com/mctofu/velux"
	"github.com/spf13/cobra"
)

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"s"},
		Short:   "Read temp, humidity and co2 sensors and position of windows",
	}

	cmd.RunE = selectCommandRunner(cmd, readStatus)

	return cmd
}

func readStatus(ctx context.Context, clientCtx *clientContext, client *velux.Client, selection *velux.WindowSelection) error {
	status, err := client.ReadStatus(ctx, *selection)
	if err != nil {
		return fmt.Errorf("ReadSensors: %v", err)
	}

	fmt.Printf("Temperature: %.1fF\n", status.SensorReading.TemperatureF)
	fmt.Printf("Relative Humidity: %.1f\n", status.SensorReading.RelativeHumidity)
	fmt.Printf("CO2 PPM: %.0f\n", status.SensorReading.CarbonDioxideLevel)

	for _, window := range status.Windows {
		fmt.Printf("%s (%s): %d (%d)\n", window.FriendlyName(), window.Code(), window.CurrentPosition, window.TargetPosition)
	}

	return nil
}
