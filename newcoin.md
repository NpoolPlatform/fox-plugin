# 币种交易

## 交易流程标准

基础检查：

- 节点可用性（检查高度、节点健康状态）
- From To 地址合法性
- Value 合法性（精度、能否解析）

交易流程：

- 交易签名前准备
  - 链上估计gasPrice
  - 检查 Gas + Value < Balance
  - 从链上From账户上下文（如Nonce、最近块儿高度等），此外个别币种还需要先获取From的ViewKey
- 交易签名
  - 获取私钥信息，签名
  - 构造已签名待上链交易
- 交易广播
  - 广播到链上
  - 拿到交易ID
- 等待交易确认
  - 用交易ID重复检查交易状态

## 其他功能

estimateGas 目前只有eth链上的支持