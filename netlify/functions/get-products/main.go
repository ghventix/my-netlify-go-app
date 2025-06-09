// netlify/functions/get-products/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jackc/pgx/v5"
)

// Product 结构体，用于映射数据库中的 products 表
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func handler(ctx context.Context) (*events.APIGatewayProxyResponse, error) {
	// 1. 从环境变量中安全地读取数据库连接字符串
	connString := os.Getenv("NEON_DATABASE_URL")
	if connString == "" {
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: "NEON_DATABASE_URL not set"}, nil
	}

	// 2. 连接到 Neon 数据库
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Unable to connect to database: %v", err)}, nil
	}
	defer conn.Close(ctx) // 确保函数结束时关闭连接

	// 3. 执行 SQL 查询
	rows, err := conn.Query(ctx, "SELECT id, name, price FROM products ORDER BY id;")
	if err != nil {
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Query failed: %v", err)}, nil
	}
	defer rows.Close() // 确保处理完结果集后关闭它

	// 4. 处理查询结果
	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Name, &p.Price)
		if err != nil {
			return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Failed to scan row: %v", err)}, nil
		}
		products = append(products, p)
	}

	// 5. 将结果序列化为 JSON
	respBody, err := json.Marshal(products)
	if err != nil {
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Failed to marshal response: %v", err)}, nil
	}

	// 6. 返回成功的响应
	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(respBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
