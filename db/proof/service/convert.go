package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"strings"

	"github.com/33cn/externaldb/db/transaction"

	"github.com/33cn/chain33/common"
	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/proof/api"
	"github.com/33cn/externaldb/db/proof/model"
	"github.com/33cn/externaldb/db/proof/proofdb"
	proofconfig "github.com/33cn/externaldb/db/proof_config"
	"github.com/33cn/externaldb/util"
)

// vars
var (
	log             = l.New("module", "db.proof")
	ErrNoPermission = errors.New("ErrNoPermission")
	ErrDisableProof = errors.New("ErrDisableProof")
	ErrParams       = errors.New("ErrParams")
	ErrDecode       = errors.New("ErrDecode")
)

//ProofConvert
type ProofConvert struct {
	Title     string
	OrgTitle  string
	Symbol    string
	Name      string
	proofDB   proofdb.IProofDB
	configDB  proofconfig.ConfigDB
	RecordGen proofdb.IProofRecord
}

// NewConvert 2注册插件需要的插件创建函数
func NewConvert(paraTitle, symbol string, name string) db.Convert {
	return &ProofConvert{OrgTitle: paraTitle, Symbol: symbol, Name: name,
		Title: db.CalcParaTitle(paraTitle)}
}

// SetDB SetDB
func (t *ProofConvert) SetDB(proofDB proofdb.IProofDB, cfgDB proofconfig.ConfigDB) error {
	t.configDB = cfgDB
	t.proofDB = proofDB
	return nil
}

//ConvertTx payload中的proof数据，目前只支持格式[{k:v},{k:v},{k:v}...]格式的jsonstr
func (t *ProofConvert) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	log.Info("convertTX", "position", t.positionID(env))
	var records []db.Record
	var err error
	tx := env.Block.Block.Txs[env.TxIndex]

	actionName := "add_proof"
	switch strings.TrimPrefix(string(tx.Execer), t.Title) {
	case api.DeleteTx, api.MainDeleteTx:
		actionName = "delete_proof"
		delproof, delerr := t.DelProof(env, op)
		err = delerr
		if err == nil {
			records = append(records, delproof...)
		} else {
			log.Error("convertTX:DelProof", "err", err)
		}
	case api.RecoverTx, api.MainRecoverTx:
		actionName = "revocer_proof"
		rproof, rerr := t.RecoverProof(env, op)
		err = rerr
		if err == nil {
			records = append(records, rproof...)
		} else {
			log.Error("convertTX:RecoverProof", "err", err)
		}
	case api.TemplateTx, api.MainTemplateTx:
		log.Debug("deal template tx", "execer", string(tx.Execer))
		addtemplate, adderr := t.AddTemplate(env, op)
		err = adderr
		if err == nil {
			log.Debug("add template tx record")
			records = append(records, addtemplate...)
		} else {
			log.Error("convertTX:AddTemplate", "err", err)
		}
	case api.AddProof, api.MainAddProof:
		log.Debug("deal add proof tx", "execer", string(tx.Execer))
		addproof, adderr := t.AddProof(env, op)
		err = adderr
		if err == nil {
			records = append(records, addproof...)
		} else {
			log.Error("convertTX:AddProof", "err", err)
		}
	default:
		log.Error("tx.Execer not match", "tx.Execer", string(tx.Execer))
	}
	// r3 record tx
	txr := transaction.ConvertTransaction(env)
	txr.ActionName = actionName
	txr.Success = true
	if err != nil {
		txr.Success = false
	}
	txRecord := transaction.TxRecord{
		IKey: transaction.NewTransactionKey(txr.Hash),
		Op:   db.NewOp(op),
		Tx:   txr,
	}
	if op == db.SeqTypeDel {
		txRecord.Op = db.NewOp(db.OpDel)
	}
	records = append(records, &txRecord)
	return records, nil
}

// ProofID proof-hash
func ProofID(hash string) string {
	return fmt.Sprintf("%s-%s", "proof", hash)
}

// TemplateID template-hash
func TemplateID(hash string) string {
	return fmt.Sprintf("%s-%s", "template", hash)
}

