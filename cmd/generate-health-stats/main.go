package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jengzang/records-backend-go/internal/health"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 获取数据库路径
	dbPath := os.Getenv("HEALTH_DB_PATH")
	if dbPath == "" {
		// 默认路径
		dbPath = filepath.Join("data", "applehealth", "health.db")
	}

	// 打开数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to database:", dbPath)

	// 创建统计生成器
	generator := health.NewStatisticsGenerator(db)

	// 获取当前统计摘要
	fmt.Println("\n=== Current Statistics Summary ===")
	summary, err := generator.GetStatisticsSummary()
	if err != nil {
		log.Printf("Warning: Failed to get summary: %v", err)
	} else {
		fmt.Printf("Total statistics: %v\n", summary["total_statistics"])
		fmt.Printf("By type: %v\n", summary["by_type"])
		fmt.Printf("By metric: %v\n", summary["by_metric"])
	}

	// 重新生成统计数据
	fmt.Println("\n=== Regenerating Statistics ===")
	fmt.Println("This may take a few minutes...")

	if err := generator.RegenerateStatistics(); err != nil {
		log.Fatalf("Failed to regenerate statistics: %v", err)
	}

	fmt.Println("✓ Statistics regenerated successfully!")

	// 获取新的统计摘要
	fmt.Println("\n=== New Statistics Summary ===")
	summary, err = generator.GetStatisticsSummary()
	if err != nil {
		log.Printf("Warning: Failed to get summary: %v", err)
	} else {
		fmt.Printf("Total statistics: %v\n", summary["total_statistics"])
		fmt.Printf("By type: %v\n", summary["by_type"])
		fmt.Printf("By metric: %v\n", summary["by_metric"])
		if lastUpdate, ok := summary["last_update"]; ok {
			fmt.Printf("Last update: %v\n", lastUpdate)
		}
	}

	fmt.Println("\n✓ Done!")
}
