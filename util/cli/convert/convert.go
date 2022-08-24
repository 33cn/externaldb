package convert

import (
	"strings"
	"time"

	l "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/converts"
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/db/account"
	"github.com/33cn/externaldb/db/address"
	"github.com/33cn/externaldb/db/block"
	"github.com/33cn/externaldb/db/blockinfo"
	"github.com/33cn/externaldb/db/contract"
	"github.com/33cn/externaldb/db/contractverify"
	"github.com/33cn/externaldb/db/evm"
	"github.com/33cn/externaldb/db/pos33"
	proofconfig "github.com/33cn/externaldb/db/proof_config"
	"github.com/33cn/externaldb/escli"
	"github.com/33cn/externaldb/proto"
	"github.com/33cn/externaldb/stat"
	"github.com/33cn/externaldb/store"
	"github.com/33cn/externaldb/store/syncseq"
	"github.com/33cn/externaldb/util"
)

const (
	paraPrefix = "user.p."
)

var (
	log         = l.New("module", "converts")
	EsWrite     escli.ESClient
	SeqNumStore store.SeqNumStore
	SeqStore    store.SeqStore
)

func InitDB(cfg *proto.ConfigNew) {
	InitWriteDB(cfg.ConvertEs, cfg.EsVersion)
	InitSeqStore(cfg)
	InitSeqNum(cfg)
	db.SetVersion(cfg.EsVersion)
	db.SetAddrID(cfg.Convert.AddressDriver)
	InitContract(cfg)
	InitConfig(cfg)
	if cfg.Chain.Symbol == "ycc" {
		pos33.SetCoinsReward(cfg.Chain.CoinPrecision, cfg.Chain.PerTicketReward)
		pos33.SetHost(cfg.Chain.Host)
	}
}

// InitWriteDB 设置存储数据的配置
func InitWriteDB(writeDB *proto.ESDB, version int32) {
	var err error
	EsWrite, err = escli.NewESLongConnect(writeDB.Host, writeDB.Prefix, version, writeDB.User, writeDB.Pwd)
	if err != nil {
		panic(err)
	}
	err = address.Init(EsWrite)
	if err != nil {
		panic(err)
	}
	err = initIndex(EsWrite)
	if err != nil {
		panic(err)
	}
}

func InitConfig(cfg *proto.ConfigNew) {
	proofconfig.InitOpenAccessControl(cfg.Convert.OpenAccessControl)
}

func InitContract(cfg *proto.ConfigNew) {
	contract.InitDetector(cfg)
	if err := contractverify.InitESDB(EsWrite); err != nil {
		panic(err)
	}
	cfydb := contractverify.NewEsDB(EsWrite)
	for _, c := range cfg.Contracts {
		cfy := contractverify.ContractVerify{
			ContractBin:  c.Bin,
			ContractAbi:  c.Abi,
			ContractType: c.Type,
		}
		_, err := cfydb.Get(cfy.GetContractBinHash())
		switch err {
		case nil:
		case db.ErrDBNotFound:
			err := cfydb.Set(contractverify.NewRecord(db.OpAdd, &cfy))
			if err != nil {
				log.Error("init contract, Set failed", "err", err, "bin-hash", cfy.GetContractBinHash())
			} else {
				log.Info("init contract success", "bin-hash", cfy.GetContractBinHash())
			}

		default:
			log.Error("init contract, Get failed", "err", err, "bin-hash", cfy.GetContractBinHash())
		}
	}
}

func InitSeqStore(cfg *proto.ConfigNew) {
	var err error
	SeqNumStore, SeqStore, err = syncseq.NewGetSeq(cfg)
	if SeqNumStore == nil && err != nil {
		log.Info("create seqNumStore failed", "err", err)
		panic(err)
	} else if SeqStore == nil && err != nil {
		log.Info("create seqStore failed", "err", err)
		panic(err)
	}
}

func InitSeqNum(cfg *proto.ConfigNew) {
	err := util.InitLastSyncSeqCache(EsWrite, cfg.Convert.AppName, cfg.Sync.StartSeq)
	if err != nil {
		log.Error("初始化 区块解析进度参数 last_seq 失败，请确保ES服务正常且 配置文件参数sync.startSeq 参数大于或等于0")
		panic(err)
	}
}

