package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/jengzang/records-backend-go/internal/health"
	_ "modernc.org/sqlite"
)

func main() {
	// 命令行参数
	dbPath := flag.String("db", "./data/applehealth/health.db", "数据库路径")
	action := flag.String("action", "generate", "操作: generate(生成), regenerate(重新生成), incremental(增量更新), summary(查看摘要)")
	days := flag.Int("days", 30, "增量更新的天数")
	flag.Parse()

	// 连接数据库
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("无法打开数据库: %v", err)
	}
	defer db.Close()

	// 启用WAL模式
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		log.Printf("警告: 无法启用WAL模式: %v", err)
	}

	// 创建统计生成器
	generator := health.NewStatisticsGenerator(db)

	// 执行操作
	startTime := time.Now()

	switch *action {
	case "generate":
		fmt.Println("开始生成统计数据...")
		if err := generator.GenerateAllStatistics(); err != nil {
			log.Fatalf("生成统计失败: %v", err)
		}
		fmt.Printf("统计生成完成! 耗时: %.2f秒\n", time.Since(startTime).Seconds())

	case "regenerate":
		fmt.Println("开始重新生成统计数据(将清除旧数据)...")
		if err := generator.RegenerateStatistics(); err != nil {
			log.Fatalf("重新生成统计失败: %v", err)
		}
		fmt.Printf("统计重新生成完成! 耗时: %.2f秒\n", time.Since(startTime).Seconds())

	case "incremental":
		fmt.Printf("开始增量更新统计数据(最近%d天)...\n", *days)
		if err := generator.IncrementalUpdate(*days); err != nil {
			log.Fatalf("增量更新失败: %v", err)
		}
		fmt.Printf("增量更新完成! 耗时: %.2f秒\n", time.Since(startTime).Seconds())

	case "summary":
		fmt.Println("获取统计摘要...")
		summary, err := generator.GetStatisticsSummary()
		if err != nil {
			log.Fatalf("获取摘要失败: %v", err)
		}

		fmt.Println("\n=== 统计摘要 ===")
		fmt.Printf("总统计数: %v\n", summary["total_statistics"])

		if byType, ok := summary["by_type"].(map[string]int); ok {
			fmt.Println("\n按类型统计:")
			for statType, count := range byType {
				fmt.Printf("  %s: %d\n", statType, count)
			}
		}

		if byMetric, ok := summary["by_metric"].(map[string]int); ok {
			fmt.Println("\n按指标统计(Top 10):")
			for metric, count := range byMetric {
				fmt.Printf("  %s: %d\n", metric, count)
			}
		}

		if lastUpdate, ok := summary["last_update"].(time.Time); ok {
			fmt.Printf("\n最后更新: %s\n", lastUpdate.Format("2006-01-02 15:04:05"))
		}

	default:
		log.Fatalf("未知操作: %s (可用: generate, regenerate, incremental, summary)", *action)
	}

	// 显示最终摘要
	if *action != "summary" {
		fmt.Println("\n获取统计摘要...")
		summary, err := generator.GetStatisticsSummary()
		if err != nil {
			log.Printf("警告: 无法获取摘要: %v", err)
			return
		}

		fmt.Println("\n=== 生成结果 ===")
		fmt.Printf("总统计数: %v\n", summary["total_statistics"])

		if byType, ok := summary["by_type"].(map[string]int); ok {
			fmt.Println("\n按类型统计:")
			for statType, count := range byType {
				fmt.Printf("  %s: %d\n", statType, count)
			}
		}
	}
}
