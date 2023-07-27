package proofconfig

import (
	"encoding/json"
	"errors"

	"github.com/33cn/chain33/common"
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/transaction"
	"github.com/33cn/externaldb/util"
)

var log = l.New("module", "db.proof_config")

// Convert 配置插件实现
type Convert struct {
	env      *db.TxEnv
	block    *db.Block
	title    string // 作为计算execor 的前缀
	orgTitle string
	db       db.WrapDB
}

// SetDB SetDB
func (c *Convert) setDB(db db.WrapDB) error {
	c.db = db
	return nil
}

// InitDB 创建db
func (c *Convert) initDB(cli db.DBCreator) error {
	err := util.InitIndex(cli, OrgDBX, OrgTableX, OrgMapping)
	if err != nil {
		return err
	}
	err = util.InitIndex(cli, DBX, TableX, MemberMapping)
	if err != nil {
		return err
	}

	return util.InitIndex(cli, DeleteDBX, DeleteTableX, DelMembleMapping)
}

// errors
var (
	ErrBadParams    = "Bad Parameters"
	ErrNoPrivilege  = "No Privilege"
	ErrDB           = "DB Error"
	ErrMemberExists = "Member exists"
	ErrMemberCount  = "organization member count error"
)

// PrivilegeDB 存证插件需要
type PrivilegeDB interface {
	IsHaveProofPermission(addr string) bool
	IsHaveDelProofPermission(send, proofOrg, proofOwner string) bool
	GetOrganizationName(addr string) (string, error)
}

// ConfigDB db 访问, 在插件执行时, 需要的db操作
type ConfigDB interface {
	GetMember(addr string) (*Member, error)
	GetMemberDel(addr string, h, i int64) (*memberDel, error)
	GetOrganization(org string) (*Organization, error)
	SetMember(addr string, m *dbMember) error
	SetMemberDel(addr string, m *dbMemberDel) error
	SetOrganization(org string, o *dbOrganization) error

	// MemberPrivilege 存证插件需要
	PrivilegeDB
}

// ConvertTx for tx convert
func (c *Convert) convertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	// tx and payload check
	tx := env.Block.Block.Txs[env.TxIndex]
	var cfg payload
	err := json.Unmarshal(tx.Payload, &cfg)
	if err != nil {
		log.Error("parse payload failed", "err", err)
		return nil, nil
	}

	err = checkInput(cfg)
	if err != nil {
		log.Error("check config payload", "err", err)
		return nil, nil
	}

	myDB := NewConfigDB(c.db)

	if OpenAccessControl {
		from := util.AddressConvert(tx.From())
		m, err := myDB.GetMember(from)
		if err != nil {
			if err == db.ErrDBNotFound {
				log.Info("send not find", "send", from, "err", errors.New(ErrNoPrivilege))
				return nil, nil
			}
			log.Error("get send from db", "err", err)
			return nil, errors.New(ErrDB)
		}
		if !m.PrivilegeToManage(cfg.Organization) {
			log.Error("config must call by manager", "send", from, "send.o", m.Organization, "new.o", cfg.Organization, "err", errors.New(ErrNoPrivilege))
			return nil, nil
		}
	}

	// 真实逻辑
	// 记录: role, org, tx
	rs := make([]db.Record, 0)
	c.env = env
	c.block = db.SetupBlock(env, util.AddressConvert(tx.From()), common.ToHex(tx.Hash()))

	if cfg.Op == AddOpX {
		rs2, err := c.AddMember(myDB, op, &cfg)
		if err == nil {
			rs = append(rs, rs2...)
		}
	} else if cfg.Op == DeleteOpX {
		rs2, err := c.DelMember(myDB, op, &cfg)
		if err == nil {
			rs = append(rs, rs2...)
		}
	} else if cfg.Op == SetOpX {
		rs2, err := c.SetMember(myDB, op, &cfg)
		if err == nil {
			rs = append(rs, rs2...)
		}
	}

	// r3 record tx
	txr := transaction.ConvertTransaction(c.env)
	txr.ActionName = "proof_config"
	txr.Success = true
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(txr.Hash),
		Op:   db.NewOp(op),
		Tx:   txr,
	}
	if op == db.SeqTypeDel {
		txRecord.Op = db.NewOp(db.OpDel)
	}
	if err != nil {
		txRecord.Tx.Success = false
	}
	rs = append(rs, &txRecord)
	return rs, err
}

