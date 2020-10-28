package velux

import (
	"context"
	"fmt"

	"github.com/mctofu/homekit/client"
	"github.com/mctofu/homekit/client/service"
)

// Status captured env readings from the controller as well as window settings.
type Status struct {
	SensorReading SensorReading
	Windows       []WindowInfo
}

// SensorReading captures environment measurments from a Velux Active controller.
type SensorReading struct {
	TemperatureF       float64
	RelativeHumidity   float64
	CarbonDioxideLevel float64
}

// SetStatus indicates the result of a request to change window parameters.
type SetStatus struct {
	Modified   []WindowInfo
	Unmodified []WindowInfo
}

// Total is the number of windows matching the set request.
func (s *SetStatus) Total() int {
	return len(s.Modified) + len(s.Unmodified)
}

// Client allows reading/writing from a Velux Active system using the Homekit integration.
type Client struct {
	HomekitClient  *client.AccessoryClient
	WindowMappings []*WindowMapping
}

// ReadStatus reads curent controller and window information.
func (c *Client) ReadStatus(ctx context.Context, selection WindowSelection) (*Status, error) {
	accessories, err := c.HomekitClient.Accessories(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch accessories: %v", err)
	}

	sensorReading := c.readSensors(accessories)
	windows := c.readWindows(accessories, &selection)

	return &Status{
		SensorReading: *sensorReading,
		Windows:       windows,
	}, nil
}

// SetPosition changes the target position of the selected windows to open or close them.
func (c *Client) SetPosition(ctx context.Context, selection WindowSelection, position byte) (*SetStatus, error) {
	status, err := c.ReadStatus(ctx, selection)
	if err != nil {
		return nil, fmt.Errorf("ReadStatus: %v", err)
	}

	var setStatus SetStatus
	var writes []client.CharacteristicWriteRequest

	for _, window := range status.Windows {
		if window.TargetPosition == position {
			setStatus.Unmodified = append(setStatus.Unmodified, window)
			continue
		}

		writes = append(writes, client.CharacteristicWriteRequest{
			AccessoryID:      window.AccessoryID,
			CharacteristicID: window.TargetPositionID,
			Value:            position,
		})
		setStatus.Modified = append(setStatus.Modified, window)
	}

	if len(writes) > 0 {
		writeReq := &client.CharacteristicsWriteRequest{Characteristics: writes}

		if _, err := c.HomekitClient.SetCharacteristics(ctx, writeReq); err != nil {
			return &setStatus, fmt.Errorf("SetCharacteristics: %v", err)
		}
	}

	return &setStatus, nil
}

func (c *Client) readSensors(accessories []*client.RawAccessory) *SensorReading {
	var readings SensorReading

	for _, acc := range accessories {
		for _, svc := range acc.Services {
			switch svc.Type {
			case service.TypeTemperatureSensor:
				temp := service.ReadTemperatureSensor(svc.Characteristics)
				readings.TemperatureF = temp.CurrentTemperature.Fahrenheit()
			case service.TypeHumiditySensor:
				humidity := service.ReadHumiditySensor(svc.Characteristics)
				readings.RelativeHumidity = humidity.CurrentRelativeHumidity.Value
			case service.TypeCarbonDioxideSensor:
				co2Level := service.ReadCarbonDioxideSensor(svc.Characteristics)
				readings.CarbonDioxideLevel = co2Level.CarbonDioxideLevel.Value
			}
		}
	}

	return &readings
}

func (c *Client) readWindows(accessories []*client.RawAccessory, selection *WindowSelection) []WindowInfo {
	var windows []WindowInfo

	for _, acc := range accessories {
		for _, svc := range acc.Services {
			var info *WindowInfo
			switch svc.Type {
			case service.TypeWindow:
				window := service.ReadWindow(svc.Characteristics)
				info = c.extractWindow(acc, window)
			case service.TypeWindowCovering:
				covering := service.ReadWindowCovering(svc.Characteristics)
				info = c.extractWindowCovering(acc, covering)
			}

			if info != nil && selection.Matches(info) {
				windows = append(windows, *info)
			}
		}
	}

	return windows
}

func (c *Client) extractWindow(a *client.RawAccessory, w *service.Window) *WindowInfo {
	serialNumber := a.Info().SerialNumber.Value
	return &WindowInfo{
		AccessoryID:      a.ID,
		SerialNumber:     serialNumber,
		Type:             w.Name.Value,
		WindowType:       WindowTypeWindow,
		CurrentPosition:  w.CurrentPosition.Value,
		TargetPosition:   w.TargetPosition.Value,
		TargetPositionID: w.TargetPosition.ID,
		Mapping:          c.mapBySerial()[serialNumber],
	}
}

func (c *Client) extractWindowCovering(a *client.RawAccessory, w *service.WindowCovering) *WindowInfo {
	serialNumber := a.Info().SerialNumber.Value
	return &WindowInfo{
		AccessoryID:      a.ID,
		SerialNumber:     serialNumber,
		Type:             w.Name.Value,
		WindowType:       WindowTypeBlind,
		CurrentPosition:  w.CurrentPosition.Value,
		TargetPosition:   w.TargetPosition.Value,
		TargetPositionID: w.TargetPosition.ID,
		Mapping:          c.mapBySerial()[serialNumber],
	}
}

func (c *Client) mapBySerial() map[string]*WindowMapping {
	serialMap := make(map[string]*WindowMapping)

	for _, mapping := range c.WindowMappings {
		serialMap[mapping.SerialNumber] = mapping
	}

	return serialMap
}

func (c *Client) mapByCode() map[string]*WindowMapping {
	codeMap := make(map[string]*WindowMapping)

	for _, mapping := range c.WindowMappings {
		codeMap[mapping.Code] = mapping
	}

	return codeMap
}

// Close releases any resources held by the Client.
func (c *Client) Close() error {
	return c.HomekitClient.Close()
}
