package screentime

import (
	"database/sql"
	"fmt"
	"strings"
)

// AppCategoryMapper handles application category classification
type AppCategoryMapper struct {
	categoryRules map[string][]string
}

// NewAppCategoryMapper creates a new category mapper with predefined rules
func NewAppCategoryMapper() *AppCategoryMapper {
	return &AppCategoryMapper{
		categoryRules: map[string][]string{
			"Productivity": {
				"Microsoft Office", "Excel", "Word", "PowerPoint", "Outlook",
				"WPS", "Notion", "Evernote", "印象笔记", "有道云笔记",
				"Obsidian", "Typora", "Markdown", "OneNote",
			},
			"Development": {
				"Visual Studio", "Code", "VSCode", "PyCharm", "IntelliJ",
				"Eclipse", "Android Studio", "Xcode", "Sublime",
				"Git", "GitHub", "Docker", "Terminal", "PowerShell",
				"Postman", "Insomnia", "DBeaver", "Navicat",
			},
			"Communication": {
				"微信", "WeChat", "QQ", "TIM", "钉钉", "DingTalk",
				"Slack", "Teams", "Zoom", "Skype", "飞书", "Feishu",
				"Telegram", "Discord", "WhatsApp", "Line",
			},
			"Entertainment": {
				"抖音", "TikTok", "哔哩哔哩", "Bilibili", "B站",
				"YouTube", "Netflix", "爱奇艺", "腾讯视频", "优酷",
				"Spotify", "网易云音乐", "QQ音乐", "Apple Music",
				"Steam", "Epic", "Origin", "Uplay",
			},
			"Social": {
				"微博", "Weibo", "小红书", "Instagram", "Ins",
				"Facebook", "Twitter", "LinkedIn", "知乎", "Zhihu",
				"豆瓣", "Douban", "贴吧", "Tieba",
			},
			"Shopping": {
				"淘宝", "Taobao", "京东", "JD", "拼多多", "Pinduoduo",
				"Amazon", "支付宝", "Alipay", "美团", "Meituan",
				"饿了么", "Eleme", "闲鱼", "Xianyu",
			},
			"Browser": {
				"Chrome", "Edge", "Firefox", "Safari", "Opera",
				"Brave", "Vivaldi", "Arc", "浏览器",
			},
			"Tools": {
				"文件管理", "File Explorer", "Finder", "Total Commander",
				"Calculator", "计算器", "日历", "Calendar", "时钟",
				"记事本", "Notepad", "截图", "Screenshot", "录屏",
			},
			"Office": {
				"Excel", "Word", "PowerPoint", "WPS", "Office",
				"PDF", "Adobe", "Acrobat", "Reader",
			},
			"Gaming": {
				"王者荣耀", "和平精英", "原神", "Genshin", "英雄联盟",
				"LOL", "DOTA", "CS", "Minecraft", "我的世界",
			},
			"System": {
				"系统桌面", "Launcher", "设置", "Settings", "System",
				"任务管理", "Task Manager", "控制面板", "Control Panel",
			},
		},
	}
}

// GetCategory returns the category for a given app name
func (m *AppCategoryMapper) GetCategory(appName string) string {
	normalized := strings.ToLower(appName)

	for category, keywords := range m.categoryRules {
		for _, keyword := range keywords {
			if strings.Contains(normalized, strings.ToLower(keyword)) {
				return category
			}
		}
	}

	return "Other"
}

// UpdateAppCategories updates categories for all apps in a table
// appNameColumn specifies the column name for app names (e.g., "app_name" or "application")
func (m *AppCategoryMapper) UpdateAppCategories(db *sql.DB, tableName, appNameColumn string) error {
	// Query all apps
	query := fmt.Sprintf("SELECT id, %s FROM %s", appNameColumn, tableName)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query apps: %w", err)
	}
	defer rows.Close()

	// Prepare update statement
	updateQuery := fmt.Sprintf("UPDATE %s SET category = ? WHERE id = ?", tableName)
	updateStmt, err := db.Prepare(updateQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer updateStmt.Close()

	// Update each app
	updateCount := 0
	for rows.Next() {
		var id int
		var appName string
		if err := rows.Scan(&id, &appName); err != nil {
			continue
		}

		category := m.GetCategory(appName)
		if _, err := updateStmt.Exec(category, id); err != nil {
			return fmt.Errorf("failed to update app %d: %w", id, err)
		}
		updateCount++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	fmt.Printf("Updated %d apps in %s\n", updateCount, tableName)
	return nil
}
