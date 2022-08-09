package proofconfig

import (
	"encoding/json"
	"fmt"

	"github.com/33cn/externaldb/db"
)

// db names
const (
	DBX          = "proof_config"
	TableX       = "proof_config"
	OrgDBX       = "proof_config_org"
	OrgTableX    = "proof_config_org"
	DeleteDBX    = "proof_config_delete"
	DeleteTableX = "proof_config_delete"
	DefaultType  = "_doc"
)

const (
	// SuperX 管理员组织
	SuperX = "system"

	// ManagerX 角色
	ManagerX = "manager"
	// MemberX  角色
	MemberX = "member"

	// AddOpX 操作
	AddOpX = "add"
	// DeleteOpX 操作
	DeleteOpX = "delete"
	// SetOpX 操作
	SetOpX = "set"

	// NameX 插件对应的虚拟合约的后缀
	NameX = "config"
)

// organization
// ID: org-{name}

// member
// ID: addr-{address}

// config payload
type payload struct {
	Op               string          `json:"op"`
	Organization     string          `json:"organization"`
	Role             string          `json:"role"`
	Address          string          `json:"address"`
	OrganizationNote string          `json:"organization_note"`
	AddressNote      string          `json:"address_note"`
	ServerName       string          `json:"server_name"` // 来源系统名称
	UserDetail       *UserDetail     `json:"user_detail"`
	PersonalAuth     *PersonalAuth   `json:"personal_auth"`
	EnterpriseAuth   *EnterpriseAuth `json:"enterprise_auth"`
}

// 可能有两条/3条记录: org, addr, tx

type memberDel struct {
	Address      string `json:"address"`
	Role         string `json:"role"`
	Organization string `json:"organization"`
	Note         string `json:"note"`
	ServerName   string `json:"server_name"`
	// 区块描述
	db.Block
	Del db.Block `json:"delete"`
}

func (m *memberDel) ID() string {
	return memberDelID(m.Address, m.Del.Height, m.Del.Index)
}

func memberDelID(addr string, h, i int64) string {
	return "member-del-" + fmt.Sprintf("%d", db.HeightIndex(h, i)) + "-" + addr
}

func (m *memberDel) genRecord(orgOp string, blockOp int) *dbMemberDel {
	// 只有在删除成员操作中才能触发
	// assert(orgOp, OpDelete)
	return &dbMemberDel{
		IKey: db.NewIKey(DeleteDBX, DeleteTableX, m.ID()),
		Op:   db.NewOp(blockOp),
		M:    m,
	}
}

// Organization 组织结构体
type Organization struct {
	Organization string `json:"organization"`
	Note         string `json:"note"`
	Count        int32  `json:"count"`
	// 区块描述
	db.Block
}

// ID 获得ID
func (m *Organization) ID() string {
	return OrganizationID(m.Organization)
}

// OrganizationID 获得ID
func OrganizationID(org string) string {
	return "org-" + org
}

func (m *Organization) genRecord(orgOp string, blockOp int) *dbOrganization {
	op := db.OpUpdate
	// 初次创建的情况和回滚的情况
	if m.Count == 1 && (orgOp == AddOpX || orgOp == SetOpX) && blockOp == db.OpAdd {
		op = db.OpAdd
	} else if m.Count == 0 && (orgOp == AddOpX || orgOp == SetOpX) && blockOp == db.OpDel {
		op = db.OpDel
	}

	dbOrg := dbOrganization{
		IKey: db.NewIKey(OrgDBX, OrgTableX, m.ID()),
		Op:   db.NewOp(op),
		M:    m,
	}

	return &dbOrg
}

// UserDetail 用户信息
type UserDetail struct {
	UserName string `json:"user_name,omitempty"` // 用户名
	UserIcon string `json:"user_icon,omitempty"` // 用户头像url
	Phone    string `json:"phone,omitempty"`     // 手机号
	Email    string `json:"email,omitempty"`     // 邮箱
	AuthType int32  `json:"auth_type,omitempty"` // 认证类型(1:个人认证，2：企业认证)
}

// PersonalAuth 银行认证信息
type PersonalAuth struct {
	RealName string `json:"real_name,omitempty"` // 真实姓名
}

// EnterpriseAuth 企业认证信息
type EnterpriseAuth struct {
	EnterpriseName string `json:"enterprise_name,omitempty"` // 企业名称
}

// Member 组织成员结构体
type Member struct {
	Address      string `json:"address"`
	Role         string `json:"role"`
	Organization string `json:"organization"`
	Note         string `json:"note"`
	ServerName   string `json:"server_name"`
	*UserDetail
	*PersonalAuth
	*EnterpriseAuth
	// 区块描述
	db.Block
}

// GetUserName 获取用户名
func (m *Member) GetUserName() string {
	if m.UserDetail != nil {
		return m.UserName
	}
	return ""
}

// GetUserRealName 获取真是姓名
func (m *Member) GetUserRealName() string {
	if m.PersonalAuth != nil {
		return m.RealName
	}
	return ""
}

// ID 获得ID
func (m *Member) ID() string {
	return MemberID(m.Address)
}

// MemberID 获得ID
func MemberID(addr string) string {
	return "member-" + addr
}

// 如果是回滚, 操作设置为反向
func rollbackOp(orgOp string, blockOp int) int {
	if blockOp == db.SeqTypeDel {
		if orgOp == AddOpX {
			return db.OpDel
		}
		return db.OpAdd
	}
	if orgOp == AddOpX {
		return db.OpAdd
	}
	return db.OpDel

}

func (m *Member) genRecord(orgOp string, blockOp int) *dbMember {
	op := rollbackOp(orgOp, blockOp)
	dbMem := dbMember{
		IKey: db.NewIKey(DBX, TableX, m.ID()),
		Op:   db.NewOp(op),
		M:    m,
	}

	return &dbMem
}

// SuperManager 超级管理员
func (m *Member) SuperManager() bool {
	return m.Role == ManagerX && m.Organization == SuperX
}

// PrivilegeToManage 普通管理权限: 包括普通管理员和超级管理员
func (m *Member) PrivilegeToManage(org string) bool {
	return m.Role == ManagerX && (m.Organization == org || m.Organization == SuperX)
}

// PrivilegeToProof 管理存证: 包括自己和普通管理员和超级管理员
func (m *Member) PrivilegeToProof(org, proofOwner string) bool {
	return (m.Address == proofOwner && m.Organization == org) || m.PrivilegeToManage(org)
}

// // delete tx payload
// type deleteP struct {
// 	TxHash string `json:"tx_hash"`
// 	Note   string `json:"note"`
// }

// 可能有两条/1条记录: del-一条, tx
// 删除回滚 要去取 原始交易再次生成
type dbMember struct {
	*db.IKey
	*db.Op

	M *Member
}

func (r *dbMember) Value() []byte {
	v, _ := json.Marshal(r.M)
	return v
}

type dbOrganization struct {
	*db.IKey
	*db.Op

	M *Organization
}

func (r *dbOrganization) Value() []byte {
	v, _ := json.Marshal(r.M)
	return v
}

type dbMemberDel struct {
	*db.IKey
	*db.Op

	M *memberDel
}

func (r *dbMemberDel) Value() []byte {
	v, _ := json.Marshal(r.M)
	return v
}
