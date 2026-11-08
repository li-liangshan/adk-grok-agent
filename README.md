# ADK Weather Report Agent

一个基于 Google ADK (Agent Development Kit) 框架开发的智能天气报告代理，能够查询城市天气信息并提供智能对话服务。


点击下方缩略图即可在 YouTube 上观看演示讲解：
[![YouTube 演示讲解视频](https://img.youtube.com/vi/Nw8njfjmxZM/hqdefault.jpg)](https://www.youtube.com/watch?v=Nw8njfjmxZM)

## 项目简介

这是一个使用 Go 语言开发的 AI 代理应用，集成了以下功能：

- **天气查询**：通过 OpenWeatherMap API 获取指定城市的实时天气信息
- **天气预报**：查询指定城市未来 1-5 天的每日天气预报
- **自然语言交互**：使用 DeepSeek (deepseek-flash) 大语言模型进行智能对话
- **错误处理**：友好的错误提示和异常处理机制

## 主要功能

### 天气报告查询 (`get_weather_report`)

- 根据用户提供的城市名称查询实时天气
- 返回城市名称、天气描述和温度信息（摄氏度）
- 支持错误处理和友好的错误提示
- 当查询失败时，会提示用户并提供替代建议

### 天气预报查询 (`get_weather_forecast`)

- 根据用户提供的城市名称查询未来多日天气预报
- 返回每天的日期、天气描述和最低/最高温度（摄氏度）
- `days` 参数可选，支持 1-5 天，默认 3 天
- 按天聚合 3 小时粒度的预报数据，取接近正午的天气描述

### 智能对话

- 使用 deepseek-flash 模型进行自然语言理解
- 自动识别用户意图并调用相应的工具
- 提供流畅的对话体验
- 智能处理天气查询请求和用户反馈

## 环境要求

- Go 1.25.0 或更高版本
- 有效的 API 密钥：
  - `DEEPSEEK_API_KEY`: DeepSeek API 密钥
  - `OWM_API_KEY`: OpenWeatherMap API 密钥

## 安装步骤

### 1. 克隆项目

```bash
git clone <repository-url>
cd adk-grok-agent
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置环境变量

创建 `.env` 文件或设置环境变量：

```bash
export DEEPSEEK_API_KEY="your-deepseek-api-key"
export OWM_API_KEY="your-openweathermap-api-key"
```

或者使用 `.env` 文件（需要安装支持工具）：

```bash
# .env
DEEPSEEK_API_KEY=your-deepseek-api-key
OWM_API_KEY=your-openweathermap-api-key
```

### 4. 获取 API 密钥

#### DeepSeek API 密钥
1. 访问 [DeepSeek](https://platform.deepseek.com)
2. 注册账号并获取 API 密钥

#### OpenWeatherMap API 密钥
1. 访问 [OpenWeatherMap](https://openweathermap.org/api)
2. 注册免费账号
3. 在控制台获取 API Key

## 运行项目

### 基本运行

```bash
# 启动 Web 服务（Web UI: http://localhost:8080/ui/）
go run main.go web api webui

# 或在终端中与控制台交互
go run main.go console
```

### 使用 ADK Launcher

项目使用 Google ADK 的 launcher 框架，支持多种运行模式：

```bash
# 查看帮助信息
go run main.go --help

# 运行代理（具体命令取决于 ADK launcher 配置）
go run main.go [launcher-options]
```

## 项目结构

```
adk-grok-agent/
├── main.go                 # 程序入口，初始化代理和启动器
├── agents/                 # 代理实现
│   └── weather_report.go  # 天气报告代理核心逻辑
├── models/                 # LLM 模型适配器
│   ├── deepseek.go        # DeepSeek 模型适配器
│   ├── grok.go            # Grok (x.ai) 模型适配器
│   ├── openai.go          # OpenAI 模型适配器
│   ├── openrouter.go      # OpenRouter 模型适配器
│   └── request2params.go  # 请求参数转换工具
├── go.mod                  # Go 模块依赖
├── go.sum                  # 依赖校验和
└── README.md              # 项目文档
```

## 使用示例

启动代理后，可以通过对话方式使用：

```
用户: "What's the weather in Beijing?"
代理: [调用 get_weather_report 工具]
代理: "The weather in Beijing is clear sky with a temperature of 15.5 degrees Celsius."

用户: "What about Shanghai?"
代理: [调用 get_weather_report 工具]
代理: "The weather in Shanghai is cloudy with a temperature of 18.2 degrees Celsius."

用户: "What's the weather forecast for Beijing for the next 3 days?"
代理: [调用 get_weather_forecast 工具，days=3]
代理: "Forecast for Beijing: 2026-11-08: clear sky, 9.4~17.8 degrees Celsius; ..."

用户: "The weather in Beijing is not available"
代理: [get_weather_report 返回错误状态]
代理: "I'm sorry, but the weather information for Beijing is not available. Would you like to try another city?"
```

## 故障排除

### 常见问题

1. **API 密钥错误**
   - 确保环境变量已正确设置
   - 检查 API 密钥是否有效
   - 验证环境变量名称是否正确（`DEEPSEEK_API_KEY` 和 `OWM_API_KEY`）

2. **天气查询失败**
   - 验证 OpenWeatherMap API 密钥是否有效
   - 检查城市名称是否正确（支持英文城市名）
   - 确认网络连接正常
   - 检查 API 配额是否已用完

3. **LLM 调用失败**
   - 检查 DeepSeek API 密钥是否正确
   - 确认账户有足够的配额
   - 查看日志中的详细错误信息
   - 验证网络连接和防火墙设置

4. **编译错误**
   - 确保 Go 版本符合要求（1.25.0+）
   - 运行 `go mod tidy` 更新依赖
   - 检查模块路径是否正确

## 许可证

查看 [LICENSE](LICENSE) 文件了解许可证信息。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关链接

- [Google ADK 文档](https://github.com/google/adk)
- [OpenWeatherMap API](https://openweathermap.org/api)
- [DeepSeek](https://platform.deepseek.com)
