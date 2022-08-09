package swagger

type ListEVMResult struct {
	ServerResponse
	Result []EVMToken `json:"result"` // 返回结果
}

type EVMToken struct {
	Amount          int      `json:"amount"`            // 金额
	CallFuncName    string   `json:"call_func_name"`    // 调用方法名称
	ContractAddr    string   `json:"contract_addr"`     // 合约地址
	ContractUsedGas int      `json:"contract_used_gas"` // 消耗gas
	EvmBlockHash    string   `json:"evm_block_hash"`    // 区块hash
	EvmBlockTime    int      `json:"evm_block_time"`    // 上链时间
	EvmEvents       string   `json:"evm_events"`        // evm事件
	EvmHeight       int      `json:"evm_height"`        // 区块高度
	EvmHeightIndex  int64    `json:"evm_height_index"`  // 高度索引
	EvmNote         string   `json:"evm_note"`          // 备注信息
	EvmParam        string   `json:"evm_param"`         // evm调用参数
	EvmTxHash       string   `json:"evm_tx_hash"`       // 交易hash
	GoodsType       int      `json:"goods_type"`        // 物品类型
	GoodsID         int      `json:"goods_id"`          // 物品唯一标识
	LabelID         string   `json:"label_id"`          // 物品标签id
	Name            string   `json:"name"`              // 物品名称
	Owner           string   `json:"owner"`             // 拥有者
	PublishTime     int      `json:"publish_time"`      // 发布时间
	Publisher       string   `json:"publisher"`         // 发布者
	Remark          string   `json:"remark"`            // 备注
	SourceHash      []string `json:"source_hash"`       // 关联交易hash
	TraceHash       []string `json:"trace_hash"`        // 关联溯源hash
}