// create index for last seq, used by all module
func initIndex(cli db.DBCreator) error {
	return util.InitIndex(cli, db.LastSeqDB, db.LastSeqDB, db.SeqMapping)
}

type App struct {
	title            string
	defaultExec      string
	dealOtherChainTx bool
	savaBlockInfo    bool
	execs            map[string]db.ExecConvert
	stats            map[string]stat.Stat
}

func NewApp(cfg *proto.ConfigNew) *App {
	a := &App{title: cfg.Chain.Title, savaBlockInfo: cfg.Convert.SaveBlockInfo, defaultExec: cfg.Convert.DefaultExec, dealOtherChainTx: cfg.Convert.DealOtherChain}
	a.execs = make(map[string]db.ExecConvert)
	for _, exec := range cfg.Convert.Data {
		newConvert, _ := converts.Load(exec.Exec)
		if newConvert == nil {
			panic("exec convert not exist: " + exec.Exec)
		}
		log.Info("load exec: " + exec.Exec + " success")
		convert := newConvert(cfg.Chain.Title, cfg.Chain.Symbol, exec.Generate)
		err := convert.InitDB(EsWrite)
		if err != nil {
			panic(err)
		}
		if c, ok := convert.(db.NeedWrapDB); ok {
			_ = c.SetDB(EsWrite)
		}
		a.execs[exec.Exec] = convert
	}

	// 先放在这， 看是否有更好的位置，来初始化 account
	account.Init(cfg.Chain.Title, cfg.Convert.ExecAddresses)

	// 统计插件
	a.stats = make(map[string]stat.Stat)
	for _, statName := range cfg.Convert.Stat {
		newStat, _ := converts.LoadStat(statName.Stat)
		if newStat == nil {
			panic("stat convert not exist: " + statName.Stat)
		}
		if cfg.Chain.Title == "bityuan" {
			a.stats[statName.Stat] = newStat(cfg.Chain.Title, cfg.Chain.Symbol, -1, -1)
		} else {
			a.stats[statName.Stat] = newStat(cfg.Chain.Title, cfg.Chain.Symbol, cfg.Chain.OtherChainGenesis, cfg.Chain.PerBlockCoin)
		}
		err := a.stats[statName.Stat].InitDB(EsWrite)
		if err != nil {
			panic(err)
		}
	}

	// 保存区块基本信息
	if a.savaBlockInfo {
		err := blockinfo.InitDB(EsWrite)
		if err != nil {
			panic(err)
		}
	}

	return a
}

// v1: 原来配置para是平行链名字, 不是平行链, 配置空
// v2: 后来配置para为链的名字 (非平行链不为空, 如bityuan)
func isChainTx(para, exec string) (bool, string) {
	// 非平行链, 将para设置为空, 用原来的逻辑
	if !strings.HasPrefix(para, paraPrefix) {
		para = ""
	}

	if para == "" {
		has := strings.HasPrefix(exec, paraPrefix)
		if has {
			return false, ""
		}
		return true, exec
	}

	if len(exec) <= len(para) {
		return false, ""
	}
	is := strings.HasPrefix(exec, para)
	if !is {
		return false, ""
	}
	return is, exec[len(para):]
}

func (a *App) ConvertTx(env *db.TxEnv, op int) ([]db.Record, error) {
	log.Debug("App.convert", "h", env.Block.Block.Height, "i", env.TxIndex)
	var convertName string

	tx := env.Block.Block.Txs[env.TxIndex]
	is, exec := isChainTx(a.title, string(tx.Execer))
	// 根据配置, 不处理它链的交易
	if !is && !a.dealOtherChainTx {
		return nil, nil
	}
	if !is {
		// 非本链交易只记录交易列表
		convertName = a.defaultExec
	} else {
		convertName = exec
	}

	execConvert, ok := a.execs[convertName]
	if !ok {
		// 不支持的合约生成交易列表
		convertName = a.defaultExec
	}

	log.Debug("App.convert", "convertName", convertName)
	execConvert, ok = a.execs[convertName]
	if !ok {
		return nil, nil
	}

	return execConvert.ConvertTx(env, op)
}

