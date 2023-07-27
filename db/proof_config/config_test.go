package proofconfig

import (
	"encoding/json"
	"math/rand"
	"regexp"
	"testing"

	"github.com/33cn/externaldb/escli/querypara"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/externaldb/db"
	util2 "github.com/33cn/externaldb/util"
	"github.com/stretchr/testify/assert"
)

var (
	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	PrivKeyC = "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" // 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
	PrivKeyD = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" // 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
	Nodes    = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
		[]byte("1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"),
		[]byte("1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"),
	}
)

// 用于构建测试用的交易
func Test_CreateTx(t *testing.T) {
	exec := "user.p.testproof.config"
	superAddress := "1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY"
	superPrivkey := "0x9c451df9e5cb05b88b28729aeaaeb3169a2414097401fcb4c79c1971df734588"

	_ = superAddress
	a1 := `{"op" : "add", "organization": "system", "role": "manager", "address" : "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4",  "address_note": "s1-a1"}`

	tx1 := createTx(exec, []byte(a1), superPrivkey)
	txBytes := common.ToHex(types.Encode(tx1))
	t.Log("tx = ", txBytes)
	// 0x0a17757365722e702e7465737470726f6f662e636f6e6669671287017b226f7022203a2022616464222c20226f7267616e697a6174696f6e223a202273797374656d222c2022726f6c65223a20226d616e61676572222c20226164647265737322203a2022314b534264313748375a4b3869543337614a7a744642323258477773505464774534222c202022616464726573735f6e6f7465223a202273312d6131227d1a6e08011221023c712ab96682fdc1d7a9101f860bf07750b719221f01a56e0e0d61547650f2551a4730450221009b48d5a5b5f72a10ab0a24ca3a13728038d9b435c3946739f39a524a25fb9711022029ad606d1656733f02b505d3eae3e881167cfef54748a55e7dc5c418ddc2765c2080c2d72f30ad9bf4f390c7c5dc373a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47
	// 0x4e06aacbb7e2b5b897d6ab2e634ab88fc631beace92d25ec2080107dde1c282b

	a2 := `{"op" : "add", "organization": "org1", "role": "manager", "address" : "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR", "organization_note": "o1", "address_note": "o1-a1"}`
	tx2 := createTx(exec, []byte(a2), superPrivkey)
	txBytes2 := common.ToHex(types.Encode(tx2))
	t.Log("tx = ", txBytes2)
	// 0x0a17757365722e702e7465737470726f6f662e636f6e666967129f017b226f7022203a2022616464222c20226f7267616e697a6174696f6e223a20226f726731222c2022726f6c65223a20226d616e61676572222c20226164647265737322203a2022314a524e6a64457170344c4a356671796355426d396179434b536565736b674d4b52222c20226f7267616e697a6174696f6e5f6e6f7465223a20226f31222c2022616464726573735f6e6f7465223a20226f312d6131227d1a6e08011221023c712ab96682fdc1d7a9101f860bf07750b719221f01a56e0e0d61547650f2551a473045022100b065a44bed226cae63f0aabc98b68fa620abda9206e11f88d513a487f2c8cbbc02206425c899f990fdd0cb81d26cc1de66e00311ba24b216cb9210d7680fccd1df2a2080c2d72f30e8c4b8a0908db5a91c3a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47
	// 0xe580411e8d388aed2c06e181c69f7e82658d608ab35edba30160cc0ca2d74f7b

	a3 := `{"op" : "add", "organization": "org1", "role": "member", "address" : "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k", "organization_note": "o1", "address_note": "o1-a1"}`
	tx3 := createTx(exec, []byte(a3), superPrivkey)
	txBytes3 := common.ToHex(types.Encode(tx3))
	t.Log("tx = ", txBytes3)
	// 0x0a17757365722e702e7465737470726f6f662e636f6e666967129e017b226f7022203a2022616464222c20226f7267616e697a6174696f6e223a20226f726731222c2022726f6c65223a20226d656d626572222c20226164647265737322203a2022314e4c4850456362545757787855336447555a426861796a7243484433707358376b222c20226f7267616e697a6174696f6e5f6e6f7465223a20226f31222c2022616464726573735f6e6f7465223a20226f312d6131227d1a6e08011221023c712ab96682fdc1d7a9101f860bf07750b719221f01a56e0e0d61547650f2551a473045022100883fcea1c638aba3098637a6f720a26f169190ffb0393dc9a0b63621ef59b7ee02204f887dc06424dd4065b7dff4a5b86b05d0fa3a54cf9f648a1e048a4d327b52752080c2d72f3090dcd6a192f896c66d3a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47
	// 0x5d9742c1c6d1bd7b2032626e12c97f2179fc78ba16e59f0603c547f0a04a2ad2

	d1 := `{"op" : "delete", "organization": "system", "address" : "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"}`
	txD1 := createTx(exec, []byte(d1), superPrivkey)
	txBytesD1 := common.ToHex(types.Encode(txD1))
	t.Log("tx = ", txBytesD1)
	// 0x0a17757365722e702e7465737470726f6f662e636f6e666967125d7b226f7022203a202264656c657465222c20226f7267616e697a6174696f6e223a202273797374656d222c20226164647265737322203a2022314b534264313748375a4b3869543337614a7a744642323258477773505464774534227d1a6d08011221023c712ab96682fdc1d7a9101f860bf07750b719221f01a56e0e0d61547650f2551a46304402206594b669433fc757976e7b998704fa19e1fc63be4274d2b9daa5d4bee847fec8022073681460c050c9ab9049c747ab6719e2b7136e541dc49daa674670234c5144c32080c2d72f30dffab3f8b299a2bd103a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47
	// 0xa4eebfe894257fc7e7abf75d192b5d77fc45ce6795c78d2d1ada0220af4f63c5

	d3 := `{"op" : "delete", "organization": "org1", "role": "member", "address" : "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k", "organization_note": "o1", "address_note": "o1-a1"}`
	txD3 := createTx(exec, []byte(d3), PrivKeyB)
	txBytesD3 := common.ToHex(types.Encode(txD3))
	t.Log("tx = ", txBytesD3)
	// 0x0a17757365722e702e7465737470726f6f662e636f6e66696712a1017b226f7022203a202264656c657465222c20226f7267616e697a6174696f6e223a20226f726731222c2022726f6c65223a20226d656d626572222c20226164647265737322203a2022314e4c4850456362545757787855336447555a426861796a7243484433707358376b222c20226f7267616e697a6174696f6e5f6e6f7465223a20226f31222c2022616464726573735f6e6f7465223a20226f312d6131227d1a6e080112210387bdf56091888507bfaf12034fbdb52ff412471edae0d11cff8a066695ec5e7f1a473045022100b4f7a7c1c51f4f85515f1e0c2fae85887d4f0e791e63130fc72553709b734ce1022004b7ce69c6f260e131ab84478637098b91df5bcdd14df0de011596064dd39b9b2080c2d72f3094afaa86aab7f8cb223a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47
	// 0x6a25335833cfec5633603a90ad418ac359f91fbabe32c38b5725272c769b9ff8

	d2 := `{"op" : "delete", "organization": "org1", "address" : "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"}`
	txD2 := createTx(exec, []byte(d2), superPrivkey)
	txBytesD2 := common.ToHex(types.Encode(txD2))
	t.Log("tx = ", txBytesD2)
	// 0x0a17757365722e702e7465737470726f6f662e636f6e666967125b7b226f7022203a202264656c657465222c20226f7267616e697a6174696f6e223a20226f726731222c20226164647265737322203a2022314a524e6a64457170344c4a356671796355426d396179434b536565736b674d4b52227d1a6d08011221023c712ab96682fdc1d7a9101f860bf07750b719221f01a56e0e0d61547650f2551a4630440220453489c6093e7c2c504fb5d0852f73dd779131fc23c459dced2b1b65a01f5bdb022041f49bcc6399f7ca80af385d0a5f1f25f71c375901fafd9f2ac4bab4756f8fde2080c2d72f308bfda7be93aab5a3433a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47
	// 0xf51cfca855a36b0fdea544c241feae48b8f2136f636a9ff09ce9061ed47d2cfa

	{
		a3 := `{"op" : "add", "organization": "org1", "role": "member", "address" : "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k", "organization_note": "o1", "address_note": "o1-a1"}`
		tx3 := createTx(exec, []byte(a3), superPrivkey)
		txBytes3 := common.ToHex(types.Encode(tx3))
		t.Log("tx = ", txBytes3)
		// 0x0a17757365722e702e7465737470726f6f662e636f6e666967129e017b226f7022203a2022616464222c20226f7267616e697a6174696f6e223a20226f726731222c2022726f6c65223a20226d656d626572222c20226164647265737322203a2022314e4c4850456362545757787855336447555a426861796a7243484433707358376b222c20226f7267616e697a6174696f6e5f6e6f7465223a20226f31222c2022616464726573735f6e6f7465223a20226f312d6131227d1a6d08011221023c712ab96682fdc1d7a9101f860bf07750b719221f01a56e0e0d61547650f2551a46304402207d41150b02f188738af93acec9896e5a175851a60633d7b853cf010816e0883b02204dd809f662d5439d26e1fbbb95383818d38f0496e63b6fada32f10dbaf8ee7ac2080c2d72f30ddb480fed988e19d7d3a22314e55646b684b4d4e33636378685564773964755458576e58686667784645796b47
		// 0x0b526b7ce2bcf5e221d2187a54c8e0be1370d35ca37806b4239ee5f83189f74e
	}

	//assert.Nil(t, t)
}
func TestCreateMember(t *testing.T) {
	p := payload{
		Op:             "add",
		Organization:   "sys",
		Role:           "mag",
		UserDetail:     &UserDetail{UserName: "name", UserIcon: "icon"},
		PersonalAuth:   &PersonalAuth{RealName: "real_name"},
		EnterpriseAuth: &EnterpriseAuth{EnterpriseName: "s"},
	}
	q := createMember(&p, "", &db.Block{})
	t.Logf("q:%+v", q)
	t.Logf("q.user:%+v", q.UserDetail)
}

