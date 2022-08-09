package nft

import (
	"encoding/json"
	"errors"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/evm"
	ndb "github.com/33cn/externaldb/db/evm/nft/db"
	"github.com/33cn/externaldb/util"
)

var (
	log = l.New("module", "db.evm.nft")
)

func init() {
	evm.RegisterFuncConvert("AddNewGoodsAssignGoodsID", addNewGoodsAssignGoodsID)
	evm.RegisterFuncConvert("batchAddNewGoodsAssignGoodsID", batchAddNewGoodsAssignGoodsID)
	evm.RegisterFuncConvert("BatchTransferWithEvent", batchTransferWithEvent)
	evm.RegisterFuncConvert("BatchMintWithEvent", batchMintWithEvent)
	evm.RegisterFuncConvert("TransferWithVerificationCode", transferWithVerificationCode)
	evm.RegisterInitDB(initDB)
}

func initDB(cli db.DBCreator) error {
	err := util.InitIndex(cli, ndb.TokenX, ndb.TokenX, ndb.TokenMapping)
	if err != nil {
		return err
	}
	err = util.InitIndex(cli, ndb.TransferX, ndb.TransferX, ndb.TransferMapping)
	if err != nil {
		return err
	}
	err = util.InitIndex(cli, ndb.AccountX, ndb.AccountX, ndb.AccountMapping)
	if err != nil {
		return err
	}

	return nil
}

func addNewGoodsAssignGoodsID(c *evm.Convert, op int, data evm.EVM) ([]db.Record, error) {
	event, err := data.GetEvent("addNewGoodsResult")
	if err != nil {
		log.Error("addNewGoodsAssignGoodsID data.GetEvent(\"addNewGoodsResult\")", "err", err)
		return nil, err
	}
	res, err := ndb.NewAddNewGoodsResult(event)
	if err != nil {
		log.Error("addNewGoodsAssignGoodsID NewAddNewGoodsResult", "err", err)
		return nil, err
	}
	token := res.GetToken(data)
	records := make([]db.Record, 0)
	records = append(records, ndb.NewRecordToken(token, op))

	// Account
	recordAccounts, err := parseToAccountRecord(c, data, op, []string{token.LabelID})
	if err != nil {
		log.Error("addNewGoodsAssignGoodsID parseToAccountRecord", "err", err)
		return records, nil
	}
	records = append(records, recordAccounts...)

	return records, nil
}

func batchAddNewGoodsAssignGoodsID(c *evm.Convert, op int, data evm.EVM) ([]db.Record, error) {
	event, err := data.GetEvent("batchAddNewGoodsResult")
	if err != nil {
		log.Error("batchAddNewGoodsAssignGoodsID data.GetEvent(\"batchAddNewGoodsResult\")", "err", err)
		return nil, err
	}
	res, err := ndb.NewBatchAddNewGoodsResult(event)
	if err != nil {
		log.Error("addNewGoodsAssignGoodsID NewAddNewGoodsResult", "err", err)
		return nil, err
	}
	records := make([]db.Record, 0)
	for i := range res.Results {
		token := res.Results[i].GetToken(data)
		records = append(records, ndb.NewRecordToken(token, op))
	}

	// Account
	recordAccounts, err := parseToAccountRecord(c, data, op, res.LabelID)
	if err != nil {
		log.Error("addNewGoodsAssignGoodsID parseToAccountRecord", "err", err)
		return records, nil
	}
	records = append(records, recordAccounts...)

	return records, nil
}

func batchTransferWithEvent(c *evm.Convert, op int, data evm.EVM) ([]db.Record, error) {
	// transfer
	event, err := data.GetEvent("batchTransferResult")
	if err != nil {
		log.Error("batchTransferWithEvent data.GetEvent(\"batchTransferResult\")", "err", err)
		return nil, err
	}
	records := make([]db.Record, 0)
	transfer, err := ndb.NewTransfer(event, data, c.ConfigDB)
	if err != nil {
		log.Error("batchTransferWithEvent NewTransfer(data)", "err", err)
		return nil, err
	}
	records = append(records, ndb.NewRecordTransfer(transfer, op))

	// account
	recordAccounts, err := parseToAccountRecord(c, data, op, nil)
	if err != nil {
		log.Error("batchTransferWithEvent parseToAccountRecord", "err", err)
		return records, nil
	}
	records = append(records, recordAccounts...)

	return records, nil
}