func (a *App) RecoverStats(client escli.ESClient, lastSeq int64) error {
	for _, st := range a.stats {
		err := st.Recover(client, lastSeq)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) ConvertBlock(blockSeq *block.Seq, detail *types.BlockDetail) ([]db.Record, error) {
	var records []db.Record
	var err error
	var startTime = time.Now()
	txs := detail.Block.GetTxs()
	for i := 0; i < len(txs); i++ {
		env := db.TxEnv{
			Block:     detail,
			TxIndex:   int64(i),
			BlockHash: blockSeq.Hash,
		}
		rs, err := a.ConvertTx(&env, blockSeq.Type)
		if err != nil {
			return nil, err
		}
		records = append(records, rs...)
	}

	// 统计地址
	addrs := make(map[string]*address.Address)
	for i := 0; i < len(txs); i++ {
		to := txs[i].To
		from := txs[i].From()
		toAddr, err := address.Manager.AddTxCount(blockSeq.Type, to)
		if err != nil {
			log.Error("address.Manager.AddTxCount", "err", err, "to", to)
		} else {
			addrs[to] = toAddr
		}
		if from != to {
			fromAddr, err := address.Manager.AddTxCount(blockSeq.Type, from)
			if err != nil {
				log.Error("address.Manager.AddTxCount", "err", err, "from", from)
			} else {
				addrs[from] = fromAddr
			}
		}
	}

	// 处理合约地址
	for _, record := range records {
		if record.Index() == contract.TableName {
			if r, ok := record.(db.SourceAbleRecord); ok {
				if t, ok := r.Source().(*contract.Contract); ok {
					addrs[t.Address] = &address.Address{Address: t.Address, TxCount: t.TxCount, AddrType: address.AccountContract}
				}
			}
		}
		// 合约内部转账记录地址也算上
		if record.Index() == evm.EVMTransferX {
			if r, ok := record.(db.SourceAbleRecord); ok {
				if t, ok := r.Source().(*evm.Transfer); ok {
					toAddr, err := address.Manager.AddEvmTransferCount(blockSeq.Type, t.To)
					if err != nil {
						log.Error("address.Manager.AddTxCount", "err", err, "from", t.From)
					} else {
						addrs[t.To] = toAddr
					}
					if t.From != t.To {
						fromAddr, err := address.Manager.AddEvmTransferCount(blockSeq.Type, t.From)
						if err != nil {
							log.Error("address.Manager.AddTxCount", "err", err, "from", t.From)
						} else {
							addrs[t.From] = fromAddr
						}
					}
				}
			}
		}
	}
	// 添加地址信息
	for _, v := range addrs {
		records = append(records, address.NewRecord(db.OpAdd, v))
	}

	// 统计插件, config enable
	for _, st := range a.stats {
		rs, err := st.Stat(detail, blockSeq.Type)
		if err != nil {
			return nil, err
		}
		records = append(records, rs...)
	}

	// 保存区块基本信息
	if a.savaBlockInfo {
		r, err := blockinfo.SaveBlock(detail.Block, blockSeq.Hash, blockSeq.Type)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	log.Info("ConvertBlock", "cost", types.Since(startTime))
	return records, err
}

type Service struct {
	mod *util.ModuleConvert
}

func NewConvertService(cfg *proto.ConfigNew) *Service {
	mod := &util.ModuleConvert{
		Name:        cfg.Convert.AppName,
		SeqNumStore: SeqNumStore,
		SeqStore:    SeqStore,
		WriteDB:     EsWrite,
		StartSeq:    cfg.Convert.StartSeq,
		ForceSeq:    false,
		AppConvert:  NewApp(cfg),
	}

	return &Service{
		mod: mod,
	}
}

// Start 启动convert服务
func (s *Service) Start() {
	for {
		if util.ConvertServerStatus.Closed() {
			return
		}
		s.mod.BlockProc()
		time.Sleep(100 * time.Millisecond)
	}
}