func TestConvert(t *testing.T) {
	a1 := `{"op" : "add", "organization": "system", "role": "manager", "address" : "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4",  "address_note": "s1-a1"}`
	a2 := `{"op" : "add", "organization": "o1", "role": "manager", "address" : "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR", "organization_note": "o1", "address_note": "o1-a1"}`
	a3 := `{"op" : "add", "organization": "o1", "role": "member", "address" : "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k", "organization_note": "o1", "address_note": "o1-a1"}`
	d2 := `{"op" : "delete", "organization": "o1", "address" : "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"}`
	a22 := `{"op" : "add", "organization": "o1", "role": "manager", "address" : "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR", "organization_note": "o1", "address_note": "o1-a1"}`
	d3 := `{"op" : "delete", "organization": "o1", "role": "member", "address" : "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k", "organization_note": "o1", "address_note": "o1-a1"}`
	a32 := `{"op" : "add", "organization": "o1", "role": "member", "address" : "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k", "organization_note": "o1", "address_note": "o1-a1"}`

	exec := "user.p.test1.config"
	tx1 := createTx(exec, []byte(a1), PrivKeyD)
	tx2 := createTx(exec, []byte(a2), PrivKeyD)
	tx3 := createTx(exec, []byte(a3), PrivKeyD)
	tx4 := createTx(exec, []byte(d2), PrivKeyA)
	tx5 := createTx(exec, []byte(a22), PrivKeyA)
	tx6 := createTx(exec, []byte(d3), PrivKeyB)
	tx7 := createTx(exec, []byte(a32), PrivKeyB)
	block := &types.BlockDetail{
		Block: &types.Block{
			Txs:    []*types.Transaction{tx1, tx2, tx3, tx4, tx5, tx6, tx7},
			Height: 1,
		},
		Receipts: []*types.ReceiptData{
			{Ty: types.ExecPack}, {Ty: types.ExecPack}, {Ty: types.ExecPack}, {Ty: types.ExecPack},
			{Ty: types.ExecPack}, {Ty: types.ExecPack}, {Ty: types.ExecPack},
		},
		KV: make([]*types.KeyValue, 7),
	}

	// 初始化管理员
	testDB := newDummyDBForTest()
	db1 := NewConfigDB(testDB)
	m1 := Member{
		Address:      string(Nodes[3]),
		Role:         "manager",
		Organization: "system",
		Note:         "test",
	}
	db1.SetMember(string(Nodes[3]), m1.genRecord(AddOpX, db.OpAdd))

	convert := NewConvert("user.p.test1.", "PROOF", nil)
	c := convert.(db.NeedWrapDB)
	c.SetDB(testDB)

	env := db.TxEnv{
		Block:     block,
		TxIndex:   0,
		BlockHash: "blockHash1",
	}
	rs, err := convert.ConvertTx(&env, db.SeqTypeAdd)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))
	debugRecords(t, "add super manager", rs)
	k0 := orgKey("system")
	v0 := `{"organization":"system","note":"","count":1,"height":1,"ts":0,"block_hash":"blockHash1","index":0,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100000}`
	k1 := "proof_config/proof_config/member-1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
	v1 := `{"address":"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4","role":"manager","organization":"system","note":"s1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":0,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100000}`
	assert.Equal(t, k0, rs[0].Key())
	assert.Equal(t, v0, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k1, rs[1].Key())
	assert.Equal(t, v1, ignoreTxHash(string(rs[1].Value())))

	env.TxIndex = 1
	rs, err = convert.ConvertTx(&env, db.SeqTypeAdd)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))
	debugRecords(t, "add manager", rs)
	k10 := orgKey("o1")
	v10 := `{"organization":"o1","note":"o1","count":1,"height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k11 := "proof_config/proof_config/member-1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
	v11 := `{"address":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR","role":"manager","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	assert.Equal(t, k10, rs[0].Key())
	assert.Equal(t, v10, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k11, rs[1].Key())
	assert.Equal(t, v11, ignoreTxHash(string(rs[1].Value())))

	env.TxIndex = 2
	rs, err = convert.ConvertTx(&env, db.SeqTypeAdd)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))
	debugRecords(t, "add member", rs)
	k20 := orgKey("o1")
	v20 := `{"organization":"o1","note":"o1","count":2,"height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k21 := "proof_config/proof_config/member-1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
	v21 := `{"address":"1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k","role":"member","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":2,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100002}`
	assert.Equal(t, k20, rs[0].Key())
	assert.Equal(t, v20, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k21, rs[1].Key())
	assert.Equal(t, v21, ignoreTxHash(string(rs[1].Value())))

	env.TxIndex = 3
	rs, err = convert.ConvertTx(&env, db.SeqTypeAdd)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(rs))
	debugRecords(t, "del manager", rs)
	k30 := orgKey("o1")
	v30 := `{"organization":"o1","note":"o1","count":1,"height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k31 := "proof_config/proof_config/member-1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
	v31 := `{"address":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR","role":"manager","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k32 := `proof_config_delete/proof_config_delete/member-del-100003-1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR`
	v32 := `{"address":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR","role":"manager","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001,"delete":{"height":1,"ts":0,"block_hash":"blockHash1","index":3,"send":"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4","tx_hash":"","height_index":100003}}`
	assert.Equal(t, k30, rs[0].Key())
	assert.Equal(t, v30, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k31, rs[1].Key())
	assert.Equal(t, v31, ignoreTxHash(string(rs[1].Value())))
	assert.Equal(t, k32, rs[2].Key())
	assert.Equal(t, v32, ignoreTxHash(string(rs[2].Value())))

	env.TxIndex = 4
	rs, err = convert.ConvertTx(&env, db.SeqTypeAdd)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))
	debugRecords(t, "add manager again", rs)
	k40 := orgKey("o1")
	v40 := `{"organization":"o1","note":"o1","count":2,"height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k41 := "proof_config/proof_config/member-1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
	v41 := `{"address":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR","role":"manager","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":4,"send":"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4","tx_hash":"","height_index":100004}`
	assert.Equal(t, k40, rs[0].Key())
	assert.Equal(t, v40, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k41, rs[1].Key())
	assert.Equal(t, v41, ignoreTxHash(string(rs[1].Value())))

	env.TxIndex = 5
	rs, err = convert.ConvertTx(&env, db.SeqTypeAdd)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(rs))
	debugRecords(t, "del member", rs)
	k50 := orgKey("o1")
	v50 := `{"organization":"o1","note":"o1","count":1,"height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k51 := `proof_config/proof_config/member-1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k`
	v51 := `{"address":"1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k","role":"member","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":2,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100002}`
	k52 := `proof_config_delete/proof_config_delete/member-del-100005-1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k`
	v52 := `{"address":"1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k","role":"member","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":2,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100002,"delete":{"height":1,"ts":0,"block_hash":"blockHash1","index":5,"send":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR","tx_hash":"","height_index":100005}}`
	assert.Equal(t, k50, rs[0].Key())
	assert.Equal(t, v50, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k51, rs[1].Key())
	assert.Equal(t, v51, ignoreTxHash(string(rs[1].Value())))
	assert.Equal(t, k52, rs[2].Key())
	assert.Equal(t, v52, ignoreTxHash(string(rs[2].Value())))

	env.TxIndex = 6
	rs, err = convert.ConvertTx(&env, db.SeqTypeAdd)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))
	debugRecords(t, "add member again", rs)
	k60 := orgKey("o1")
	v60 := `{"organization":"o1","note":"o1","count":2,"height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k61 := "proof_config/proof_config/member-1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
	v61 := `{"address":"1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k","role":"member","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":6,"send":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR","tx_hash":"","height_index":100006}`
	assert.Equal(t, k60, rs[0].Key())
	assert.Equal(t, v60, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k61, rs[1].Key())
	assert.Equal(t, v61, ignoreTxHash(string(rs[1].Value())))

	env.TxIndex = 6
	rs, err = convert.ConvertTx(&env, db.SeqTypeDel)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))
	debugRecords(t, "add member again ROOLBACK", rs)
	k60R := orgKey("o1")
	v60R := `{"organization":"o1","note":"o1","count":1,"height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k61R := "proof_config/proof_config/member-1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
	v61R := `{"address":"1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k","role":"member","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":6,"send":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR","tx_hash":"","height_index":100006}`
	assert.Equal(t, k60R, rs[0].Key())
	assert.Equal(t, v60R, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k61R, rs[1].Key())
	assert.Equal(t, v61R, ignoreTxHash(string(rs[1].Value())))

	env.TxIndex = 5
	rs, err = convert.ConvertTx(&env, db.SeqTypeDel)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(rs))
	debugRecords(t, "del member ROOLBACK", rs)
	k50R := orgKey("o1")
	v50R := `{"organization":"o1","note":"o1","count":2,"height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k51R := `proof_config/proof_config/member-1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k`
	v51R := `{"address":"1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k","role":"member","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":2,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100002}`
	k52R := `proof_config_delete/proof_config_delete/member-del-100005-1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k`
	v52R := `{"address":"1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k","role":"member","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":2,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100002,"delete":{"height":1,"ts":0,"block_hash":"blockHash1","index":5,"send":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR","tx_hash":"","height_index":100005}}`
	assert.Equal(t, k50R, rs[0].Key())
	assert.Equal(t, v50R, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k51R, rs[1].Key())
	assert.Equal(t, v51R, ignoreTxHash(string(rs[1].Value())))
	assert.Equal(t, k52R, rs[2].Key())
	assert.Equal(t, v52R, ignoreTxHash(string(rs[2].Value())))

	env.TxIndex = 4
	rs, err = convert.ConvertTx(&env, db.SeqTypeDel)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))
	debugRecords(t, "add manager again ROOLBACK", rs)
	k40R := orgKey("o1")
	v40R := `{"organization":"o1","note":"o1","count":1,"height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k41R := "proof_config/proof_config/member-1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
	v41R := `{"address":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR","role":"manager","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":4,"send":"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4","tx_hash":"","height_index":100004}`
	assert.Equal(t, k40R, rs[0].Key())
	assert.Equal(t, v40R, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k41R, rs[1].Key())
	assert.Equal(t, v41R, ignoreTxHash(string(rs[1].Value())))

	env.TxIndex = 3
	rs, err = convert.ConvertTx(&env, db.SeqTypeDel)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(rs))
	debugRecords(t, "del manager ROOLBACK", rs)
	k30R := orgKey("o1")
	v30R := `{"organization":"o1","note":"o1","count":2,"height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k31R := "proof_config/proof_config/member-1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
	v31R := `{"address":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR","role":"manager","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k32R := `proof_config_delete/proof_config_delete/member-del-100003-1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR`
	v32R := `{"address":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR","role":"manager","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001,"delete":{"height":1,"ts":0,"block_hash":"blockHash1","index":3,"send":"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4","tx_hash":"","height_index":100003}}`
	assert.Equal(t, k30R, rs[0].Key())
	assert.Equal(t, v30R, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k31R, rs[1].Key())
	assert.Equal(t, v31R, ignoreTxHash(string(rs[1].Value())))
	assert.Equal(t, k32R, rs[2].Key())
	assert.Equal(t, v32R, ignoreTxHash(string(rs[2].Value())))

	env.TxIndex = 2
	rs, err = convert.ConvertTx(&env, db.SeqTypeDel)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))
	debugRecords(t, "add member ROOLBACK", rs)
	k20R := orgKey("o1")
	v20R := `{"organization":"o1","note":"o1","count":1,"height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k21R := "proof_config/proof_config/member-1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
	v21R := `{"address":"1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k","role":"member","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":2,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100002}`
	assert.Equal(t, k20R, rs[0].Key())
	assert.Equal(t, v20R, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k21R, rs[1].Key())
	assert.Equal(t, v21R, ignoreTxHash(string(rs[1].Value())))

	env.TxIndex = 1
	rs, err = convert.ConvertTx(&env, db.SeqTypeDel)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))
	debugRecords(t, "add manager ROOLBACK", rs)
	k10R := orgKey("o1")
	v10R := `{"organization":"o1","note":"o1","count":0,"height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	k11R := "proof_config/proof_config/member-1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
	v11R := `{"address":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR","role":"manager","organization":"o1","note":"o1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":1,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100001}`
	assert.Equal(t, k10R, rs[0].Key())
	assert.Equal(t, v10R, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k11R, rs[1].Key())
	assert.Equal(t, v11R, ignoreTxHash(string(rs[1].Value())))

	env.TxIndex = 0
	rs, err = convert.ConvertTx(&env, db.SeqTypeDel)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rs))
	debugRecords(t, "add super manager ROOLBACK", rs)
	k0R := orgKey("system")
	v0R := `{"organization":"system","note":"","count":0,"height":1,"ts":0,"block_hash":"blockHash1","index":0,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100000}`
	k1R := "proof_config/proof_config/member-1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
	v1R := `{"address":"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4","role":"manager","organization":"system","note":"s1-a1","height":1,"ts":0,"block_hash":"blockHash1","index":0,"send":"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs","tx_hash":"","height_index":100000}`
	assert.Equal(t, k0R, rs[0].Key())
	assert.Equal(t, v0R, ignoreTxHash(string(rs[0].Value())))
	assert.Equal(t, k1R, rs[1].Key())
	assert.Equal(t, v1R, ignoreTxHash(string(rs[1].Value())))

	// 全部回滚后剩下 超级管理员
	assert.Equal(t, 3, len(testDB.db))
	assert.Equal(t, 1, len(testDB.db[DBX]))
	assert.Equal(t, 0, len(testDB.db[DeleteDBX]))
	assert.Equal(t, 0, len(testDB.db[OrgDBX]))
}

