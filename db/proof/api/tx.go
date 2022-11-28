package api

// 交易名称
const (
	AddProof       = "proof"
	MainAddProof   = "user.proof"
	DeleteTx       = "proof_delete"
	MainDeleteTx   = "user.proof_delete"
	RecoverTx      = "proof_recover"
	MainRecoverTx  = "user.proof_recover"
	TemplateTx     = "template"
	MainTemplateTx = "user.template"
)

// ProofInfo (tx.Payload 对应结构体的 jsonstr)
// Option, Data, Note:原始数据格式为json, 需要 base64 编码, 再设置到结构体中
//解析原理：通过version解析Option字段，然后再根据解析后的Option来解析Data
//比如通过指定的算法解密Data，或者通过指定的压缩算法解压Data数据。等等
//Data: 存证的原始数据
//Version:"1.0.0"
//Option:一个json字符串，jsonbuf := `{"encrypt":"rsa","archive":"zip"}`
//Ext:jsonbuf := `{"basehash":"1","prehash":"2","update_hash":"1","source_hash":"[\"hash1\",\"hash2\"]"}`
type ProofInfo struct {
	Data    string `json:"data"`
	Version string `json:"version"`
	Option  string `json:"option"`
	Note    string `json:"note"`
	Ext     string `json:"ext"`
}

type TemplateInfo struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// DeleteProof (tx.Payload 对应结构体的 jsonstr)删除存证时转入的id是交易hash值
type DeleteProof struct {
	ID    string `json:"id"`
	Note  string `json:"note"`
	Force bool   `json:"force"`
}

// RecoverProof (tx.Payload 对应结构体的 jsonstr)恢复存证
type RecoverProof struct {
	ID   string `json:"id"`
	Note string `json:"note"`
}

// 业务错误
const (
	ErrNotBaseProof    = "Not Base Proof"
	ErrProofDeleted    = "Proof Has Deleted"
	ErrProofNotDeleted = "Proof Is Not Deleted"
)
