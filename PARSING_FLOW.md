# JSON 流式解析系统流程图

## 系统架构概览

```
+--------------+      +--------------+      +-------------+      +---------------+
| Input Stream | ---> |   Tokenizer  | ---> |    Parser   | ---> | Output Events |
+--------------+      +--------------+      +-------------+      +---------------+
       |                    |                     |                     |
       | Char: { " s t a t u s " : " r u n n i n g " }                  |
       |                    |                     |                     |
       +--------------------+---------------------+---------------------+
                            |                     |
                    Token: LBrace             Event: ObjectStart
                    Token: StringChunk        Event: FieldValue
                    Token: StringEnd          Event: FieldValue (Complete)
                    Token: Colon              ...
                    Token: StringChunk        Event: ObjectEnd
                    Token: StringEnd
                    Token: RBrace
```

## ASCII 动画流程图

### 示例：解析 `{"status": "running"}`

```
+-------------------------------------------------------------------------+
|                         Input Character Stream                          |
|  {  "  s  t  a  t  u  s  "  :  "  r  u  n  n  i  n  g  "  }             |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
|                      Tokenizer (Lexical Analysis)                       |
+-------------------------------------------------------------------------+
|  Char: {                                                                |
|    -> Token: TokenLBrace                                                |
|                                                                         |
|  Char: "                                                                |
|    -> State: tString (enter string state)                               |
|                                                                         |
|  Char: s t a t u s                                                      |
|    -> Token: TokenStringChunk("s")                                      |
|    -> Token: TokenStringChunk("t")                                      |
|    -> Token: TokenStringChunk("a")                                      |
|    -> Token: TokenStringChunk("t")                                      |
|    -> Token: TokenStringChunk("u")                                      |
|    -> Token: TokenStringChunk("s")                                      |
|                                                                         |
|  Char: "                                                                |
|    -> Token: TokenStringEnd                                             |
|    -> State: tIdle (return to idle state)                               |
|                                                                         |
|  Char: :                                                                |
|    -> Token: TokenColon                                                 |
|                                                                         |
|  Char: "                                                                |
|    -> State: tString                                                    |
|                                                                         |
|  Char: r u n n i n g                                                    |
|    -> Token: TokenStringChunk("r")                                      |
|    -> Token: TokenStringChunk("u")                                      |
|    -> Token: TokenStringChunk("n")                                      |
|    -> Token: TokenStringChunk("n")                                      |
|    -> Token: TokenStringChunk("i")                                      |
|    -> Token: TokenStringChunk("n")                                      |
|    -> Token: TokenStringChunk("g")                                      |
|                                                                         |
|  Char: "                                                                |
|    -> Token: TokenStringEnd                                             |
|                                                                         |
|  Char: }                                                                |
|    -> Token: TokenRBrace                                                |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
|                      Parser (Syntactic Analysis)                        |
+-------------------------------------------------------------------------+
|  Token: TokenLBrace                                                     |
|    -> onObjectStart()                                                   |
|    -> Stack.push(frameObject)                                           |
|    -> State: pObjExpectKey                                              |
|    -> Emit: EventObjectStart                                            |
|                                                                         |
|  Token: TokenStringChunk("s"), TokenStringChunk("t"), ...               |
|    -> onStringChunk("s")                                                |
|    -> curString.append("s")                                             |
|    -> (accumulate string, no event emitted)                             |
|                                                                         |
|  Token: TokenStringEnd                                                  |
|    -> onStringEnd()                                                     |
|    -> Current state: pObjExpectKey (this is a key)                      |
|    -> frame.key = "status"                                              |
|    -> curString.reset()                                                 |
|    -> State: pObjAfterKey                                               |
|                                                                         |
|  Token: TokenColon                                                      |
|    -> onColon()                                                         |
|    -> State: pObjExpectValue                                            |
|                                                                         |
|  Token: TokenStringChunk("r")                                           |
|    -> onStringChunk("r")                                                |
|    -> curString.append("r")                                             |
|    -> chunkBuffer.append("r")                                           |
|    -> curValueKind = valString                                          |
|    -> Emit: EventFieldValue(Value="r", Append=true)                     |
|                                                                         |
|  Token: TokenStringChunk("u"), TokenStringChunk("n"), ...               |
|    -> onStringChunk("u")                                                |
|    -> curString.append("u")                                             |
|    -> chunkBuffer.append("u")                                           |
|    -> Emit: EventFieldValue(Value="u", Append=true)                     |
|    ... (continue accumulating)                                          |
|                                                                         |
|  Token: TokenStringEnd                                                  |
|    -> onStringEnd()                                                     |
|    -> Current state: pObjExpectValue (this is a value)                  |
|    -> Emit: EventFieldValue(Value="running", Complete=true)             |
|    -> curString.reset()                                                 |
|    -> chunkBuffer.reset()                                               |
|    -> curValueKind = valNone                                            |
|    -> State: pObjAfterValue                                             |
|                                                                         |
|  Token: TokenRBrace                                                     |
|    -> onObjectEnd()                                                     |
|    -> Emit: EventObjectEnd                                              |
|    -> Stack.pop()                                                       |
|    -> State: pIdle                                                      |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
|                 Path Matching and Event Dispatch                        |
+-------------------------------------------------------------------------+
|  Event: EventObjectStart                                                |
|    -> Build PathSegments: [] (root path)                                |
|    -> Path: "$"                                                         |
|    -> Match subscriptions: $.status, $.*, ...                           |
|    -> Invoke matched handlers                                           |
|                                                                         |
|  Event: EventFieldValue(Path="$.status", Value="running")               |
|    -> Build PathSegments: [Field("status")]                             |
|    -> Path: "$.status"                                                  |
|    -> Match subscription: $.status  (matched)                           |
|    -> Invoke handler(event)                                             |
|                                                                         |
|  Event: EventObjectEnd                                                  |
|    -> Build PathSegments: [] (root path)                                |
|    -> Path: "$"                                                         |
|    -> Match subscriptions and invoke handlers                           |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
|                             Output Events                               |
|  EventObjectStart                                                       |
|  EventFieldValue(Path="$.status", Value="running", Complete=true)       |
|  EventObjectEnd                                                         |
+-------------------------------------------------------------------------+
```