// 交易hash那次都不一样
func ignoreTxHash(v string) string {
	reg := regexp.MustCompile(`tx_hash":"[a-zA-Z0-9]*`)
	return string(reg.ReplaceAll([]byte(v), []byte(`tx_hash":"`)))
}

func TestIgnoreTxHash(t *testing.T) {
	v := `xxxx "tx_hash":"0xb31eadf2c10515029ea8045689f01ec097ded0eeca5ae7cac5a9b61b5a740169"`
	e := `xxxx "tx_hash":""`
	assert.Equal(t, e, ignoreTxHash(v))
}

func debugRecords(t *testing.T, msg string, rs []db.Record) {
	for i, r := range rs {
		t.Log(i, msg, r.OpType(), r.Key(), string(r.Value()))
	}
}

func createTx(exec string, payload []byte, privkey string) *types.Transaction {
	tx := &types.Transaction{Payload: payload}
	tx.Execer = []byte(exec)
	tx.Nonce = rand.Int63() // nolint
	tx.Fee = 100000000
	tx.To = util2.AddressConvert(db.ExecAddress(string(tx.Execer)))

	priv := util.HexToPrivkey(privkey)
	tx.Sign(types.SECP256K1, priv)
	return tx
}

type dummyDB struct {
	db map[string]map[string]*json.RawMessage
}

