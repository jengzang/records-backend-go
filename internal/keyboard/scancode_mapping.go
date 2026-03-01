package keyboard

// ScancodeInfo represents information about a scancode
type ScancodeInfo struct {
	KeyName     string
	KeyCategory string
	Description string
	Hand        string // "left", "right", "both", or "neutral"
}

// ScancodeMappingData provides in-memory mapping of scancodes to key information
// Based on standard PC keyboard scan codes (Set 1)
// Hand assignment: left (Q/W/E/A/S/D/Z/X/C/1-5), right (U/I/O/J/K/L/N/M/6-0), both (space/enter), neutral (function keys)
var ScancodeMappingData = map[int]ScancodeInfo{
	// Function keys (neutral - typically pressed with either hand)
	1:  {"Esc", "special", "Escape key", "left"},
	59: {"F1", "function", "Function key F1", "left"},
	60: {"F2", "function", "Function key F2", "left"},
	61: {"F3", "function", "Function key F3", "left"},
	62: {"F4", "function", "Function key F4", "left"},
	63: {"F5", "function", "Function key F5", "left"},
	64: {"F6", "function", "Function key F6", "right"},
	65: {"F7", "function", "Function key F7", "right"},
	66: {"F8", "function", "Function key F8", "right"},
	67: {"F9", "function", "Function key F9", "right"},
	68: {"F10", "function", "Function key F10", "right"},
	87: {"F11", "function", "Function key F11", "right"},
	88: {"F12", "function", "Function key F12", "right"},

	// Number row
	41: {"`", "special", "Grave accent / Tilde", "left"},
	2:  {"1", "number", "Number 1", "left"},
	3:  {"2", "number", "Number 2", "left"},
	4:  {"3", "number", "Number 3", "left"},
	5:  {"4", "number", "Number 4", "left"},
	6:  {"5", "number", "Number 5", "left"},
	7:  {"6", "number", "Number 6", "right"},
	8:  {"7", "number", "Number 7", "right"},
	9:  {"8", "number", "Number 8", "right"},
	10: {"9", "number", "Number 9", "right"},
	11: {"0", "number", "Number 0", "right"},
	12: {"-", "special", "Minus / Underscore", "right"},
	13: {"=", "special", "Equals / Plus", "right"},
	14: {"Backspace", "special", "Backspace key", "right"},

	// Top letter row (QWERTY)
	15: {"Tab", "special", "Tab key", "left"},
	16: {"Q", "letter", "Letter Q", "left"},
	17: {"W", "letter", "Letter W", "left"},
	18: {"E", "letter", "Letter E", "left"},
	19: {"R", "letter", "Letter R", "left"},
	20: {"T", "letter", "Letter T", "left"},
	21: {"Y", "letter", "Letter Y", "right"},
	22: {"U", "letter", "Letter U", "right"},
	23: {"I", "letter", "Letter I", "right"},
	24: {"O", "letter", "Letter O", "right"},
	25: {"P", "letter", "Letter P", "right"},
	26: {"[", "special", "Left bracket", "right"},
	27: {"]", "special", "Right bracket", "right"},
	43: {"\\", "special", "Backslash / Pipe", "right"},

	// Middle letter row (ASDF)
	58: {"Caps Lock", "special", "Caps Lock key", "left"},
	30: {"A", "letter", "Letter A", "left"},
	31: {"S", "letter", "Letter S", "left"},
	32: {"D", "letter", "Letter D", "left"},
	33: {"F", "letter", "Letter F", "left"},
	34: {"G", "letter", "Letter G", "left"},
	35: {"H", "letter", "Letter H", "right"},
	36: {"J", "letter", "Letter J", "right"},
	37: {"K", "letter", "Letter K", "right"},
	38: {"L", "letter", "Letter L", "right"},
	39: {";", "special", "Semicolon / Colon", "right"},
	40: {"'", "special", "Apostrophe / Quote", "right"},
	28: {"Enter", "special", "Enter key", "right"},

	// Bottom letter row (ZXCV)
	42: {"Left Shift", "modifier", "Left Shift key", "left"},
	44: {"Z", "letter", "Letter Z", "left"},
	45: {"X", "letter", "Letter X", "left"},
	46: {"C", "letter", "Letter C", "left"},
	47: {"V", "letter", "Letter V", "left"},
	48: {"B", "letter", "Letter B", "left"},
	49: {"N", "letter", "Letter N", "right"},
	50: {"M", "letter", "Letter M", "right"},
	51: {",", "special", "Comma / Less than", "right"},
	52: {".", "special", "Period / Greater than", "right"},
	53: {"/", "special", "Slash / Question mark", "right"},
	54: {"Right Shift", "modifier", "Right Shift key", "right"},

	// Bottom row (modifiers and space)
	29:  {"Left Ctrl", "modifier", "Left Control key", "left"},
	91:  {"Left Win", "modifier", "Left Windows key", "left"},
	56:  {"Left Alt", "modifier", "Left Alt key", "left"},
	57:  {"Space", "special", "Space bar", "both"},
	100: {"Right Alt", "modifier", "Right Alt key", "right"},
	92:  {"Right Win", "modifier", "Right Windows key", "right"},
	93:  {"Menu", "special", "Context menu key", "right"},
	97:  {"Right Ctrl", "modifier", "Right Control key", "right"},

	// Navigation cluster (extended scancodes, typically 0xE0 prefix)
	210: {"Insert", "special", "Insert key", "right"},
	211: {"Delete", "special", "Delete key", "right"},
	199: {"Home", "special", "Home key", "right"},
	207: {"End", "special", "End key", "right"},
	201: {"Page Up", "special", "Page Up key", "right"},
	209: {"Page Down", "special", "Page Down key", "right"},
	200: {"Up", "special", "Up arrow", "right"},
	208: {"Down", "special", "Down arrow", "right"},
	203: {"Left", "special", "Left arrow", "right"},
	205: {"Right", "special", "Right arrow", "right"},

	// Numpad (right hand)
	69: {"Num Lock", "special", "Num Lock key", "right"},
	98: {"/", "special", "Numpad divide", "right"},
	55: {"*", "special", "Numpad multiply", "right"},
	74: {"-", "special", "Numpad minus", "right"},
	78: {"+", "special", "Numpad plus", "right"},
	96: {"Enter", "special", "Numpad Enter", "right"},
	79: {"1", "number", "Numpad 1", "right"},
	80: {"2", "number", "Numpad 2", "right"},
	81: {"3", "number", "Numpad 3", "right"},
	75: {"4", "number", "Numpad 4", "right"},
	76: {"5", "number", "Numpad 5", "right"},
	77: {"6", "number", "Numpad 6", "right"},
	71: {"7", "number", "Numpad 7", "right"},
	72: {"8", "number", "Numpad 8", "right"},
	73: {"9", "number", "Numpad 9", "right"},
	82: {"0", "number", "Numpad 0", "right"},
	83: {".", "special", "Numpad decimal", "right"},

	// Special keys
	70: {"Scroll Lock", "special", "Scroll Lock key", "neutral"},
	84: {"Print Screen", "special", "Print Screen key", "neutral"},
	85: {"Pause", "special", "Pause / Break key", "neutral"},
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

// GetKeyHand returns the hand for a scancode, or "neutral" if not found
func GetKeyHand(scancode int) string {
	if info, exists := ScancodeMappingData[scancode]; exists {
		return info.Hand
	}
	return "neutral"
}