// DelMember 增加成员
func (c *Convert) DelMember(myDB ConfigDB, blockOp int, cfg *payload) ([]db.Record, error) {
	rs := make([]db.Record, 0)
	tx := c.env.Block.Block.Txs[c.env.TxIndex]
	block := db.SetupBlock(c.env, util.AddressConvert(tx.From()), common.ToHex(tx.Hash()))

	if db.SeqTypeDel == blockOp {
		// r2 member
		mDel, err := myDB.GetMemberDel(cfg.Address, block.Height, block.Index)
		if err != nil {
			if err == db.ErrDBNotFound {
				log.Crit("get DelMember from db", "err", err, "address", cfg.Address)
				return nil, nil
			}
			return nil, errors.New(ErrDB)
		}
		dbMemDel := mDel.genRecord(cfg.Op, blockOp)

		m2 := &Member{
			Address:      mDel.Address,
			Role:         mDel.Role,
			Organization: mDel.Organization,
			Note:         mDel.Note,
			ServerName:   mDel.ServerName,
			Block:        mDel.Block,
		}
		dbMem := m2.genRecord(cfg.Op, blockOp)

		// r1 organization
		o2, err := myDB.GetOrganization(m2.Organization)
		if err != nil {
			if err == db.ErrDBNotFound {
				log.Crit("rollback: get Organization of Delete Member from db", "err", err, "address", cfg.Address)
				return nil, nil
			}
			log.Error("get Organization from db", "err", err, "organization", m2.Organization)
			return nil, errors.New(ErrDB)
		}
		o2.Count++
		dbOrg := o2.genRecord(cfg.Op, blockOp)
		rs = append(rs, dbOrg)
		rs = append(rs, dbMem)
		rs = append(rs, dbMemDel)

		myDB.SetMember(m2.Address, dbMem)
		myDB.SetMemberDel(mDel.Address, dbMemDel)
		myDB.SetOrganization(o2.Organization, dbOrg)
	}

	if db.SeqTypeAdd == blockOp {
		// r2 member
		memToDel, err := myDB.GetMember(cfg.Address)
		if err != nil {
			if err == db.ErrDBNotFound {
				log.Error("rollback: get added Member from db", "err", err, "address", cfg.Address)
				return nil, nil
			}
			log.Error("get Member from db", "err", err)
			return nil, errors.New(ErrDB)
		}
		dbMem := memToDel.genRecord(cfg.Op, blockOp)

		mDel := memberDel{
			Address:      memToDel.Address,
			Role:         memToDel.Role,
			Organization: memToDel.Organization,
			Note:         memToDel.Note,
			Block:        memToDel.Block,
			ServerName:   memToDel.ServerName,
			Del:          *c.block,
		}
		dbMemDel := mDel.genRecord(cfg.Op, blockOp)

		// r1 organization
		o2, err := myDB.GetOrganization(memToDel.Organization)
		if err != nil {
			if err == db.ErrDBNotFound {
				log.Crit("rollback: get Organization of added Member from db", "err", err, "address", cfg.Address)
				return nil, nil
			}
			log.Error("get Organization from db", "err", err, "organization", memToDel.Organization)
			return nil, errors.New(ErrDB)
		}
		o2.Count--
		dbOrg := o2.genRecord(cfg.Op, blockOp)
		rs = append(rs, dbOrg)
		rs = append(rs, dbMem)
		rs = append(rs, dbMemDel)

		myDB.SetMember(memToDel.Address, dbMem)
		myDB.SetMemberDel(mDel.Address, dbMemDel)
		myDB.SetOrganization(o2.Organization, dbOrg)
	}

	return rs, nil
}

