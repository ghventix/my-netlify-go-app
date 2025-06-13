// netlify/functions/get-users/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log" // 1. 引入 log 包
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jackc/pgx/v5"
)

// ... (User 和 APIResponse 结构体保持不变)
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
type APIResponse struct {
	Data           []User `json:"data"`
	ConnectionTime string `json:"connection_time_ms"`
	QueryTime      string `json:"query_time_ms"`
}

func handler(ctx context.Context) (*events.APIGatewayProxyResponse, error) {
	connString := os.Getenv("SUPABASE_DATABASE_URL")
	if connString == "" {
		log.Println("ERROR: SUPABASE_DATABASE_URL environment variable not set!")
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: "SUPABASE_DATABASE_URL not set"}, nil
	}

	// 2. 添加决定性的诊断日志
	log.Printf("INFO: Attempting to connect using DSN: %s", connString)

	// 测量数据库连接耗时
	connStartTime := time.Now()
	conn, err := pgx.Connect(ctx, connString)
	// ... (后续代码保持不变)
	connDuration := time.Since(connStartTime)

	if err != nil {
		body := fmt.Sprintf("Unable to connect to database (took %s): %v", connDuration, err)
		log.Printf("ERROR: %s", body) // 也将错误记录到日志
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: body}, nil
	}
	defer conn.Close(ctx)

	// 测量 SQL 查询执行耗时
	queryStartTime := time.Now()
	rows, err := conn.Query(ctx, "SELECT id, username, email FROM users ORDER BY id;")
	queryDuration := time.Since(queryStartTime)

	if err != nil {
		body := fmt.Sprintf("Query failed (took %s): %v", queryDuration, err)
		log.Printf("ERROR: %s", body) // 也将错误记录到日志
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: body}, nil
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email); err != nil {
			log.Printf("ERROR: Failed to scan row: %v", err)
			return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Failed to scan row: %v", err)}, nil
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: Error during rows iteration: %v", err)
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Error during rows iteration: %v", err)}, nil
	}

	responseObject := APIResponse{
		Data:           users,
		ConnectionTime: fmt.Sprintf("%.2fms", float64(connDuration.Microseconds())/1000.0),
		QueryTime:      fmt.Sprintf("%.2fms", float64(queryDuration.Microseconds())/1000.0),
	}

	responseBody, err := json.Marshal(responseObject)
	if err != nil {
		log.Printf("ERROR: Failed to marshal response: %v", err)
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Failed to marshal response: %v", err)}, nil
	}

	log.Println("INFO: Request processed successfully.")
	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(responseBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