// AddProof 增加存证: payload中的proof数据，目前只支持格式[{k:v},{k:v},{k:v}...]格式的jsonstr
func (t *ProofConvert) AddProof(env *db.TxEnv, op int) ([]db.Record, error) {
	var records []db.Record
	var proof model.Proof

	defer func() {
		if r := recover(); r != nil {
			log.Error(" AddProof panic error", "err", r, "stack", util.GetStack())
		}
	}()

	// 签名地址是否有权限,没有直接返回
	height := env.Block.Block.Height
	tx := env.Block.Block.Txs[env.TxIndex]
	fromAddr := util.AddressConvert(tx.From())

	if !t.configDB.IsHaveProofPermission(fromAddr) {
		log.Error("AddProof:IsHaveProofPermission", "err", ErrNoPermission)
		return nil, ErrNoPermission
	}

	org, err := t.configDB.GetOrganizationName(fromAddr)
	if err != nil {
		log.Error("AddProof:GetOrganizationName", "err", err)
		return nil, err
	}

	mapinfo, err := parseProof(t.proofDB, t.Title, height, tx.Payload)
	if err != nil {
		log.Error("AddProof:parseProof", "err", err)
		return nil, err
	}

	sourceHash := make([]string, 0)
	if value, ok := mapinfo["source_hash"]; ok {
		sourceHashInfo, ok1 := value.([]interface{})
		if ok1 {
			for _, v := range sourceHashInfo {
				if hashVal, ok2 := v.(string); ok2 {
					sourceHash = append(sourceHash, hashVal)
				} else {
					log.Error("AddProof source_hash has error value", "height", height, "tx", tx, "source_hash", value)
				}
			}
		}
	}

	//填写展开后的信息
	mapinfo["proof_sender"] = fromAddr
	mapinfo["proof_organization"] = org
	mapinfo["proof_tx_hash"] = common.ToHex(tx.Hash())
	mapinfo["proof_height_index"] = db.HeightIndex(height, env.TxIndex)

	mapinfo["proof_block_hash"] = env.BlockHash
	mapinfo["proof_height"] = height
	mapinfo["proof_block_time"] = env.Block.Block.BlockTime

	mapinfo["proof_deleted"] = "" //此字段在后面删除此存证时填写删除此存证的交易hash值
	mapinfo["proof_deleted_note"] = ""
	mapinfo["proof_deleted_flag"] = false

	mapinfo["proof_id"] = ProofID(common.ToHex(tx.Hash()))
	if mapinfo["basehash"] == nil || mapinfo["basehash"] == "" {
		mapinfo["basehash"] = "null"
	}
	mapinfo["source_hash"] = sourceHash

	//处理更新存证
	//若有新增存证，需先将存证记录做一份储存，再在原先存证记录上进行修改
	if value, ok := mapinfo["update_hash"]; ok && mapinfo["update_hash"] != "null" && mapinfo["update_hash"] != "" {
		updateHash := value.(string)
		prevProof, err := t.proofDB.GetProof(ProofID(updateHash))
		if err != nil {
			log.Error("Get previous proof failed", "err", err)
			return nil, err
		}

		mapinfo["basehash"] = prevProof.Proof["basehash"]
		mapinfo["prehash"] = prevProof.Proof["prehash"]
		if mapinfo["source_hash"] == nil {
			mapinfo["source_hash"] = prevProof.Proof["source_hash"]
		}

		if v, ok := prevProof.Proof["update_version"]; ok {
			updateVersion := int(v.(float64))
			switch op {
			case db.SeqTypeAdd:
				mapinfo["update_version"] = updateVersion + 1

				//备份update proof
				prevProof.Proof["update_hash"] = updateHash
				prevProof.Proof["proof_id"] = ProofID(prevProof.Proof["proof_tx_hash"].(string))
				ProofUpdateRecord := t.RecordGen.ProofUpdateRecord(prevProof, prevProof.Proof["proof_id"].(string), db.OpAdd)
				records = append(records, ProofUpdateRecord)
			case db.SeqTypeDel:
				//获取上一个版本的数据
				data, err := t.dealRollback(mapinfo, updateVersion-1, updateHash)
				if err != nil {
					return nil, err
				}

				//在备份中将该版本删除
				ProofUpdateRecord := t.RecordGen.ProofUpdateRecord(data, data.Proof["proof_id"].(string), db.OpDel)
				records = append(records, ProofUpdateRecord)
			}
		} else {
			//老版本，没有设置有关更新存证的字段
			mapinfo["update_version"] = 1

			//备份update proof
			prevProof.Proof["update_hash"] = updateHash
			prevProof.Proof["proof_id"] = ProofID(prevProof.Proof["proof_tx_hash"].(string))
			ProofUpdateRecord := t.RecordGen.ProofUpdateRecord(prevProof, prevProof.Proof["proof_id"].(string), db.OpAdd)
			records = append(records, ProofUpdateRecord)
		}

		//更新proof
		mapinfo["update_hash"] = "null"
		mapinfo["proof_id"] = ProofID(updateHash)
		proof.Proof = mapinfo
		m := t.RecordGen.Proof(&proof, proof.Proof["proof_id"].(string), db.OpUpdate)
		records = append(records, m)
	} else {
		mapinfo["update_version"] = 0
		mapinfo["update_hash"] = "null"
		mapinfo["proof_id"] = ProofID(common.ToHex(tx.Hash()))

		proof.Proof = mapinfo
		ProofRecord := t.RecordGen.Proof(&proof, ProofID(common.ToHex(tx.Hash())), op)
		records = append(records, ProofRecord)
	}

	if err := t.addUserInfo(fromAddr, mapinfo); err != nil {
		log.Error("AddProof:addUserInfo", "err", err)
		return nil, err
	}

	return records, nil
}