## 完整解析流程图

```mermaid
graph TB
    Start([开始]) --> Tokenizer[Tokenizer: 字符 -> Token]
    Tokenizer --> TokenType{Token 类型?}
    TokenType -->|"{"| ObjStart[onObjectStart<br/>状态: pObjExpectKey<br/>输出: EventObjectStart]
    TokenType -->|"}"| ObjEnd[onObjectEnd<br/>状态: pObjAfterValue<br/>输出: EventObjectEnd]
    TokenType -->|"["| ArrStart[onArrayStart<br/>状态: pArrExpectValue<br/>输出: EventArrayStart]
    TokenType -->|"]"| ArrEnd[onArrayEnd<br/>状态: pArrAfterValue<br/>输出: EventArrayEnd]
    TokenType -->|":"| Colon[onColon<br/>状态: pObjExpectValue]
    TokenType -->|","| Comma[onComma<br/>状态转换]
    TokenType -->|StringChunk| StringChunk[onStringChunk<br/>累积到 curString]
    TokenType -->|StringEnd| StringEnd[onStringEnd<br/>处理字符串完成]
    TokenType -->|NumberChunk| NumberChunk[onNumberChunk<br/>累积到 curNumber]
    TokenType -->|NumberEnd| NumberEnd[onNumberEnd<br/>处理数字完成]
    TokenType -->|Bool/Null| Primitive[onPrimitive<br/>处理原始值]
    
    ObjStart --> StackPush1[Stack.push frameObject]
    ArrStart --> StackPush2[Stack.push frameArray]
    
    StringEnd --> StringEndType{当前状态?}
    StringEndType -->|pObjExpectKey| KeyProcess[处理 Key<br/>保存到 frame.key<br/>状态: pObjAfterKey]
    StringEndType -->|pObjExpectValue| ValueProcess[处理 Value<br/>输出: EventFieldValue<br/>状态: pObjAfterValue]
    StringEndType -->|pArrExpectValue| ArrValueProcess[处理数组值<br/>输出: EventFieldValue<br/>状态: pArrAfterValue]
    
    KeyProcess --> Colon
    ValueProcess --> AdvanceValue[advanceAfterValue]
    ArrValueProcess --> AdvanceValue
    
    NumberEnd --> NumberEmit[输出: EventFieldValue<br/>Value: Number]
    NumberEmit --> AdvanceValue
    
    Primitive --> PrimitiveEmit[输出: EventFieldValue<br/>Value: Bool/Null]
    PrimitiveEmit --> AdvanceValue
    
    AdvanceValue --> AdvanceType{父级类型?}
    AdvanceType -->|frameObject| ObjAfterValue[状态: pObjAfterValue]
    AdvanceType -->|frameArray| ArrItem[输出: EventArrayItem<br/>index++<br/>状态: pArrAfterValue]
    
    ObjEnd --> StackPop1[Stack.pop]
    ArrEnd --> StackPop2[Stack.pop]
    
    StackPop1 --> CheckParent1{父级是数组?}
    StackPop2 --> CheckParent2{父级是数组?}
    
    CheckParent1 -->|是| ArrItemFromObj[输出: EventArrayItem<br/>index++]
    CheckParent2 -->|是| ArrItemFromArr[输出: EventArrayItem<br/>index++]
    
    ArrItemFromObj --> EndState1[状态: pArrAfterValue]
    ArrItemFromArr --> EndState2[状态: pArrAfterValue]
    CheckParent1 -->|否| EndState3[状态: pObjAfterValue/pArrAfterValue]
    CheckParent2 -->|否| EndState3
    
    Comma --> CommaType{当前 frame 类型?}
    CommaType -->|frameObject| CommaObj[状态: pObjExpectKey]
    CommaType -->|frameArray| CommaArr[状态: pArrExpectValue]
    
    EndState1 --> PathMatch[路径匹配<br/>match pattern segments]
    EndState2 --> PathMatch
    EndState3 --> PathMatch
    ObjAfterValue --> PathMatch
    ArrItem --> PathMatch
    ArrItemFromObj --> PathMatch
    ArrItemFromArr --> PathMatch
    CommaObj --> PathMatch
    CommaArr --> PathMatch
    
    PathMatch --> Emit[emit Event<br/>调用匹配的 Handler]
    Emit --> NextToken{还有字符?}
    
    NextToken -->|是| Tokenizer
    NextToken -->|否| AutoEnd[自动检测 JSON 完整<br/>输出: EventStreamEnd]
    AutoEnd --> End([结束])
    
    ObjEnd -->|Stack 为空| CheckComplete{检查 JSON 完整?}
    ArrEnd -->|Stack 为空| CheckComplete
    CheckComplete -->|完整| AutoEnd
    CheckComplete -->|不完整| NextToken
```

