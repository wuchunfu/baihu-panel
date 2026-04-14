# 消息中心

消息中心集成了一整套灵活且现代的消息分发引擎，支持多场景自动推送到外部 IM 工具。

## 消息通道

- **企业 IM**：支持集成 **企业微信** (WeCom)、**钉钉** (DingTalk)、**飞书** (Lark)。
- **个人推送到位**：支持 **Telegram** Bot、**Bark** (支持自建)、**VoceChat** (支持自建) 以及基于 **Wpush** 的推送服务。
- **公共渠道**：标准的 **SMTP 邮件** 及 **Webhook** 回调。

## 事件通知规则

- **多事件配置**：您可以灵活定义在哪些场景下触发通知，包括但不限于：
    - **任务失败**：定时任务在 Cron 触发后运行报错。
    - **任务超时**：任务由于运行过长被系统中止。
    - **登录安全**：检测到异地登录或多次密码错误。
    - **服务下线**：Agent 节点掉线提醒。

## 推送使用路径

白虎面板提供了两种不同层面的通知推送方式，满足从“自动报警”到“程序内自定义推送”的全场景需求。

---

### 路径一：任务绑定通知（零代码自动化）

这是最常用的方式，用于在定时任务执行完成后，根据结果自动发送通知。

1. **入口**：在 **「定时任务」** 页面，点击任务右侧的 **「编辑」**。
2. **配置**：在弹窗底部的 **「通知配置」** 栏目中：
    - **选择渠道**：指定发送消息的 IM 渠道。
    - **触发时机**：勾选 `成功时`、`失败时` 或 `超时时`（建议至少勾选失败和超时）。
    - **附带日志**：开启后可在消息中直接预览报错日志，支持设置截取长度。
3. **生效**：保存后，该任务每次运行结束都会按设定的逻辑自动推信。

---

### 路径二：脚本手动调用 (API)

如果您需要在脚本逻辑内部（例如：当抓取到特定数据时）主动触发通知，可以使用此方式。

#### 1. 快速获取代码
为了极大降低集成门槛，面板内置了代码生成器：
- 进入 **「消息推送」** -> **「脚本调用」** 标签。
- 页面会根据您已配置的渠道，自动生成包含 **通知 Token** 和 **渠道 ID** 的完整代码示例。
- 支持 **Python**、**Node.js** 和 **Shell (Curl)** 格式，直接复制即可使用。

#### 2. 代码参考示例
如果您需要手动编写逻辑，请参考以下实现：

##### Python 示例
```python
import requests

def send_notify(title, content):
    url = "http://localhost:8052/api/v1/notify/send"
    headers = { "notify-token": "您的_NOTIFY_TOKEN" }
    data = {
        "channel_id": "您的_渠道_ID",
        "title": title,
        "text": content
    }
    requests.post(url, headers=headers, json=data)
```

##### Node.js 示例
```javascript
const axios = require('axios');

async function sendNotify(title, content) {
  await axios.post('http://localhost:8052/api/v1/notify/send', {
    channel_id: '您的_渠道_ID',
    title: title,
    text: content
  }, {
    headers: { 'notify-token': '您的_NOTIFY_TOKEN' }
  });
}
```

##### Shell 示例
```bash
curl -X POST "http://localhost:8052/api/v1/notify/send" \
  -H "notify-token: 您的_NOTIFY_TOKEN" \
  -d '{"channel_id":"渠道ID", "title":"标题", "text":"内容"}'
```

---

## 消息中心管理

除了配置发送路径，您还可以在消息中心进行以下操作：

- **发送记录 (审计)**：实时记录每一条通过白虎面板发送至外部的消息，方便追溯。
- **回执查询**：在 **「消息日志」** 页面查看到每条推送的详细状态，如果发送失败，会提供原始的错误响应代码以供排查。
