// netlify/functions/get-users/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jackc/pgx/v5"
	"os"
	"time" // 引入 time 包
)

// User 结构体保持不变
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// 1. 更新 API 响应结构体，以包含两个独立的计时
type APIResponse struct {
	Data           []User `json:"data"`
	ConnectionTime string `json:"connection_time_ms"`
	QueryTime      string `json:"query_time_ms"`
}

func handler(ctx context.Context) (*events.APIGatewayProxyResponse, error) {
	// 从环境变量中安全地读取数据库连接字符串
	connString := os.Getenv("SUPABASE_DATABASE_URL")
	if connString == "" {
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: "SUPABASE_DATABASE_URL not set"}, nil
	}

	// 2. 测量数据库连接耗时
	connStartTime := time.Now()
	conn, err := pgx.Connect(ctx, connString)
	connDuration := time.Since(connStartTime) // 计算连接耗时

	if err != nil {
		// 即使连接失败，也可以记录耗时（如果需要的话）
		body := fmt.Sprintf("Unable to connect to database (took %s): %v", connDuration, err)
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: body}, nil
	}
	defer conn.Close(ctx)

	// 3. 测量 SQL 查询执行耗时
	queryStartTime := time.Now()
	rows, err := conn.Query(ctx, "SELECT id, username, email FROM users ORDER BY id;")
	queryDuration := time.Since(queryStartTime) // 计算查询耗时

	if err != nil {
		body := fmt.Sprintf("Query failed (took %s): %v", queryDuration, err)
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: body}, nil
	}
	defer rows.Close()

	// 处理查询结果
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email); err != nil {
			return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Failed to scan row: %v", err)}, nil
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Error during rows iteration: %v", err)}, nil
	}

	// 4. 构建包含所有信息的最终响应对象
	responseObject := APIResponse{
		Data:           users,
		ConnectionTime: fmt.Sprintf("%.2fms", float64(connDuration.Microseconds())/1000.0),
		QueryTime:      fmt.Sprintf("%.2fms", float64(queryDuration.Microseconds())/1000.0),
	}

	// 将最终对象序列化为 JSON
	responseBody, err := json.Marshal(responseObject)
	if err != nil {
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Failed to marshal response: %v", err)}, nil
	}

	// 返回成功的响应
	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(responseBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
