// netlify/functions/get-products/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time" // 1. 导入 time 包

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

// 2. 创建一个新的响应结构体，用于包含数据和执行时间
type APIResponse struct {
	Data            []Product `json:"data"`
	ExecutionTimeMS int64     `json:"execution_time_ms"`
	SQLQueryTimeMS  int64     `json:"sql_query_time_ms"`
}

func handler(ctx context.Context) (*events.APIGatewayProxyResponse, error) {
	// 3. 记录整个函数的开始时间
	totalStart := time.Now()

	// ---- SQL 代码块开始 ----
	// 4. 记录 SQL 代码块的开始时间
	sqlStart := time.Now()

	// 从环境变量中安全地读取数据库连接字符串
	connString := os.Getenv("NEON_DATABASE_URL")
	if connString == "" {
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: "NEON_DATABASE_URL not set"}, nil
	}

	// 连接到 Neon 数据库
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Unable to connect to database: %v", err)}, nil
	}
	defer conn.Close(ctx) // 确保函数结束时关闭连接

	// 执行 SQL 查询
	rows, err := conn.Query(ctx, "SELECT id, name, price FROM products ORDER BY id;")
	if err != nil {
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Query failed: %v", err)}, nil
	}
	defer rows.Close() // 确保处理完结果集后关闭它

	// 处理查询结果
	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Name, &p.Price)
		if err != nil {
			return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Failed to scan row: %v", err)}, nil
		}
		products = append(products, p)
	}

	// 5. 记录 SQL 代码块的结束，并计算耗时（单位：毫秒）
	sqlDuration := time.Since(sqlStart).Milliseconds()
	// ---- SQL 代码块结束 ----

	// 6. 准备最终的 API 响应数据
	apiResponse := APIResponse{
		Data:           products,
		SQLQueryTimeMS: sqlDuration,
		// 总执行时间将在最后计算
	}

	// 将包含时间和数据的完整结构体序列化为 JSON
	responseBody, err := json.Marshal(apiResponse)
	if err != nil {
		return &events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("Failed to marshal response: %v", err)}, nil
	}

	// 7. 在函数返回前，计算总执行时间
	totalDuration := time.Since(totalStart).Milliseconds()

	// 由于 responseBody 已经生成，我们无法直接将总时间加入其中
	// 所以我们重新构建一个 map 来生成最终的 JSON
	finalResponseMap := make(map[string]interface{})
	json.Unmarshal(responseBody, &finalResponseMap)       // 先将已有的数据解码到 map
	finalResponseMap["execution_time_ms"] = totalDuration // 添加总时间字段

	finalResponseBody, _ := json.Marshal(finalResponseMap) // 再次编码

	// 8. 返回成功的响应
	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(finalResponseBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
