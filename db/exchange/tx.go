package exchange

// Exchange Exchange
type Exchange struct {
	Symbol       string `json:"symbol,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
	Introduction string `json:"introduction,omitempty"`
}

// TxOption TxOption
type TxOption struct {
	// exchange
	Symbol       string `json:"symbol,omitempty"`
	Address      string `json:"address,omitempty"`
	To           string `json:"to,omitempty"`
	ExecName     string `json:"exec_name,omitempty"`
	Amount       int64  `json:"amount,omitempty"`
	Name         string `json:"name,omitempty"`
	Introduction string `json:"introduction,omitempty"`
	Total        int64  `json:"total,omitempty"`
	Note         string `json:"note,omitempty"`
}

type Order struct {
	OrderID int64 `protobuf:"varint,1,opt,name=orderID,proto3" json:"orderID,omitempty"`
	// Types that are assignable to Value:
	//	*Order_LimitOrder
	//	*Order_MarketOrder
	Value interface{} `protobuf:"value" json:"value,omitempty"`
	//挂单类型
	Ty int32 `protobuf:"varint,4,opt,name=ty,proto3" json:"ty,omitempty"`
	//已经成交的数量
	Executed int64 `protobuf:"varint,5,opt,name=executed,proto3" json:"executed,omitempty"`
	//成交均价
	AVGPrice int64 `protobuf:"varint,6,opt,name=AVG_price,json=AVGPrice,proto3" json:"AVG_price,omitempty"`
	//余额
	Balance int64 `protobuf:"varint,7,opt,name=balance,proto3" json:"balance,omitempty"`
	//状态,0 挂单中ordered， 1 完成completed， 2撤回 revoked
	Status int32 `protobuf:"varint,8,opt,name=status,proto3" json:"status,omitempty"`
	//用户地址
	Addr string `protobuf:"bytes,9,opt,name=addr,proto3" json:"addr,omitempty"`
	//更新时间
	UpdateTime int64 `protobuf:"varint,10,opt,name=updateTime,proto3" json:"updateTime,omitempty"`
	//索引
	Index int64 `protobuf:"varint,11,opt,name=index,proto3" json:"index,omitempty"`
	//手续费率
	Rate int32 `protobuf:"varint,12,opt,name=rate,proto3" json:"rate,omitempty"`
	//手续费
	DigestedFee int64 `protobuf:"varint,13,opt,name=digestedFee,proto3" json:"digestedFee,omitempty"`
	//最小手续费
	MinFee int64 `protobuf:"varint,14,opt,name=minFee,proto3" json:"minFee,omitempty"`
	//挂单hash
	Hash string `protobuf:"bytes,15,opt,name=hash,proto3" json:"hash,omitempty"`
	//撤单hash
	RevokeHash string `protobuf:"bytes,16,opt,name=revokeHash,proto3" json:"revokeHash,omitempty"`
	//创建时间
	CreateTime int64 `protobuf:"varint,17,opt,name=createTime,proto3" json:"createTime,omitempty"`
}