// AddMember 增加成员
func (c *Convert) AddMember(myDB ConfigDB, blockOp int, cfg *payload) ([]db.Record, error) {
	rs := make([]db.Record, 0)
	tx := c.env.Block.Block.Txs[c.env.TxIndex]
	from := util.AddressConvert(tx.From())
	block := db.SetupBlock(c.env, from, common.ToHex(tx.Hash()))

	if db.SeqTypeDel == blockOp {
		// r2 member
		memToDel, err := myDB.GetMember(cfg.Address)
		if err != nil {
			if err == db.ErrDBNotFound {
				log.Crit("rollback: get added Member from db", "err", err, "address", cfg.Address)
				return nil, nil
			}
			log.Error("get Member from db", "err", err)
			return nil, errors.New(ErrDB)
		}
		dbMem := memToDel.genRecord(cfg.Op, db.SeqTypeDel)

		// r1 organization
		o2, err := myDB.GetOrganization(cfg.Organization)
		if err != nil {
			if err == db.ErrDBNotFound {
				log.Crit("rollback: get Organization of added Member from db", "err", err, "address", cfg.Address)
				return nil, nil
			}
			log.Error("get Organization from db", "err", err)
			return nil, errors.New(ErrDB)
		}
		o2.Count--
		dbOrg := o2.genRecord(cfg.Op, blockOp)
		rs = append(rs, dbOrg)
		rs = append(rs, dbMem)

		myDB.SetMember(memToDel.Address, dbMem)
		myDB.SetOrganization(o2.Organization, dbOrg)
	}

	if db.SeqTypeAdd == blockOp {
		// check member exists
		// 要判断地址的唯一性, 地址对交易签名作为对应组织的对应角色做的存证
		_, err := myDB.GetMember(cfg.Address)
		if err != nil && err != db.ErrDBNotFound {
			log.Error("get Member from db", "err", err)
			return nil, errors.New(ErrDB)
		}
		if err == nil {
			log.Error("member exists when add")
			return nil, nil
		}

		// r1 organization
		o2, err := myDB.GetOrganization(cfg.Organization)
		if err != nil && err != db.ErrDBNotFound {
			log.Error("get Organization from db", "err", err)
			return nil, errors.New(ErrDB)
		}
		if err != nil && err == db.ErrDBNotFound {
			o2 = createOrganization(cfg, from, block)
		}
		o2.Count++
		dbOrg := o2.genRecord(cfg.Op, blockOp)
		rs = append(rs, dbOrg)

		// r2 member
		m2 := createMember(cfg, from, block)
		dbMem := m2.genRecord(cfg.Op, blockOp)
		rs = append(rs, dbMem)

		myDB.SetMember(m2.Address, dbMem)
		myDB.SetOrganization(o2.Organization, dbOrg)
	}
	return rs, nil
}

// SetMember 设置成员信息，若之前不存在则新增
func (c *Convert) SetMember(myDB ConfigDB, blockOp int, cfg *payload) ([]db.Record, error) {
	rs := make([]db.Record, 0)
	tx := c.env.Block.Block.Txs[c.env.TxIndex]
	from := util.AddressConvert(tx.From())
	block := db.SetupBlock(c.env, from, common.ToHex(tx.Hash()))

	// 回滚新增：执行删除
	// 不支持回滚更新，若要达到该效果应将原有内容执行更新
	if db.SeqTypeDel == blockOp {
		dbOrg, err := c.DelOrganizationMember(myDB, cfg)
		if err != nil {
			return nil, err
		}
		rs = append(rs, dbOrg)
		// delete member
		m2 := createMember(cfg, from, block)
		dbMem := dbMember{
			IKey: db.NewIKey(DBX, TableX, m2.ID()),
			Op:   db.NewOp(db.OpDel),
			M:    m2,
		}
		rs = append(rs, &dbMem)
	} else if db.SeqTypeAdd == blockOp {
		op := db.OpUpdate
		// 检查存在
		_, err := myDB.GetMember(cfg.Address)
		if err != nil {
			if err == db.ErrDBNotFound {
				op = db.OpAdd
				// update organization
				dbOrg, err := c.AddOrganizationMember(myDB, cfg)
				if err != nil {
					return rs, err
				}
				rs = append(rs, dbOrg)
			} else {
				log.Error("get Member from db", "err", err)
				return nil, errors.New(ErrDB)
			}
		}

		// insert/update member
		m2 := createMember(cfg, from, block)
		dbMem := dbMember{
			IKey: db.NewIKey(DBX, TableX, m2.ID()),
			Op:   db.NewOp(op),
			M:    m2,
		}
		rs = append(rs, &dbMem)
	}
	return rs, nil
}