func (t *ProofConvert) dealRollback(mapinfo map[string]interface{}, updateVersion int, updateHash string) (*model.Proof, error) {
	p, err := t.proofDB.GetProofUpdateRecord(ProofID(updateHash), updateVersion)
	if err != nil {
		log.Error("GetUpdateProof failed", "err", err)
		return nil, err
	}

	mapinfo = nil
	mapinfo = p.Proof

	return p, nil
}

func (t *ProofConvert) addUserInfo(addr string, mapinfo map[string]interface{}) error {
	mem, err := t.configDB.GetMember(addr)
	if err != nil {
		if err == db.ErrDBNotFound {
			log.Info("sender info not find", "sender", addr)
			return nil
		}
		return err
	}
	if mem.UserDetail != nil {
		mapinfo["user_name"] = mem.UserName
		mapinfo["user_icon"] = mem.UserIcon
		mapinfo["user_phone"] = mem.Phone
		mapinfo["user_email"] = mem.Email
		mapinfo["user_auth_type"] = mem.AuthType
	}
	if mem.PersonalAuth != nil {
		mapinfo["user_real_name"] = mem.RealName
	}
	if mem.EnterpriseAuth != nil {
		mapinfo["user_enterprise_name"] = mem.EnterpriseName
	}
	return nil
}

// DelProof 删除存证
func (t *ProofConvert) DelProof(env *db.TxEnv, op int) ([]db.Record, error) {
	tx := env.Block.Block.Txs[env.TxIndex]
	from := util.AddressConvert(tx.From())

	var d api.DeleteProof
	err := json.Unmarshal(tx.Payload, &d)
	if err != nil {
		return nil, errors.New(proofconfig.ErrBadParams)
	}

	p, err := t.proofDB.GetProof(ProofID(d.ID))
	if err != nil {
		return nil, err
	}

	if !t.configDB.IsHaveDelProofPermission(from, p.Proof["proof_organization"].(string), p.Proof["proof_sender"].(string)) {
		return nil, errors.New(proofconfig.ErrNoPrivilege)
	}

	if baseHash, ok := p.Proof["basehash"]; ok && baseHash.(string) != "null" {
		return nil, errors.New(api.ErrNotBaseProof)
	}

	if deletedFlag, ok := p.Proof["proof_deleted_flag"]; ok && deletedFlag.(bool) {
		return nil, errors.New(api.ErrProofDeleted)
	}

	proofs, err := t.proofDB.ListProof(d.ID)
	if err != nil {
		return nil, err
	}

	hash := common.ToHex(tx.Hash())
	del := newDeleter(env, hash, &d)
	records, err := del.del(p, proofs, op, t.RecordGen)

	log.Info("convertTX delete", "proof", p.Proof)
	return records, err
}

