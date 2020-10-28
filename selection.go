package velux

// WindowSelection matches windows with matching attributes
type WindowSelection struct {
	Codes   []string
	Serials []string
	Types   []WindowType
}

// Matches checks if the provided window matches the selection's attributes.
func (w *WindowSelection) Matches(info *WindowInfo) bool {
	if !w.codeMatch(info) {
		return false
	}
	if !w.serialMatch(info) {
		return false
	}
	if !w.typeMatch(info) {
		return false
	}

	return true
}

func (w *WindowSelection) codeMatch(info *WindowInfo) bool {
	if len(w.Codes) == 0 {
		return true
	}

	if info.Mapping == nil {
		return false
	}

	for _, code := range w.Codes {
		if code == info.Mapping.Code {
			return true
		}
	}

	return false
}

func (w *WindowSelection) serialMatch(info *WindowInfo) bool {
	if len(w.Serials) == 0 {
		return true
	}

	for _, serial := range w.Serials {
		if serial == info.SerialNumber {
			return true
		}
	}

	return false
}

func (w *WindowSelection) typeMatch(info *WindowInfo) bool {
	if len(w.Types) == 0 {
		return true
	}

	for _, wType := range w.Types {
		if wType == info.WindowType {
			return true
		}
	}

	return false
}
