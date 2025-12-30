# 🚀 stream

> **专为 LLM 流式输出设计的工程级 JSON 解析器**  
> 告别等待完整 JSON，拥抱 True-time 语义解析

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.21-blue.svg)](https://golang.org)
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


## 🎯 项目初衷

在使用 LLM 的 Stream 接口时，模型通常会以 chunk / token 的形式逐步输出字符串，但在工程实践中，我们往往希望模型最终生成的是一个 结构化的 JSON 数据。

例如，一个由多个对象组成的数组，其中每个对象都代表一条业务结果。

传统的处理方式需要 等待模型完全输出、JSON 结构完整闭合后，才能进行反序列化和后续处理。这会带来一个明显的问题：
在模型执行过程中，前端长时间没有任何可见反馈，用户只能被动等待。

本项目的初衷，是 将“流式能力”真正引入到结构化 JSON 的解析过程中。

通过对 LLM Stream 接口返回的 JSON 字符串进行 基于 chunk 的增量解析，一旦某个字段值或某个对象结构已经稳定，就立即触发对应的事件回调。
后端可以在第一时间对该对象进行加工处理，并将处理结果持续推送给前端。

最终，用户看到的不再是一次性返回的完整结果，而是：
- 基于模型输出 chunk 实时解析出的文本片段持续呈现
- 列表类数据按对象粒度逐条增加
- 明确感知系统已经开始响应，而不是长时间“卡住不动”

这使得 结构化数据的生成过程 也具备了与文本流式生成相近的实时体验。

--- 

## 🛠 社区与贡献

欢迎提出 Issue 或贡献 Pull Request！
请阅读代码注释以了解更多细节。

---

## 🏗️ 工作原理

查看 [PARSING_FLOW.md](./PARSING_FLOW.md) 了解完整的解析流程图。

---

## 📜 许可证

MIT License © 2025 — 欢迎自由使用、修改与传播。
