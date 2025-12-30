# 🚀 stream

> **专为 LLM 流式输出设计的工程级 JSON 解析器**  
> 告别等待完整 JSON，拥抱 True-time 语义解析

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.18-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ✨ 核心特性

- ⚡ **True-time 解析**：字段一旦稳定立即触发，无需等待完整 JSON
- 🔄 **流式输入**：支持 chunk/token/bytes 级别的增量输入
- 📡 **事件驱动**：基于订阅制，只关注你需要的路径
- 🎯 **JSONPath 匹配**：支持 `$.items[*].id` 等路径模式（含通配符）
- 🛡️ **容错性强**：优雅处理截断、未完成的 JSON

## 🎬 快速开始

### 安装

```bash
go get github.com/codeforgee/stream
```

### 使用示例

当 LLM 生成的内容是 JSON 片段时（可能被截断），用 stream 库实时解析：

```go
package main

import (
	"fmt"
	"github.com/codeforgee/stream"
)

func main() {
	p := stream.NewParser()

	// 订阅业务字段 - 一旦完成立即处理
	p.On("$.status", func(ev stream.Event) {
		if ev.Value != nil && ev.Value.Complete {
			fmt.Printf("✅ 状态: %s\n", ev.Value.String())
		}
	})

	p.On("$.items[*].id", func(ev stream.Event) {
		if ev.Value != nil && ev.Value.Complete {
			fmt.Printf("📦 收到 ID: %d\n", ev.Value.Int64())
		}
	})

	// 模拟 LLM 流式发送的 JSON 片段（可能被截断）
	chunks := []string{
		`{"status": "run`,      // 被截断
		`ning", "items": [`,    // 继续
		`{"id": 1}, `,          // 第一个 item
		`{"id": 2}`,            // 第二个 item
		`]}`,
	}

	// 流式解析每个片段
	for _, chunk := range chunks {
		p.FeedString(chunk)
	}
	p.Close(true)
}
```

**输出：**
```
✅ 状态: running
📦 收到 ID: 1
📦 收到 ID: 2
```

**关键点：**
- 即使 JSON 片段被截断（如 `{"status": "run`），也能实时处理已解析的部分
- 字段一旦稳定立即触发，无需等待完整 JSON

### 关键 API

```go
// 创建解析器
p := stream.NewParser()

// 订阅路径（支持通配符）
p.On("$.field", func(ev stream.Event) {
    if ev.Value != nil && ev.Value.Complete {
        fmt.Println(ev.Value.String())
    }
})

// 流式输入
p.Feed([]byte(chunk))
p.FeedString(chunk)

// 关闭解析器
p.Close(true)  // true = 正常结束, false = 中断
```

**支持的路径格式：**
- `$.field` - 对象字段
- `$.items[*].id` - 数组通配符
- `$.data.items[0].name` - 嵌套路径

---

## 🏗️ 工作原理

查看 [PARSING_FLOW.md](./PARSING_FLOW.md) 了解完整的解析流程图。

---

## 📄 许可证

MIT License

---

**Made with ❤️ for the LLM community**
