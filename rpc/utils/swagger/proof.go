package swagger

type ListProofResult struct {
	ServerResponse
	Result []Proof `json:"result"` // 返回结果
}

type Proof struct {
	Basehash           string      `json:"basehash"`                       // 增量存证依赖的主hash
	EvidenceName       string      `json:"evidenceName"`                   // 存证名称
	Prehash            string      `json:"prehash"`                        // 增量存证前一个hash
	ProofBlockHash     string      `json:"proof_block_hash"`               // 区块hash
	ProofBlockTime     int         `json:"proof_block_time"`               // 上链时间
	ProofData          string      `json:"proof_data"`                     // 存证数据
	ProofDeleted       string      `json:"proof_deleted"`                  // 删除存证交易hash
	ProofDeletedFlag   bool        `json:"proof_deleted_flag"`             // 删除标志
	ProofDeletedNote   string      `json:"proof_deleted_note"`             // 删除备注
	ProofHeight        int         `json:"proof_height"`                   // 存证高度
	ProofHeightIndex   int         `json:"proof_height_index"`             //  存证高度索引
	ProofID            string      `json:"proof_id"`                       // 存证id
	ProofNote          string      `json:"proof_note"`                     // 存证备注
	ProofOrganization  string      `json:"proof_organization"`             // 组织
	ProofOriginal      string      `json:"proof_original"`                 // 来源
	ProofSender        string      `json:"proof_sender"`                   // 存证发起者
	ProofTxHash        string      `json:"proof_tx_hash"`                  // 交易哈希
	SourceHash         interface{} `json:"source_hash"`                    // 依赖交易哈希
	UpdateHash         string      `json:"update_hash"`                    // 更新依赖主哈希
	UpdateVersion      int         `json:"update_version"`                 // 更新版本
	Version            int         `json:"version"`                        // 存证版本
	UserAuthType       int         `json:"user_auth_type,omitempty"`       // 用户认证类型
	UserEmail          string      `json:"user_email,omitempty"`           // 用户邮箱
	UserIcon           string      `json:"user_icon,omitempty"`            // 用户头像链接地址
	UserName           string      `json:"user_name,omitempty"`            // 用户名
	UserPhone          string      `json:"user_phone,omitempty"`           // 用户手机号
	UserRealName       string      `json:"user_real_name,omitempty"`       // 用户真是名称
	UserEnterpriseName string      `json:"user_enterprise_name,omitempty"` // 用户企业名称
}

type Template struct {
	TemplateBlockHash    string `json:"template_block_hash"`   // 区块哈希
	TemplateBlockTime    int    `json:"template_block_time"`   // 上链时间
	TemplateData         string `json:"template_data"`         // 模板数据
	TemplateDeleted      string `json:"template_deleted"`      // 删除交易哈希
	TemplateDeletedFlag  bool   `json:"template_deleted_flag"` // 删除标志
	TemplateDeletedNote  string `json:"template_deleted_note"` // 删除备注
	TemplateHeight       int    `json:"template_height"`       // 高度
	TemplateHeightIndex  int    `json:"template_height_index"` // 高度索引
	TemplateID           string `json:"template_id"`           // 模板id
	TemplateName         string `json:"template_name"`         // 模板名称
	TemplateOrganization string `json:"template_organization"` // 组织
	TemplateSender       string `json:"template_sender"`       // 交易发送人
	TemplateTxHash       string `json:"template_tx_hash"`      // 交易哈希
}

type VolunteerStatsResult struct {
	ServerResponse
	Result VolunteerStats `json:"result"` // 返回结果
}

type VolunteerStats struct {
	TermsAgges []TermsAgg `json:"termsAgges"` // 聚合
	Count      int64      `json:"count"`      // 总数
}

type TermsAgg struct {
	Count         int64           `json:"count"`         // 数量
	TermsAggKey   string          `json:"termsAggKey"`   // 聚合键值
	SubTermsAgges []SubTermsAgges `json:"subTermsAgges"` // 子聚合
}

type SubTermsAgges struct {
	SubTermsAggKey string `json:"subTermsAggKey"` // 子聚合键值
	Count          int64  `json:"count"`          // 聚合数量
}

type DonationStats struct {
	Items []DonationStatItem `json:"itemes"` // 列表
}

type DonationStatItem struct {
	Name  string `json:"name"`            // 名称
	Total int64  `json:"total,omitempty"` // 总合
	Count int    `json:"count"`           // 数量
}

type Member struct {
	Address        string `json:"address"`                   // 地址
	Role           string `json:"role"`                      // 角色
	Organization   string `json:"organization"`              // 组织
	Note           string `json:"note"`                      // 备注
	Height         int    `json:"height"`                    // 高度
	Ts             int    `json:"ts"`                        // 上链时间
	BlockHash      string `json:"block_hash"`                // 区块哈希
	Index          int    `json:"index"`                     // 交易索引号
	Send           string `json:"send"`                      // 交易发起人
	TxHash         string `json:"tx_hash"`                   // 交易hash
	HeightIndex    int64  `json:"height_index"`              // 高度索引
	UserName       string `json:"user_name,omitempty"`       // 用户名
	Phone          string `json:"phone,omitempty"`           // 手机号
	UserIcon       string `json:"user_icon,omitempty"`       // 头像地址链接
	AuthType       int    `json:"auth_type,omitempty"`       // 认证类型
	RealName       string `json:"real_name,omitempty"`       // 真实姓名
	EnterpriseName string `json:"enterprise_name,omitempty"` // 企业名称
	Email          string `json:"email,omitempty"`           // 邮箱
}

type Organization struct {
	Organization string `json:"organization"` // 组织名
	Note         string `json:"note"`         // 备注
	Count        int    `json:"count"`        // 数量
}
