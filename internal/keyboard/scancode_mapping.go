package keyboard

// ScancodeInfo represents information about a scancode
type ScancodeInfo struct {
	KeyName     string
	KeyCategory string
	Description string
}

// ScancodeMappingData provides in-memory mapping of scancodes to key information
// Based on standard PC keyboard scan codes (Set 1)
var ScancodeMappingData = map[int]ScancodeInfo{
	// Function keys
	1:  {"Esc", "special", "Escape key"},
	59: {"F1", "function", "Function key F1"},
	60: {"F2", "function", "Function key F2"},
	61: {"F3", "function", "Function key F3"},
	62: {"F4", "function", "Function key F4"},
	63: {"F5", "function", "Function key F5"},
	64: {"F6", "function", "Function key F6"},
	65: {"F7", "function", "Function key F7"},
	66: {"F8", "function", "Function key F8"},
	67: {"F9", "function", "Function key F9"},
	68: {"F10", "function", "Function key F10"},
	87: {"F11", "function", "Function key F11"},
	88: {"F12", "function", "Function key F12"},

	// Number row
	41: {"`", "special", "Grave accent / Tilde"},
	2:  {"1", "number", "Number 1"},
	3:  {"2", "number", "Number 2"},
	4:  {"3", "number", "Number 3"},
	5:  {"4", "number", "Number 4"},
	6:  {"5", "number", "Number 5"},
	7:  {"6", "number", "Number 6"},
	8:  {"7", "number", "Number 7"},
	9:  {"8", "number", "Number 8"},
	10: {"9", "number", "Number 9"},
	11: {"0", "number", "Number 0"},
	12: {"-", "special", "Minus / Underscore"},
	13: {"=", "special", "Equals / Plus"},
	14: {"Backspace", "special", "Backspace key"},

	// Top letter row (QWERTY)
	15: {"Tab", "special", "Tab key"},
	16: {"Q", "letter", "Letter Q"},
	17: {"W", "letter", "Letter W"},
	18: {"E", "letter", "Letter E"},
	19: {"R", "letter", "Letter R"},
	20: {"T", "letter", "Letter T"},
	21: {"Y", "letter", "Letter Y"},
	22: {"U", "letter", "Letter U"},
	23: {"I", "letter", "Letter I"},
	24: {"O", "letter", "Letter O"},
	25: {"P", "letter", "Letter P"},
	26: {"[", "special", "Left bracket"},
	27: {"]", "special", "Right bracket"},
	43: {"\\", "special", "Backslash / Pipe"},

	// Middle letter row (ASDF)
	58: {"Caps Lock", "special", "Caps Lock key"},
	30: {"A", "letter", "Letter A"},
	31: {"S", "letter", "Letter S"},
	32: {"D", "letter", "Letter D"},
	33: {"F", "letter", "Letter F"},
	34: {"G", "letter", "Letter G"},
	35: {"H", "letter", "Letter H"},
	36: {"J", "letter", "Letter J"},
	37: {"K", "letter", "Letter K"},
	38: {"L", "letter", "Letter L"},
	39: {";", "special", "Semicolon / Colon"},
	40: {"'", "special", "Apostrophe / Quote"},
	28: {"Enter", "special", "Enter key"},

	// Bottom letter row (ZXCV)
	42: {"Left Shift", "modifier", "Left Shift key"},
	44: {"Z", "letter", "Letter Z"},
	45: {"X", "letter", "Letter X"},
	46: {"C", "letter", "Letter C"},
	47: {"V", "letter", "Letter V"},
	48: {"B", "letter", "Letter B"},
	49: {"N", "letter", "Letter N"},
	50: {"M", "letter", "Letter M"},
	51: {",", "special", "Comma / Less than"},
	52: {".", "special", "Period / Greater than"},
	53: {"/", "special", "Slash / Question mark"},
	54: {"Right Shift", "modifier", "Right Shift key"},

	// Bottom row (modifiers and space)
	29:  {"Left Ctrl", "modifier", "Left Control key"},
	91:  {"Left Win", "modifier", "Left Windows key"},
	56:  {"Left Alt", "modifier", "Left Alt key"},
	57:  {"Space", "special", "Space bar"},
	100: {"Right Alt", "modifier", "Right Alt key"},
	92:  {"Right Win", "modifier", "Right Windows key"},
	93:  {"Menu", "special", "Context menu key"},
	97:  {"Right Ctrl", "modifier", "Right Control key"},

	// Navigation cluster (extended scancodes, typically 0xE0 prefix)
	210: {"Insert", "special", "Insert key"},
	211: {"Delete", "special", "Delete key"},
	199: {"Home", "special", "Home key"},
	207: {"End", "special", "End key"},
	201: {"Page Up", "special", "Page Up key"},
	209: {"Page Down", "special", "Page Down key"},
	200: {"Up", "special", "Up arrow"},
	208: {"Down", "special", "Down arrow"},
	203: {"Left", "special", "Left arrow"},
	205: {"Right", "special", "Right arrow"},

	// Numpad
	69: {"Num Lock", "special", "Num Lock key"},
	98: {"/", "special", "Numpad divide"},
	55: {"*", "special", "Numpad multiply"},
	74: {"-", "special", "Numpad minus"},
	78: {"+", "special", "Numpad plus"},
	96: {"Enter", "special", "Numpad Enter"},
	79: {"1", "number", "Numpad 1"},
	80: {"2", "number", "Numpad 2"},
	81: {"3", "number", "Numpad 3"},
	75: {"4", "number", "Numpad 4"},
	76: {"5", "number", "Numpad 5"},
	77: {"6", "number", "Numpad 6"},
	71: {"7", "number", "Numpad 7"},
	72: {"8", "number", "Numpad 8"},
	73: {"9", "number", "Numpad 9"},
	82: {"0", "number", "Numpad 0"},
	83: {".", "special", "Numpad decimal"},

	// Special keys
	70: {"Scroll Lock", "special", "Scroll Lock key"},
	84: {"Print Screen", "special", "Print Screen key"},
	85: {"Pause", "special", "Pause / Break key"},
}

// GetScancodeInfo returns information for a given scancode
func GetScancodeInfo(scancode int) (ScancodeInfo, bool) {
	info, exists := ScancodeMappingData[scancode]
	return info, exists
}

// GetKeyName returns the key name for a scancode, or "Unknown" if not found
func GetKeyName(scancode int) string {
	if info, exists := ScancodeMappingData[scancode]; exists {
		return info.KeyName
	}
	return "Unknown"
}

// GetKeyCategory returns the key category for a scancode, or "unknown" if not found
func GetKeyCategory(scancode int) string {
	if info, exists := ScancodeMappingData[scancode]; exists {
		return info.KeyCategory
	}
	return "unknown"
}