// parseToAccountRecord 解析账户信息。默认解析event "balanceResult"，
func parseToAccountRecord(c *evm.Convert, info evm.EVM, op int, labelIDs []string) ([]db.Record, error) {
	var res = make([]ndb.AccountBalance, 0)
	if m, err := info.GetEvent("balanceResult"); err == nil {
		infoByte, _ := json.Marshal(m["balanceList"])
		if err := json.Unmarshal(infoByte, &res); err != nil {
			log.Error("parseToAccountRecord json.Unmarshal(infoByte, &res)", "err", err)
			return nil, err
		}
	} else {
		log.Error("parseToAccountRecord data.GetEvent(\"balanceResult\")", "err", err)
		return nil, err
	}

	records := make([]db.Record, 0)
	for i, re := range res {
		acc := re.GetAccount(info)
		// 获取labelID，当为交易类型时，需从es中获取相应的labelID
		if labelIDs != nil && i < len(labelIDs) && labelIDs[i] != "" {
			acc.LabelID = labelIDs[i]
		} else {
			token, err := c.NftTokenDb.GetNft(acc.ContractAddr, acc.GoodsID)
			if err != nil {
				log.Error("get token failed", "err", err)
			} else {
				acc.LabelID = token.LabelID
			}
		}
		records = append(records, ndb.NewRecordAccount(acc, op))
	}

	return records, nil
}

func batchMintWithEvent(c *evm.Convert, op int, data evm.EVM) ([]db.Record, error) {
	// Tokens
	BatchMintResult, err := data.GetEvent("BatchMintResult")
	if err != nil {
		log.Error("batchMintWithEvent data.GetEvent(\"BatchMintResult\")", err, "err")
		return nil, err
	}
	records := make([]db.Record, 0)
	labelIds := make([]string, 0)
	if goods, ok := BatchMintResult["mintGoodsList"]; ok {
		res, err := ndb.NewMintGoodsResults(goods)
		if err != nil {
			log.Error("batchMintWithEvent NewAddNewGoodsResult", "err", err, "good", goods)
			return nil, err
		}
		for _, g := range res {
			labelIds = append(labelIds, g.LabelID)
			token := g.GetToken(data)
			if token.Publisher == "" && token.PublisherAddress != "" {
				m, err := c.ConfigDB.GetMember(token.PublisherAddress)
				if err == nil {
					token.Publisher = m.GetUserName()
				}
			}
			records = append(records, ndb.NewRecordToken(token, op))
		}
	} else {
		log.Error("batchMintWithEvent BatchMintResult[\"mintGoodsList\"].([]map[string]interface{}) failed")
	}

	// Accounts
	recordAccounts, err := parseToAccountRecord(c, data, op, labelIds)
	if err != nil {
		log.Error("batchMintWithEvent parseToAccountRecord", "err", err)
		return records, nil
	}
	records = append(records, recordAccounts...)

	return records, nil
}

func transferWithVerificationCode(c *evm.Convert, op int, data evm.EVM) ([]db.Record, error) {
	records := make([]db.Record, 0)
	// transfer
	tr, err := batchTransferWithEvent(c, op, data)
	if err != nil {
		log.Error("transferWithVerificationCode batchTransferWithEvent", "err", err)
	}
	records = append(records, tr...)

	// GoodsUsedResult
	event, err := data.GetEvent("GoodsUsedResult")
	if err != nil {
		log.Error("transferWithVerificationCode data.GetEvent(\"GoodsUsedResult\")", "err", err)
		return nil, err
	}
	r, err := ndb.NewGoodsUsedResult(event)
	if err != nil {
		log.Error("transferWithVerificationCode NewGoodsUsedResult(event)", "err", err)
		return nil, err
	}
	rd, err := GetToken(r, c, data, op)
	if err != nil {
		log.Error("transferWithVerificationCode NewGoodsUsedResult(event)", "err", err)
		return nil, err
	}
	records = append(records, ndb.NewRecordToken(rd, db.OpUpdate))
	return records, nil
}

var ErrContractAddrNotFound = errors.New("contract_addr not found")

func GetToken(r *ndb.GoodsUsedResult, c *evm.Convert, info map[string]interface{}, op int) (*ndb.Token, error) {
	var contractAddr string
	if v, ok := info["contract_addr"].(string); ok {
		contractAddr = v
	} else {
		log.Error("GoodsUsedResult.GetToken, info[\"contract_addr\"].(string) failed")
		return nil, ErrContractAddrNotFound
	}
	t, err := c.NftTokenDb.GetNft(contractAddr, r.GoodsID)
	if err != nil {
		log.Error("GoodsUsedResult.GetToken,okenCli.GetToken", "err", err)
		return nil, err
	}
	t.GoodsID = r.GoodsID
	if op == db.SeqTypeDel {
		t.IsUsed = false
		t.UseTime = 0
		t.UserName = ""
		t.UserAddr = ""
		t.UseCode = ""
		t.UserRealName = ""
		return t, nil
	}
	t.IsUsed = r.IsUsed
	t.UseCode = r.Code
	t.UserAddr = r.To
	if v, ok := info["evm_block_time"].(int64); ok {
		t.UseTime = v
	}
	if m, err := c.ConfigDB.GetMember(r.To); err == nil {
		t.UserRealName = m.GetUserRealName()
		t.UserName = m.GetUserName()
	} else {
		log.Error("GoodsUsedResult.GetToken, ConfigCli.GetMember(r.To)", "err", err, "addr", r.To)
	}
	return t, nil
}
