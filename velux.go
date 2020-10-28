package velux

// WindowType identifies a window device as a window or blind
type WindowType int

// WindowType enums
const (
	WindowTypeWindow WindowType = iota
	WindowTypeBlind
)

// WindowInfo capture identifying information and current position of a window.
type WindowInfo struct {
	AccessoryID      uint64
	SerialNumber     string
	Type             string
	WindowType       WindowType
	CurrentPosition  byte
	TargetPosition   byte
	TargetPositionID uint64
	Mapping          *WindowMapping
}

// FriendlyName returns the mapped name if found. Serial number otherwise.
func (w *WindowInfo) FriendlyName() string {
	if w.Mapping == nil || w.Mapping.Name == "" {
		return w.SerialNumber
	}

	return w.Mapping.Name
}

// Code returns the mapped short code if found.
func (w *WindowInfo) Code() string {
	if w.Mapping == nil || w.Mapping.Code == "" {
		return w.Type
	}

	return w.Mapping.Code
}

// WindowMapping maps a window's SerialNumber to a friendly name and code
type WindowMapping struct {
	SerialNumber string
	Name         string
	Code         string
}