## 详细状态转换图

```mermaid
stateDiagram-v2
    [*] --> Idle: 初始化
    Idle --> ObjExpectKey: 遇到 {
    Idle --> ArrExpectValue: 遇到 [
    ObjExpectKey --> ObjAfterKey: 字符串结束(Key)
    ObjAfterKey --> ObjExpectValue: 遇到冒号
    ObjExpectValue --> ObjAfterValue: 值完成
    ObjAfterValue --> ObjExpectKey: 遇到 ,
    ObjAfterValue --> Idle: 遇到 }
    
    ArrExpectValue --> ArrAfterValue: 值完成
    ArrAfterValue --> ArrExpectValue: 遇到 ,
    ArrAfterValue --> Idle: 遇到 ]
    
    note right of ObjExpectKey
        等待对象字段名
    end note
    
    note right of ObjExpectValue
        等待字段值
    end note
    
    note right of ArrExpectValue
        等待数组元素
    end note
```

## Tokenizer 词法分析流程

```mermaid
graph LR
    Input[输入字符] --> State{当前状态}

    State -->|Idle| IdleProcess{字符类型}
    State -->|String| StringProcess{字符类型}
    State -->|Number| NumberProcess{字符类型}
    State -->|Keyword| KeywordProcess{字符类型}

    IdleProcess -->|SingleCharToken| SingleToken[单字符 Token<br/>立即输出]
    IdleProcess -->|QUOTE| StringState[进入 String 状态]
    IdleProcess -->|DigitOrMinus| NumberState[进入 Number 状态<br/>输出 NumberChunk]
    IdleProcess -->|t/f/n| KeywordState[进入 Keyword 状态]

    StringProcess -->|NormalChar| StringChunk[输出 StringChunk]
    StringProcess -->|ESCAPE| EscapeState[进入 Escape 状态]
    StringProcess -->|QUOTE| StringEnd[输出 StringEnd<br/>返回 Idle]

    EscapeState --> StringProcess[输出 StringChunk<br/>返回 String]

    NumberProcess -->|DigitOrExp| NumberChunk[输出 NumberChunk]
    NumberProcess -->|Other| NumberEnd[输出 NumberEnd<br/>返回 Idle<br/>重新处理字符]

    KeywordProcess -->|Letter| KeywordAppend[累积到 buf]
    KeywordProcess -->|NonLetter| KeywordEnd[输出 Bool/Null<br/>返回 Idle<br/>重新处理字符]

    SingleToken --> Parser[Parser 处理]
    StringChunk --> Parser
    StringEnd --> Parser
    NumberChunk --> Parser
    NumberEnd --> Parser
    KeywordEnd --> Parser
```

