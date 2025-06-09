// netlify/functions/hello/main.go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is a custom struct for our JSON response
type Response struct {
	Message string `json:"message"`
}

// handler is the main function logic
func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// 1. 从查询参数中获取 'name'
	name := request.QueryStringParameters["name"]

	// 2. 如果 'name' 为空，则提供一个默认值
	if name == "" {
		name = "World"
	}

	// 3. 构建我们的响应消息
	message := fmt.Sprintf("Hello, %s!333", name)

	// 4. 将消息封装到我们的自定义 Response 结构体中
	responseObject := Response{
		Message: message,
	}

	// 5. 将 Response 结构体序列化为 JSON 字符串
	responseBody, err := json.Marshal(responseObject)
	if err != nil {
		return nil, err
	}

	// 6. 返回一个标准的 API Gateway 响应
	return &events.APIGatewayProxyResponse{
		StatusCode: 200, // OK
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	// 将我们的 handler 函数包装成一个 Lambda 服务
	lambda.Start(handler)
}