// RecoverProof 恢复存证
func (t *ProofConvert) RecoverProof(env *db.TxEnv, op int) ([]db.Record, error) {
	tx := env.Block.Block.Txs[env.TxIndex]
	from := util.AddressConvert(tx.From())

	var d api.RecoverProof
	err := json.Unmarshal(tx.Payload, &d)
	if err != nil {
		return nil, errors.New(proofconfig.ErrBadParams)
	}

	p, err := t.proofDB.GetProof(ProofID(d.ID))
	if err != nil {
		return nil, err
	}

	if !t.configDB.IsHaveDelProofPermission(from, p.Proof["proof_organization"].(string), p.Proof["proof_sender"].(string)) {
		return nil, errors.New(proofconfig.ErrNoPrivilege)
	}

	if baseHash, ok := p.Proof["basehash"]; ok && baseHash.(string) != "null" {
		return nil, errors.New(api.ErrNotBaseProof)
	}

	if deletedFlag, ok := p.Proof["proof_deleted_flag"]; ok && !deletedFlag.(bool) {
		return nil, errors.New(api.ErrProofNotDeleted)
	}

	proofs, err := t.proofDB.ListProof(d.ID)
	if err != nil {
		return nil, err
	}

	hash := common.ToHex(tx.Hash())
	rcv := newRecover(env, hash, &d)
	records, err := rcv.recover(p, proofs, op, t.RecordGen)

	log.Info("convertTX recover", "proof", p.Proof)
	return records, err
}

func (t *ProofConvert) positionID(env *db.TxEnv) string {
	return util.PositionID(t.Name, env.Block.Block.Height, env.TxIndex)
}

// AddTemplate 保存模板
func (t *ProofConvert) AddTemplate(env *db.TxEnv, op int) ([]db.Record, error) {
	//错误拦截 panic异常错误
	defer func() {
		if r := recover(); r != nil {
			log.Error(" AddTemplate panic error", "err", r)
		}
	}()

	// 签名地址是否有权限,没有直接返回
	height := env.Block.Block.Height
	tx := env.Block.Block.Txs[env.TxIndex]
	fromAddr := util.AddressConvert(tx.From())
	//判断是否有权限
	if !t.configDB.IsHaveProofPermission(fromAddr) {
		log.Error("AddTemplate:IsHaveProofPermission", "err", ErrNoPermission)
		return nil, ErrNoPermission
	}
	//判断组织名称
	org, err := t.configDB.GetOrganizationName(fromAddr)
	if err != nil {
		log.Error("AddTemplate:GetOrganizationName", "err", err)
		return nil, err
	}
	//解析模板
	mapinfo, err := parseTemplate(t.Title, height, tx.Payload)
	if err != nil {
		log.Error("AddTemplate:parseProof", "err", err)
		return nil, err
	}

	//填写展开后的信息
	mapinfo["template_sender"] = fromAddr
	mapinfo["template_organization"] = org
	mapinfo["template_tx_hash"] = common.ToHex(tx.Hash())
	mapinfo["template_height_index"] = db.HeightIndex(height, env.TxIndex)

	mapinfo["template_block_hash"] = env.BlockHash
	mapinfo["template_height"] = height
	mapinfo["template_block_time"] = env.Block.Block.BlockTime

	mapinfo["template_deleted"] = "" //此字段在后面删除此存证时填写删除此存证的交易hash值
	mapinfo["template_deleted_note"] = ""
	mapinfo["template_deleted_flag"] = false

	mapinfo["template_id"] = TemplateID(common.ToHex(tx.Hash()))

	log.Debug("AddTemplate mapinfo", "mapinfo", mapinfo)
	var records []db.Record
	var template model.Template
	template.Template = mapinfo
	ProofRecord := t.RecordGen.Template(&template, TemplateID(common.ToHex(tx.Hash())), op)
	log.Debug("AddTemplate ProofRecord", "ProofRecord", ProofRecord)

	records = append(records, ProofRecord)
	return records, nil
}