## 示例：解析 `{"status": "running", "progress": 42}`

### 步骤详解

```
输入字符流: {"status": "running", "progress": 42}
```

| 步骤 | 字符 | Tokenizer 输出 | Parser 状态 | Stack | 输出事件 |
|------|------|----------------|-------------|-------|----------|
| 1 | `{` | TokenLBrace | pObjExpectKey | `[frameObject]` | EventObjectStart |
| 2 | `"` | - | pObjExpectKey | `[frameObject]` | - |
| 3 | `s` | TokenStringChunk("s") | pObjExpectKey | `[frameObject]` | - |
| 4 | `t` | TokenStringChunk("t") | pObjExpectKey | `[frameObject]` | - |
| 5 | `a` | TokenStringChunk("a") | pObjExpectKey | `[frameObject]` | - |
| 6 | `t` | TokenStringChunk("t") | pObjExpectKey | `[frameObject]` | - |
| 7 | `u` | TokenStringChunk("u") | pObjExpectKey | `[frameObject]` | - |
| 8 | `s` | TokenStringChunk("s") | pObjExpectKey | `[frameObject]` | - |
| 9 | `"` | TokenStringEnd | pObjAfterKey | `[frameObject{key:"status"}]` | - |
| 10 | `:` | TokenColon | pObjExpectValue | `[frameObject{key:"status"}]` | - |
| 11 | `"` | - | pObjExpectValue | `[frameObject{key:"status"}]` | - |
| 12 | `r` | TokenStringChunk("r") | pObjExpectValue | `[frameObject{key:"status"}]` | EventFieldValue(Value="r", Append=true) |
| 13 | `u` | TokenStringChunk("u") | pObjExpectValue | `[frameObject{key:"status"}]` | EventFieldValue(Value="u", Append=true) |
| 14 | `n` | TokenStringChunk("n") | pObjExpectValue | `[frameObject{key:"status"}]` | EventFieldValue(Value="n", Append=true) |
| 15 | `n` | TokenStringChunk("n") | pObjExpectValue | `[frameObject{key:"status"}]` | EventFieldValue(Value="n", Append=true) |
| 16 | `i` | TokenStringChunk("i") | pObjExpectValue | `[frameObject{key:"status"}]` | EventFieldValue(Value="i", Append=true) |
| 17 | `n` | TokenStringChunk("n") | pObjExpectValue | `[frameObject{key:"status"}]` | EventFieldValue(Value="n", Append=true) |
| 18 | `g` | TokenStringChunk("g") | pObjExpectValue | `[frameObject{key:"status"}]` | EventFieldValue(Value="g", Append=true) |
| 19 | `"` | TokenStringEnd | pObjAfterValue | `[frameObject{key:"status"}]` | EventFieldValue(Value="running", Complete=true) |
| 20 | `,` | TokenComma | pObjExpectKey | `[frameObject{key:"status"}]` | - |
| 21 | `"` | - | pObjExpectKey | `[frameObject{key:"status"}]` | - |
| 22 | `p` | TokenStringChunk("p") | pObjExpectKey | `[frameObject{key:"status"}]` | - |
| ... | ... | ... | ... | ... | ... |
| 23 | `"` | TokenStringEnd | pObjAfterKey | `[frameObject{key:"progress"}]` | - |
| 24 | `:` | TokenColon | pObjExpectValue | `[frameObject{key:"progress"}]` | - |
| 25 | `4` | TokenNumberChunk("4") | pObjExpectValue | `[frameObject{key:"progress"}]` | - |
| 26 | `2` | TokenNumberChunk("2") | pObjExpectValue | `[frameObject{key:"progress"}]` | - |
| 27 | `}` | TokenNumberEnd | pObjAfterValue | `[frameObject{key:"progress"}]` | EventFieldValue(Value="42", Complete=true) |
| 28 | `}` | TokenRBrace | pIdle | `[]` | EventObjectEnd, EventStreamEnd |

## 路径匹配与事件分发

```mermaid
graph TB
    Event[Event 创建] --> Segments[构建 PathSegments<br/>从 Stack 生成]
    
    Segments --> MatchLoop[遍历所有订阅]
    
    MatchLoop --> MatchCheck{Pattern 匹配?}
    
    MatchCheck -->|匹配| Handler[调用 Handler<br/>sub.Handler ev]
    MatchCheck -->|不匹配| NextSub[下一个订阅]
    
    Handler --> NextSub
    NextSub --> MoreSubs{还有订阅?}
    
    MoreSubs -->|是| MatchLoop
    MoreSubs -->|否| Done[完成]
```