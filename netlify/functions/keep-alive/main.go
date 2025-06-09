// netlify/functions/keep-alive/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jackc/pgx/v5"
)

// handleRequest是函数的核心逻辑。
// 注意：它的签名比API函数更简单，因为它不处理HTTP事件。
func handleRequest(ctx context.Context) error {
	log.Println("Running keep-alive function to ping Neon database...")

	// 从环境变量安全地读取数据库连接字符串
	connString := os.Getenv("NEON_DATABASE_URL")
	if connString == "" {
		log.Println("Error: NEON_DATABASE_URL environment variable not set.")
		return fmt.Errorf("NEON_DATABASE_URL not set")
	}

	// 连接数据库
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		log.Printf("Error: Unable to connect to database: %v\n", err)
		return err
	}
	defer conn.Close(ctx)

	// 执行一个最轻量的查询，目的仅仅是“触摸”一下数据库
	var result int
	err = conn.QueryRow(ctx, "SELECT 1;").Scan(&result)
	if err != nil {
		log.Printf("Error: Keep-alive query failed: %v\n", err)
		return err
	}

	// 成功的日志对于验证至关重要
	log.Printf("Keep-alive ping to Neon successful. Query result: %d", result)
	return nil
}

func main() {
	// 启动 Lambda 处理程序
	lambda.Start(handleRequest)
}
