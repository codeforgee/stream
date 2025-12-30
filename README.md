# 🚀 stream

> **专为 LLM 流式输出设计的工程级 JSON 解析器**  
> 告别等待完整 JSON，拥抱 True-time 语义解析

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.18-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ✨ 特性

- ⚡ **True-time 解析**：字段一旦稳定立即触发，无需等待完整 JSON
- 🔄 **流式输入**：支持 chunk/token/bytes 级别的增量输入
- 📡 **事件驱动**：基于订阅制，只关注你需要的路径
- 🎯 **JSONPath 匹配**：支持 `$.items[*].id` 等路径模式（含通配符）
- 🛡️ **容错性强**：优雅处理截断、未完成的 JSON

## 🎬 安装

使用 Go Modules：

```bash
go get github.com/codeforgee/stream
```

## 🧠 使用示例

下面示例演示如何订阅字段并实时处理：

```go
package main

import (
	"fmt"
	"github.com/codeforgee/stream"
)

func main() {
	p := stream.NewParser()

	// 订阅 status 字段
	p.On("$.status", func(ev stream.Event) {
		if ev.Value != nil && ev.Value.Complete {
			fmt.Printf("状态: %q\n", ev.Value.String())
		}
	})

	// 订阅 items 数组中 id 字段
	p.On("$.items[*].id", func(ev stream.Event) {
		if ev.Value != nil && ev.Value.Complete {
			fmt.Printf("收到 ID: %d\n", ev.Value.Int64())
		}
	})

	// 模拟流式输入 JSON 片段
	fragments := []string{
		`{"status": "run`,
		`ning", "items": [`,
		`{"id": 1}, `,
		`{"id": 2}`,
		`]}`,
	}

	for _, frag := range fragments {
		p.FeedString(frag)
	}
	
	// 关闭解析器
	p.Close(true)
}
```

**输出示例：**
```
状态: running
收到 ID: 1
收到 ID: 2
```

## 📌 设计理念

在很多场景下（比如 LLM 流输出、日志聚合、HTTP chunked JSON），无法等待完整 JSON。
传统的 encoding/json 需要整个数据到齐才能解析，而 stream 库能够逐片段解析，并实时触发事件。

## 关键 API

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

## 🛠 社区与贡献

欢迎提出 Issue 或贡献 Pull Request！
请阅读代码注释以了解更多细节。

---

## 🏗️ 工作原理

查看 [PARSING_FLOW.md](./PARSING_FLOW.md) 了解完整的解析流程图。

---

## 📜 许可证

MIT License © 2025 — 欢迎自由使用、修改与传播。