// AddOrganizationMember 组织中添加一个成员
func (c *Convert) AddOrganizationMember(myDB ConfigDB, cfg *payload) (db.Record, error) {
	return c.processOrganization(myDB, db.OpAdd, cfg)
}

// DelOrganizationMember 删除组织中一个成员
func (c *Convert) DelOrganizationMember(myDB ConfigDB, cfg *payload) (db.Record, error) {
	return c.processOrganization(myDB, db.OpDel, cfg)
}

// processOrganization 处理Organization，默认做更新操作
// 1. op = 1， 数量加一， 加一为1时改为新增操作
// 2. op = 2， 数量减一， 减一为0时改为删除操作
func (c *Convert) processOrganization(myDB ConfigDB, op int, cfg *payload) (db.Record, error) {
	tx := c.env.Block.Block.Txs[c.env.TxIndex]
	from := util.AddressConvert(tx.From())
	block := db.SetupBlock(c.env, from, common.ToHex(tx.Hash()))

	o2, err := myDB.GetOrganization(cfg.Organization)
	if err != nil && err != db.ErrDBNotFound {
		log.Error("get Organization from db", "err", err)
		return nil, errors.New(ErrDB)
	}
	if err != nil && err == db.ErrDBNotFound {
		o2 = createOrganization(cfg, from, block)
	}
	if op == db.OpAdd {
		o2.Count++
	} else {
		o2.Count--
	}
	if o2.Count < 0 {
		return nil, errors.New(ErrMemberCount)
	}
	dbOrg := o2.genRecord(cfg.Op, op)
	return dbOrg, nil
}

func checkInput(cfg payload) error {
	if cfg.Op != AddOpX && cfg.Op != DeleteOpX && cfg.Op != SetOpX {
		return errors.New(ErrBadParams)
	}
	if cfg.Address == "" {
		return errors.New(ErrBadParams)
	}
	if cfg.Op == AddOpX {
		if cfg.Role != ManagerX && cfg.Role != MemberX {
			return errors.New(ErrBadParams)
		}
		if cfg.Organization == "" {
			return errors.New(ErrBadParams)
		}
	}
	return nil
}

func (c *Convert) delProof(env *db.TxEnv, op int) ([]db.Record, error) {
	tx := env.Block.Block.Txs[env.TxIndex]
	from := util.AddressConvert(tx.From())
	if !isManager(from) {
		return nil, errors.New(ErrNoPrivilege)
	}

	var cfg payload
	err := json.Unmarshal(tx.Payload, &cfg)
	if err != nil {
		return nil, errors.New("bad payload format")
	}

	return nil, nil
}

// dummy
func isManager(addr string) bool {
	return true
}

// InitDB
// 创建db, 加载现有的数据, 在内存中有一个 权限相关的缓存
// 其他插件需要判断权限相关的

// Permission 其他插件需要判断权限相关的判断
// 调用这个插件的一些接口
type Permission interface {
	IsManager(address string, organization string) bool
	// 成员包括管理员, 可以用管理员直接进行操作
	IsMember(address string, organization string) bool
}

// // 从数据库取数据, cache 不能解决问题, 因为数据条目是一直增长的.
// func getMember(addr string) (*Member, error) {
// 	return nil, nil
// }
//
// func getOrganization(org string) (*Organization, error) {
// 	return nil, nil
// }

// helpful functions
func createMember(i *payload, from string, block *db.Block) *Member {
	return &Member{
		Address:        i.Address,
		Role:           i.Role,
		Organization:   i.Organization,
		Note:           i.AddressNote,
		ServerName:     i.ServerName,
		Block:          *block,
		UserDetail:     i.UserDetail,
		PersonalAuth:   i.PersonalAuth,
		EnterpriseAuth: i.EnterpriseAuth,
	}
}

func createOrganization(i *payload, from string, block *db.Block) *Organization {
	return &Organization{
		Organization: i.Organization,
		Note:         i.OrganizationNote,
		Count:        0,
		Block:        *block,
	}
}