// newConfigDBForTest 测试用
func newDummyDBForTest() *dummyDB {
	return &dummyDB{db: make(map[string]map[string]*json.RawMessage)}
}

func (d *dummyDB) Get(k1, k2, id string) (*json.RawMessage, error) {
	m1, ok := d.db[k1]
	if !ok {
		return nil, db.ErrDBNotFound
	}
	v1, ok := m1[id]
	if !ok {
		return nil, db.ErrDBNotFound
	}
	return v1, nil
}

func (d *dummyDB) Set(k1, k2, id string, r db.Record) error {
	if _, ok := d.db[k1]; !ok {
		d.db[k1] = make(map[string]*json.RawMessage)
	}
	if r.OpType() == db.OpDel {
		delete(d.db[k1], id)
		return nil
	}
	v1 := json.RawMessage(r.Value())
	d.db[k1][id] = &v1
	return nil
}

// List dummy
func (d *dummyDB) List(k1, k2 string, kv []*db.ListKV) ([]*json.RawMessage, error) {
	return nil, nil
}

func (d *dummyDB) Del(k1, k2, id string) error {
	if m1, ok := d.db[k1]; ok {
		delete(m1, id)
	}
	return nil
}

func (d *dummyDB) Search(idx, typ string, query *querypara.Query, decode func(x *json.RawMessage) (interface{}, error)) ([]interface{}, error) {
	return nil, nil
}

func orgKey(x string) string {
	return "proof_config_org/proof_config_org/org-" + x
}
