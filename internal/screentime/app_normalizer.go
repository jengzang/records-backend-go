package screentime

import (
	"strings"
)

// AppNameNormalizer handles app name normalization and matching
type AppNameNormalizer struct {
	// Mapping of normalized names to their variants
	knownApps map[string][]string
}

// NewAppNameNormalizer creates a new normalizer with predefined mappings
func NewAppNameNormalizer() *AppNameNormalizer {
	return &AppNameNormalizer{
		knownApps: map[string][]string{
			"edge": {"Microsoft Edge", "Edge", "edge"},
			"chrome": {"Google Chrome", "Chrome", "chrome"},
			"firefox": {"Mozilla Firefox", "Firefox", "firefox"},
			"wechat": {"微信", "WeChat", "wechat"},
			"qq": {"QQ", "qq", "腾讯QQ"},
			"telegram": {"Telegram", "telegram"},
			"spotify": {"Spotify", "spotify"},
			"vscode": {"Visual Studio Code", "Code", "vscode", "VSCode"},
			"notion": {"Notion", "notion"},
			"slack": {"Slack", "slack"},
			"discord": {"Discord", "discord"},
			"zoom": {"Zoom", "zoom", "Zoom Meetings"},
			"teams": {"Microsoft Teams", "Teams", "teams"},
			"excel": {"Microsoft Excel", "Excel", "excel", "WPS表格"},
			"word": {"Microsoft Word", "Word", "word", "WPS文字"},
			"powerpoint": {"Microsoft PowerPoint", "PowerPoint", "powerpoint", "WPS演示"},
			"outlook": {"Microsoft Outlook", "Outlook", "outlook"},
			"onenote": {"Microsoft OneNote", "OneNote", "onenote"},
			"evernote": {"Evernote", "evernote", "印象笔记"},
			"dropbox": {"Dropbox", "dropbox"},
			"onedrive": {"OneDrive", "onedrive"},
			"googledrive": {"Google Drive", "googledrive"},
			"photoshop": {"Adobe Photoshop", "Photoshop", "photoshop"},
			"illustrator": {"Adobe Illustrator", "Illustrator", "illustrator"},
			"youtube": {"YouTube", "youtube"},
			"netflix": {"Netflix", "netflix"},
			"bilibili": {"哔哩哔哩", "bilibili", "Bilibili"},
			"douyin": {"抖音", "douyin", "TikTok"},
			"alipay": {"支付宝", "Alipay", "alipay"},
			"taobao": {"淘宝", "Taobao", "taobao"},
			"jd": {"京东", "JD", "jd"},
			"meituan": {"美团", "Meituan", "meituan"},
		},
	}
}

// NormalizeAppName normalizes an app name for comparison
func (n *AppNameNormalizer) NormalizeAppName(name string) string {
	// Remove common prefixes and suffixes
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "Microsoft ")
	name = strings.TrimPrefix(name, "Google ")
	name = strings.TrimPrefix(name, "Adobe ")
	name = strings.TrimSuffix(name, ".exe")
	name = strings.TrimSuffix(name, " (64-bit)")
	name = strings.TrimSuffix(name, " (32-bit)")

	// Convert to lowercase for comparison
	normalized := strings.ToLower(name)

	// Check if this matches any known app
	for canonicalName, variants := range n.knownApps {
		for _, variant := range variants {
			if strings.ToLower(variant) == normalized ||
			   strings.Contains(normalized, strings.ToLower(variant)) ||
			   strings.Contains(strings.ToLower(variant), normalized) {
				return canonicalName
			}
		}
	}

	return normalized
}

// GetCanonicalName returns the display name for a normalized name
func (n *AppNameNormalizer) GetCanonicalName(normalized string) string {
	if variants, ok := n.knownApps[normalized]; ok && len(variants) > 0 {
		return variants[0]
	}
	return normalized
}

// FindCrossPlatformApps identifies apps that exist on both platforms
func (n *AppNameNormalizer) FindCrossPlatformApps(phoneApps, computerApps []string) []string {
	// Normalize phone apps
	phoneNormalized := make(map[string]string) // normalized -> original
	for _, app := range phoneApps {
		normalized := n.NormalizeAppName(app)
		if _, exists := phoneNormalized[normalized]; !exists {
			phoneNormalized[normalized] = app
		}
	}

	// Normalize computer apps
	computerNormalized := make(map[string]string) // normalized -> original
	for _, app := range computerApps {
		normalized := n.NormalizeAppName(app)
		if _, exists := computerNormalized[normalized]; !exists {
			computerNormalized[normalized] = app
		}
	}

	// Find intersection
	crossPlatform := []string{}
	seen := make(map[string]bool)

	for normalized := range phoneNormalized {
		if _, existsInComputer := computerNormalized[normalized]; existsInComputer {
			canonicalName := n.GetCanonicalName(normalized)
			if !seen[canonicalName] {
				crossPlatform = append(crossPlatform, canonicalName)
				seen[canonicalName] = true
			}
		}
	}

	return crossPlatform
}

// FilterPhoneOnlyApps returns apps that only exist on phone
func (n *AppNameNormalizer) FilterPhoneOnlyApps(phoneApps, computerApps []string, limit int) []string {
	// Normalize computer apps for lookup
	computerNormalized := make(map[string]bool)
	for _, app := range computerApps {
		normalized := n.NormalizeAppName(app)
		computerNormalized[normalized] = true
	}

	// Find phone-only apps
	phoneOnly := []string{}
	seen := make(map[string]bool)

	for _, app := range phoneApps {
		normalized := n.NormalizeAppName(app)
		if !computerNormalized[normalized] && !seen[app] {
			phoneOnly = append(phoneOnly, app)
			seen[app] = true
			if len(phoneOnly) >= limit {
				break
			}
		}
	}

	return phoneOnly
}

// FilterComputerOnlyApps returns apps that only exist on computer
func (n *AppNameNormalizer) FilterComputerOnlyApps(phoneApps, computerApps []string, limit int) []string {
	// Normalize phone apps for lookup
	phoneNormalized := make(map[string]bool)
	for _, app := range phoneApps {
		normalized := n.NormalizeAppName(app)
		phoneNormalized[normalized] = true
	}

	// Find computer-only apps
	computerOnly := []string{}
	seen := make(map[string]bool)

	for _, app := range computerApps {
		normalized := n.NormalizeAppName(app)
		if !phoneNormalized[normalized] && !seen[app] {
			computerOnly = append(computerOnly, app)
			seen[app] = true
			if len(computerOnly) >= limit {
				break
			}
		}
	}

	return computerOnly
}
