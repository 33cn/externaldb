package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	proofconfig "github.com/33cn/externaldb/db/proof_config"
	"github.com/33cn/externaldb/escli/querypara"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	"github.com/33cn/externaldb/db"
	dbcom "github.com/33cn/externaldb/db/common"
	"github.com/33cn/externaldb/db/proof/api"
	"github.com/33cn/externaldb/db/proof/model"
	proofUtil "github.com/33cn/externaldb/util/proof"
	"github.com/stretchr/testify/assert"
)

// TODO fix test

//addr:12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
var Priv = "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
var zeroHash [32]byte

// var rsaPrivKey = "0x3078020100300b06092a864886f70d01010104663064020100021100ce8fee27c53555fc9bd4fd40e96fc5990203010001021100a418f3b9e4915a9cc648719de8926f81020900e2e6f864dd1ca567020900e90d368c9e1a5cff020900cd53674996d13a57020900ca445d83cdf4b3a1020863997caaaf1c973a"
var rsaPubKey = "0x302c300d06092a864886f70d0101010500031b003018021100ce8fee27c53555fc9bd4fd40e96fc5990203010001"

//HexToPrivkey ： convert hex string to private key
func HexToPrivkey(key string) crypto.PrivKey {
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	if err != nil {
		panic(err)
	}
	bkey, err := common.FromHex(key)
	if err != nil {
		panic(err)
	}
	priv, err := cr.PrivKeyFromBytes(bkey)
	if err != nil {
		panic(err)
	}
	return priv
}

func Test_Proof(t *testing.T) {
	//k:v
	jsonstr1 := "{\"key\":\"user-id-0\",\"data\":{\"value\":\"2006-01-02 15:04:05\",\"type\":\"date\",\"format\":\"rfc\"}}"
	//k:[v1,v2]
	jsonstr2 := "{\"key\": \"user-id-1\",\"data\": [{\"value\": \"0.123456789\",\"type\": \"number\",\"format\": \"float64\"}, {\"value\": \"2\",\"type\": \"number\",\"format\": \"int64\"}]}"
	// k:{k:v}
	jsonstr3 := "{\"key\": \"user-id-2\",\"data\": {\"key\": \"user-id-3\",\"data\": {\"value\": \"1\",\"type\": \"number\",\"format\": \"int32\"}}}"
	// k:[k:v,k:v]
	jsonstr4 := "{\"key\": \"user-id-4\",\"data\": [{\"key\": \"user-id-5\",\"data\": {\"value\": \"1\",\"type\": \"number\",\"format\": \"int32\"}}, {\"key\": \"user-id-6\",\"data\": {\"value\": \"2\",\"type\": \"number\",\"format\": \"int32\"}}]}"
	jsonstr5 := "{\"key\": \"ref_hashes\",\"data\": [{\"value\": \"0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f\",\"type\": \"file\",\"format\": \"hash\"}, {\"value\": \"0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56\",\"type\": \"file\",\"format\": \"hash\"}]}"

	jsonstr := "[" + jsonstr1 + "," + jsonstr2 + "," + jsonstr3 + "," + jsonstr4 + "," + jsonstr5 + "]"
	// jsonstr := "[" + jsonstr5 + "]"

	// 对存证内容进行rsa的加密使用rsaPubKey
	enJSONStr, err := dbcom.PublicEncrypt(rsaPubKey, jsonstr)
	assert.Nil(t, err)
	xx := dbcom.EncodeToString([]byte(enJSONStr))

	optionStr := `{"encrypt":"rsa","archive":"zip"}`
	enOptionStr := dbcom.EncodeToString([]byte(optionStr))

	notestr := "{\"userName\":\"18701358726\",\"userIcon\":\"\",\"evidenceName\":\"公益捐赠yyh\",\"stepName\":\"\",\"version\":1}"
	enNoteStr := dbcom.EncodeToString([]byte(notestr))

	payload := "{\"version\":\"1.0.0\",\"option\":\"" + enOptionStr + "\",\"data\":\"" + xx + "\",\"note\":\"" + enNoteStr + "\"}"

	t.Log("Test_Proof", "payload", payload)

	var payload2 api.ProofInfo
	err = json.Unmarshal([]byte(payload), &payload2)
	if err != nil {
		t.Log("Test_Proof1111err", "err", err)

		return
	}

	yy, _ := dbcom.DecodeString(payload2.Data)
	t.Log("Test_Proof1111", "payload.data", string(yy))

	t.Log("Test_Proof1111", "payload.Version", payload2.Version)

	t.Log("Test_Proof1111", "payload.Option", payload2.Option)

	tx := &types.Transaction{Execer: []byte("user.p.sanhe.proof"), Payload: []byte(payload)}
	tx.To = db.ExecAddress("user.p.sanhe.proof")
	priv := HexToPrivkey(Priv)
	if priv != nil {
		tx.Sign(types.SECP256K1, priv)
	}

	//构建区块
	newblock := &types.Block{}
	newblock.Height = 1
	newblock.BlockTime = types.Now().Unix()
	newblock.ParentHash = zeroHash[:]
	newblock.Txs = []*types.Transaction{tx}
	newblock.TxHash = zeroHash[:]
	env := db.TxEnv{
		TxIndex:   0,
		Block:     &types.BlockDetail{},
		BlockHash: common.ToHex(newblock.HashByForkHeight(0)),
	}
	blockTime := newblock.BlockTime
	blockHash := env.BlockHash
	env.Block.Block = newblock
	rec := types.ReceiptData{Ty: types.ExecPack}
	env.Block.Receipts = append(env.Block.Receipts, &rec)

	convert := NewConvert("sanhe", "bty", "proof")
	c := convert.(*ProofConvert)
	c.RecordGen = NewRecordGen("p1", "p2", "l1", "l2", "t1", "t2", "u1", "u2")
	c.configDB = &proofconfig.None{}
	records, err := convert.ConvertTx(&env, db.SeqTypeAdd)

	assert.Nil(t, err)
	for i, r := range records {
		t.Log(string(r.Value()))
		if i == 0 && !proofUtil.IsParserToKV {
			decodeProof(t, r.Value())
		}
	}

	re := records[0].(*ProofRecord)
	t.Log("Test_Proof", "ikey", re.Key())
	t.Log("Test_Proof", "OpType", re.OpType())

	pInfo := re.Proof.Proof

	assert.Equal(t, common.ToHex(tx.Hash()), pInfo["proof_tx_hash"])
	assert.Equal(t, fmt.Sprintf("%s-%s", model.ProofID, common.ToHex(tx.Hash())), pInfo["proof_id"])

	assert.Equal(t, int64(1), pInfo["proof_height"])
	assert.Equal(t, blockTime, pInfo["proof_block_time"])
	assert.Equal(t, blockHash, pInfo["proof_block_hash"])

	assert.Equal(t, "fuzamei", pInfo["proof_organization"])
	assert.Equal(t, "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", pInfo["proof_sender"])
	assert.Equal(t, int64(100000), pInfo["proof_height_index"])
	assert.Equal(t, notestr, pInfo["proof_note"])

	if !proofUtil.IsParserToKV {
		assert.Equal(t, int64(1136214245), pInfo["user-id-0"])
		assert.Equal(t, 0.123456789, pInfo["user-id-1"].([]interface{})[0])
		assert.Equal(t, int64(2), pInfo["user-id-1"].([]interface{})[1])

		assert.Equal(t, int64(1), pInfo["user-id-2"].(map[string]interface{})["user-id-3"])

		assert.Equal(t, int64(1), pInfo["user-id-4"].([]interface{})[0].(map[string]interface{})["user-id-5"])

		assert.Equal(t, int64(2), pInfo["user-id-4"].([]interface{})[1].(map[string]interface{})["user-id-6"])

		assert.Equal(t, "0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f", pInfo["ref_hashes"].([]interface{})[0])
		assert.Equal(t, "0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56", pInfo["ref_hashes"].([]interface{})[1])
	} else {
		assert.Equal(t, int64(1136214245), pInfo["user-id-0"])
		assert.Equal(t, 0.123456789, pInfo["user-id-1"].([]interface{})[0])
		assert.Equal(t, int64(2), pInfo["user-id-1"].([]interface{})[1])

		assert.Equal(t, int64(1), pInfo["user-id-3"])

		assert.Equal(t, int64(1), pInfo["user-id-5"])

		assert.Equal(t, int64(2), pInfo["user-id-6"])

		assert.Equal(t, "0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f", pInfo["ref_hashes"].([]interface{})[0])
		assert.Equal(t, "0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56", pInfo["ref_hashes"].([]interface{})[1])

	}
}

func Test_Proof1(t *testing.T) {
	//hexpayload := "0x7b2264617461223a22573373695a47463059534936573373695a4746305953493665794a306558426c496a6f696447563464434973496d5a76636d316864434936496e4e30636d6c755a794973496e5a686248566c496a6f69353636583559716235706d3635627154496e3073496e5235634755694f6a4173496d746c65534936497557506b656976676561637575616568434973496d7868596d5673496a6f6935592b52364b2b423570793635703645496e307365794a6b59585268496a7037496e5235634755694f694a305a586830496977695a6d397962574630496a6f69633352796157356e49697769646d4673645755694f694c6b7571666b754a726c6a4c726c6e5a66706b37377071356a6e7571666b7572726d6959336f7234486b756159696653776964486c775a5349364d43776961325635496a6f69364b2b42354c6d6d355a434e3536657749697769624746695a5777694f694c6f7234486b7561626c6b49336e7037416966537837496d5268644745694f6e736964486c775a534936496e526c654851694c434a6d62334a74595851694f694a7a64484a70626d63694c434a32595778315a534936497561646a7561496b4f693070434a394c434a306558426c496a6f774c434a725a586b694f694c6c7261626c6b5a6a6c7035506c6b4930694c434a7359574a6c6243493649755774707557526d4f576e6b2b57516a534a394c4873695a4746305953493665794a306558426c496a6f696447563464434973496d5a76636d316864434936496e4e30636d6c755a794973496e5a686248566c496a6f6935355333496e3073496e5235634755694f6a4173496d746c6553493649755774707557526d4f6141702b574971794973496d7868596d5673496a6f693561326d355a4759356f436e35596972496e307365794a6b59585268496a7037496e5235634755694f694a305a586830496977695a6d397962574630496a6f69633352796157356e49697769646d4673645755694f694a54544552594d6a41794d4441314d6a41774d4467696653776964486c775a5349364d43776961325635496a6f69364b2b42354c6d6d3537795735592b3349697769624746695a5777694f694c6f7234486b7561626e764a626c6a37636966537837496d5268644745694f6e736964486c775a534936496e526c654851694c434a6d62334a74595851694f694a7a64484a70626d63694c434a32595778315a534936496a49774d6a41744d4455744d5445696653776964486c775a5349364d43776961325635496a6f69356279413561326d35706532365a653049697769624746695a5777694f694c6c7649446c7261626d6c3762706c37516966537837496d5268644745694f6e736964486c775a534936496e526c654851694c434a6d62334a74595851694f694a7a64484a70626d63694c434a32595778315a534936496a49774d6a41744d4455744d546b696653776964486c775a5349364d43776961325635496a6f6935377554354c696135706532365a653049697769624746695a5777694f694c6e7535506b754a726d6c3762706c37516966563073496e5235634755694f6a4d73496d746c655349364975697667655335707569767075614468534a3958513d3d222c2276657273696f6e223a22312e302e30222c226f7074696f6e223a2265794a6c626d4e7965584230496a6f694969776959584a6a61476c325a53493649694a39227d"
	//hexpayload := "0x7b2264617461223a22573373695a47463059534936573373695a4746305953493665794a306558426c496a6f696447563464434973496d5a76636d316864434936496e4e30636d6c755a794973496e5a686248566c496a6f6935364b6e3571433535703663496e3073496e5235634755694f6a4173496d746c6553493649755336702b5754676557516a65656e73434973496d7868596d5673496a6f69354c716e355a4f42355a434e35366577496e307365794a6b59585268496a7037496e5235634755694f694a305a586830496977695a6d397962574630496a6f69633352796157356e49697769646d4673645755694f694c6d745a6e6d735a2f6d6e61336c7435376b754c546c726f6e706c4b626c6a4a666f6f5a6670675a4d696653776964486c775a5349364d43776961325635496a6f6935355366354c716e355a574749697769624746695a5777694f694c6e6c4a2f6b7571666c6c59596966563073496e5235634755694f6a4d73496d746c6553493649755336702b5754676557506775615673434a394c4873695a47463059534936573373695a4746305953493665794a306558426c496a6f696447563464434973496d5a76636d316864434936496e4e30636d6c755a794973496e5a686248566c496a6f69364b654235354f3235592b6a356f6957356f6d5435364342496e3073496e5235634755694f6a4173496d746c65534936497565556e2b5336702b6158706561636e794973496d7868596d5673496a6f6935355366354c716e3570656c35707966496e307365794a6b59585268496a7037496e5235634755694f694a305a586830496977695a6d397962574630496a6f69633352796157356e49697769646d4673645755694f694c706d4c546c68346e6c75624c6e6836586c704951696653776964486c775a5349364d43776961325635496a6f6935594b6f356132593570613535724f5649697769624746695a5777694f694c6c67716a6c725a6a6d6c726e6d7335556966537837496d5268644745694f6e736964486c775a534936496e526c654851694c434a6d62334a74595851694f694a7a64484a70626d63694c434a32595778315a534936496a49304d4f576b71534a394c434a306558426c496a6f774c434a725a586b694f694c6b7635336f7234486d6e4a2f706d5a41694c434a7359574a6c624349364975532f6e656976676561636e2b6d5a6b434a394c4873695a4746305953493665794a306558426c496a6f696447563464434973496d5a76636d316864434936496e4e30636d6c755a794973496e5a686248566c496a6f693562794135614f7a3559327a36614f66496e3073496e5235634755694f6a4173496d746c6553493649756d6a6e2b6555714f61577565617a6c534973496d7868596d5673496a6f6936614f663535536f3570613535724f56496e31644c434a306558426c496a6f7a4c434a725a586b694f694c706f352f6c6b34486c7436586f69626f6966537837496d5268644745694f6c7437496d5268644745694f6e736964486c775a534936496e526c654851694c434a6d62334a74595851694f694a7a64484a70626d63694c434a32595778315a534936496a4934496e3073496e5235634755694f6a4173496d746c6553493649755337742b616776434973496d7868596d5673496a6f69354c753335714338496e307365794a6b59585268496a7037496e5235634755694f694a305a586830496977695a6d397962574630496a6f69633352796157356e49697769646d4673645755694f694c6c7062626d73726e6c6b624d314d44446c68597376354c696b35373251496e3073496e5235634755694f6a4173496d746c6553493649756d486a656d486a794973496d7868596d5673496a6f693659654e36596550496e31644c434a306558426c496a6f7a4c434a725a586b694f694c6b7571666c6b34486f7034546d6f4c77696656303d222c2276657273696f6e223a22312e302e30222c226f7074696f6e223a2265794a6c626d4e7965584230496a6f694969776959584a6a61476c325a53493649694a39227d"
	hexpayload := "0x7b2264617461223a22573373695a47463059534936573373695a4746305953493665794a306558426c496a6f696447563464434973496d5a76636d316864434936496e4e30636d6c755a794973496e5a686248566c496a6f694d5449696653776964486c775a5349364d43776961325635496a6f69356f3251364c5767354c713649697769624746695a5777694f694c6d6a5a446f7461446b75726f6966537837496d5268644745694f6e736964486c775a534936496e526c654851694c434a6d62334a74595851694f694a7a64484a70626d63694c434a32595778315a534936496a4579496e3073496e5235634755694f6a4173496d746c655349364975614e6b4f69316f4f53367575532f6f65614272794973496d7868596d5673496a6f69356f3251364c5767354c7136354c2b68356f4776496e307365794a6b59585268496a7037496e5235634755694f694a305a586830496977695a6d397962574630496a6f69633352796157356e49697769646d4673645755694f694c6d6c72446c6e6f766c6871446e6972626e6c34586d72354c6f6772726e676f376e6c71766d673455696653776962334230615739756379493657794c6d6c72446c6e6f766c6871446e6972626e6c34586d72354c6f6772726e676f376e6c71766d673455694c434c6e7372376c6834626d6962626f744b766d6c5a486c69716b694c434c6c75497a6d6e4a766c7436586e7149766e694c486c7634506c69716e6c726159695853776964486c775a5349364e53776961325635496a6f69355957733535754b366147353535757549697769624746695a5777694f694c6c68617a6e6d3472706f626e6e6d36346966537837496d5268644745694f6e736964486c775a534936496e526c654851694c434a6d62334a74595851694f694a7a64484a70626d63694c434a32595778315a534936496a4978496e3073496e5235634755694f6a4173496d746c655349364975614e6b4f69316f4f5735732b575073434973496d7868596d5673496a6f69356f3251364c576735626d7a35592b77496e307365794a6b59585268496a7037496e5235634755694f694a6b5958526c496977695a6d397962574630496a6f696458526a49697769646d4673645755694f6a45314f5441314f5455794d4442394c434a306558426c496a6f324c434a725a586b694f694c6d6a5a446f7461446d6c3762706c3751694c434a7359574a6c624349364975614e6b4f69316f4f615874756d5874434a394c4873695a4746305953493665794a306558426c496a6f696447563464434973496d5a76636d316864434936496e4e30636d6c755a794973496e5a686248566c496a6f69354c71363572435235626942496e3073496d397764476c76626e4d694f6c7369354c713635724352356269424969776935597937353561583535536f355a4f424969776935355366357253373535536f355a4f424969776936614f66355a4f424969776935595732354c7557496c3073496e5235634755694f6a5573496d746c655349364975614e6b4f69316f4f654a7165693168434973496d7868596d5673496a6f69356f3251364c576735346d70364c5745496e307365794a6b59585268496a7037496e5235634755694f694a305a586830496977695a6d397962574630496a6f69633352796157356e49697769646d4673645755694f69497a4d694a394c434a306558426c496a6f774c434a725a586b694f694c6d6a5a446f7461446d6c624470683438694c434a7359574a6c624349364975614e6b4f69316f4f6156734f6d486a794a394c4873695a4746305953493665794a306558426c496a6f696447563464434973496d5a76636d316864434936496e4e30636d6c755a794973496e5a686248566c496a6f694d7a49696653776964486c775a5349364d43776961325635496a6f693561534835724f6f49697769624746695a5777694f694c6c7049666d7336676966537837496d5268644745694f6e736964486c775a534936496e526c654851694c434a6d62334a74595851694f694a7a64484a70626d63694c434a32595778315a534936496a4d79496e3073496e5235634755694f6a4173496d746c655349364975572f672b6145762b5776684f697672534973496d7868596d5673496a6f6935622b44356f532f35612b45364b2b74496e307365794a6b59585268496a7037496e5235634755694f694a305a586830496977695a6d397962574630496a6f69633352796157356e49697769646d4673645755694f69497a4d694a394c434a306558426c496a6f774c434a725a586b694f694c6d6a5a446f7461446c6836336f723445694c434a7359574a6c624349364975614e6b4f69316f4f57487265697667534a395853776964486c775a5349364d79776961325635496a6f69364b2b42354c6d6d354c2b68356f4776496e307365794a6b59585268496a7037496e5235634755694f694a305a586830496977695a6d397962574630496a6f69633352796157356e49697769646d4673645755694f694c6c68617a6e6d34726d6a5a446f74614235655767696653776964486c775a5349364d43776961325635496a6f6935613259364b2b42355a434e3536657749697769624746695a5777694f694c6c725a6a6f7234486c6b49336e703741696656303d222c2276657273696f6e223a22312e302e30222c226f7074696f6e223a2265794a6c626d4e7965584230496a6f694969776959584a6a61476c325a53493649694a39222c226e6f7465223a22496e7463496e567a5a584a4f5957316c58434936584349784f4463774d544d314f4463794e6c77694c46776964584e6c636b6c6a62323563496a7063496c77694c4677695a585a705a4756755932564f5957316c5843493658434c6c68617a6e6d34726d6a5a446f746142356557686349697863496e4e305a58424f5957316c5843493658434a6349697863496e5a6c636e4e7062323563496a6f786653493d227d"
	bytepayload, err := common.FromHex(hexpayload)
	if err != nil {
		panic(err)
	}
	tx := &types.Transaction{Execer: []byte("user.p.sanhe.proof"), Payload: bytepayload}
	tx.To = db.ExecAddress("user.p.sanhe.proof")
	priv := HexToPrivkey(Priv)
	if priv != nil {
		tx.Sign(types.SECP256K1, priv)
	}

	//构建区块
	newblock := &types.Block{}
	newblock.Height = 1
	newblock.BlockTime = types.Now().Unix()
	newblock.ParentHash = zeroHash[:]
	newblock.Txs = []*types.Transaction{tx}
	newblock.TxHash = zeroHash[:]
	env := db.TxEnv{
		TxIndex:   0,
		Block:     &types.BlockDetail{},
		BlockHash: common.ToHex(newblock.HashByForkHeight(0)),
	}
	blockTime := newblock.BlockTime
	blockHash := env.BlockHash
	env.Block.Block = newblock
	rec := types.ReceiptData{Ty: types.ExecPack}
	env.Block.Receipts = append(env.Block.Receipts, &rec)

	convert := NewConvert("sanhe", "bty", "proof")
	c := convert.(*ProofConvert)
	c.RecordGen = NewRecordGen("p1", "p2", "l1", "l2", "t1", "t2", "u1", "u2")
	c.configDB = &proofconfig.None{}
	records, err := convert.ConvertTx(&env, db.SeqTypeAdd)

	assert.Nil(t, err)
	for _, r := range records {
		t.Log(string(r.Value()))
	}

	re := records[0].(*ProofRecord)
	t.Log("Test_Proof", "ikey", re.Key())
	t.Log("Test_Proof", "OpType", re.OpType())

	pInfo := re.Proof.Proof

	assert.Equal(t, common.ToHex(tx.Hash()), pInfo["proof_tx_hash"])
	assert.Equal(t, fmt.Sprintf("%s-%s", model.ProofID, common.ToHex(tx.Hash())), pInfo["proof_id"])

	assert.Equal(t, int64(1), pInfo["proof_height"])
	assert.Equal(t, blockTime, pInfo["proof_block_time"])
	assert.Equal(t, blockHash, pInfo["proof_block_hash"])

	assert.Equal(t, "fuzamei", pInfo["proof_organization"])
	assert.Equal(t, "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", pInfo["proof_sender"])
	assert.Equal(t, int64(100000), pInfo["proof_height_index"])

}

func Test_ProofNestedCombination(t *testing.T) {
	payload := creatProofpayload(1)
	testProofNestedCombination(t, payload)
	payload = creatProofpayload(2)
	testProofNestedCombination(t, payload)
	payload = creatProofpayload(3)
	testProofNestedCombination(t, payload)
	payload = creatProofpayload(4)
	testProofNestedCombination(t, payload)
}

func creatProofpayload(flag int) string {

	//构建一个proof存证区块，存证类容如下;
	//k:v
	//jsonstr1 := "{\"key\":\"user-id-0\",\"data\":{\"value\":\"0\",\"type\":\"number\",\"format\":\"int\"}}"

	jsonstr1 := "{\"key\":\"user-id-0\",\"type\":2,\"label\":\"捐赠证书pdf\",\"data\":{\"value\":\"0\",\"type\":\"number\",\"format\":\"int\"}}"

	//k:[v1,v2]
	jsonstr2 := "{\"key\": \"user-id-1\",\"data\": [{\"value\": \"0\",\"type\": \"number\",\"format\": \"int32\"}, {\"value\": \"2\",\"type\": \"number\",\"format\": \"int64\"}]}"

	//k:{k:{k:[v1,v2]}}
	jsonstr3 := "{\"key\": \"user-id-2\",\"data\": {\"key\": \"user-id-3\",\"data\": [{\"value\": \"0\",\"type\": \"number\",\"format\": \"int32\"}, {\"value\": \"2\",\"type\": \"number\",\"format\": \"int64\"}]}}"

	//k:[k:v,k:v,v]
	jsonstr4 := "{\"key\": \"user-id-4\",\"data\": [{\"key\": \"user-id-5\",\"data\": {\"value\": \"1\",\"type\": \"number\",\"format\": \"int32\"}}, {\"key\": \"user-id-6\",\"data\": {\"value\": \"2\",\"type\": \"number\",\"format\": \"int32\"}}]}"

	jsonstr5 := "{\"key\": \"ref_hashes\",\"data\": [{\"value\": \"0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f\",\"type\": \"file\",\"format\": \"hash\"}, {\"value\": \"0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56\",\"type\": \"file\",\"format\": \"hash\"}]}"

	jsonstr := "[" + jsonstr1 + "," + jsonstr2 + "," + jsonstr3 + "," + jsonstr4 + "," + jsonstr5 + "]"

	xx := base64.StdEncoding.EncodeToString([]byte(jsonstr))

	var payload string
	//payload 不加密
	//1.option字段没有
	if flag == 1 {
		payload = "{\"version\":\"1.0.0\",\"data\":\"" + xx + "\"}"
	} else if flag == 2 {
		//2.version字段没有
		optionStr := `{"archive":"zip"}`
		enOptionStr := dbcom.EncodeToString([]byte(optionStr))
		payload = "{\"option\":\"" + enOptionStr + "\",\"data\":\"" + xx + "\"}"
	} else if flag == 3 {
		//3.version字段填写空
		optionStr := `{"archive":"zip"}`
		enOptionStr := dbcom.EncodeToString([]byte(optionStr))
		payload = "{\"version\":\"\",\"option\":\"" + enOptionStr + "\",\"data\":\"" + xx + "\"}"

	} else if flag == 4 {
		//4.version字段大于1.0.0
		optionStr := `{"archive":"zip"}`
		enOptionStr := dbcom.EncodeToString([]byte(optionStr))
		payload = "{\"version\":\"1.1.1\",\"option\":\"" + enOptionStr + "\",\"data\":\"" + xx + "\"}"

	}

	return payload
}
func testProofNestedCombination(t *testing.T, payload string) {
	tx := &types.Transaction{Execer: []byte("user.p.sanhe.proof"), Payload: []byte(payload)}
	tx.To = db.ExecAddress("user.p.sanhe.proof")
	priv := HexToPrivkey(Priv)
	if priv != nil {
		tx.Sign(types.SECP256K1, priv)
	}

	//构建区块
	newblock := &types.Block{}
	newblock.Height = 1
	newblock.BlockTime = types.Now().Unix()
	newblock.ParentHash = zeroHash[:]
	newblock.Txs = []*types.Transaction{tx}
	newblock.TxHash = zeroHash[:]

	env := db.TxEnv{
		TxIndex: 0,
		Block:   &types.BlockDetail{},
	}
	env.Block.Block = newblock
	rec := types.ReceiptData{Ty: types.ExecPack}
	env.Block.Receipts = append(env.Block.Receipts, &rec)

	convert := NewConvert("sanhe", "bty", "proof")
	c := convert.(*ProofConvert)
	c.RecordGen = NewRecordGen("p1", "p2", "l1", "l2", "t1", "t2", "u1", "u2")
	c.configDB = &proofconfig.None{}

	records, err := convert.ConvertTx(&env, db.SeqTypeAdd)

	assert.Nil(t, err)
	for _, r := range records {
		t.Log("Test_ProofNestedCombination", "jsonstr", string(r.Value()))
	}
	re := records[0].(*ProofRecord)
	t.Log("Test_Proof", "ikey", re.Key())
	t.Log("Test_Proof", "OpType", re.OpType())

	pInfo := re.Proof.Proof

	assert.Equal(t, "fuzamei", pInfo["proof_organization"])
	assert.Equal(t, "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", pInfo["proof_sender"])
	assert.Equal(t, int64(100000), pInfo["proof_height_index"])

	if !proofUtil.IsParserToKV {
		assert.Equal(t, 0, pInfo["user-id-0"])

		assert.Equal(t, int64(0), pInfo["user-id-1"].([]interface{})[0])
		assert.Equal(t, int64(2), pInfo["user-id-1"].([]interface{})[1])

		assert.Equal(t, int64(0), pInfo["user-id-2"].(map[string]interface{})["user-id-3"].([]interface{})[0])
		assert.Equal(t, int64(2), pInfo["user-id-2"].(map[string]interface{})["user-id-3"].([]interface{})[1])

		assert.Equal(t, int64(1), pInfo["user-id-4"].([]interface{})[0].(map[string]interface{})["user-id-5"])
		assert.Equal(t, int64(2), pInfo["user-id-4"].([]interface{})[1].(map[string]interface{})["user-id-6"])

		assert.Equal(t, "0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f", pInfo["ref_hashes"].([]interface{})[0])
		assert.Equal(t, "0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56", pInfo["ref_hashes"].([]interface{})[1])
	} else {
		assert.Equal(t, 0, pInfo["user-id-0"])

		assert.Equal(t, int64(0), pInfo["user-id-1"].([]interface{})[0])
		assert.Equal(t, int64(2), pInfo["user-id-1"].([]interface{})[1])

		assert.Equal(t, int64(0), pInfo["user-id-3"].([]interface{})[0])
		assert.Equal(t, int64(2), pInfo["user-id-3"].([]interface{})[1])

		assert.Equal(t, int64(1), pInfo["user-id-5"])
		assert.Equal(t, int64(2), pInfo["user-id-6"])

		assert.Equal(t, "0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f", pInfo["ref_hashes"].([]interface{})[0])
		assert.Equal(t, "0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56", pInfo["ref_hashes"].([]interface{})[1])

	}
}

func decodeProof(t *testing.T, x interface{}) (interface{}, error) {
	tempjsonstr := make(map[string]interface{})
	pInfo := make(map[string]interface{})
	err := json.Unmarshal(x.([]byte), &pInfo)
	if err != nil {
		log.Info("decodeProof-Unmarshal", "err", err)
		return x, err
	}

	assert.Equal(t, "fuzamei", pInfo["proof_organization"])
	assert.Equal(t, "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", pInfo["proof_sender"])
	assert.Equal(t, float64(100000), pInfo["proof_height_index"])

	assert.Equal(t, float64(1136214245), pInfo["user-id-0"])

	assert.Equal(t, 0.123456789, pInfo["user-id-1"].([]interface{})[0])
	assert.Equal(t, float64(2), pInfo["user-id-1"].([]interface{})[1])

	assert.Equal(t, float64(1), pInfo["user-id-2"].(map[string]interface{})["user-id-3"])

	assert.Equal(t, float64(1), pInfo["user-id-4"].([]interface{})[0].(map[string]interface{})["user-id-5"])

	assert.Equal(t, float64(2), pInfo["user-id-4"].([]interface{})[1].(map[string]interface{})["user-id-6"])

	assert.Equal(t, "0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f", pInfo["ref_hashes"].([]interface{})[0])
	assert.Equal(t, "0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56", pInfo["ref_hashes"].([]interface{})[1])

	return tempjsonstr, nil
}
func printq(q *querypara.Query) {
	if q.Page != nil {
		log.Info("printq", "Page", q.Page)
	}
	if q.Sort != nil && q.Sort[0] != nil {
		log.Info("printq", "Sort", q.Sort[0])
	}
	if q.Range != nil && q.Range[0] != nil {
		log.Info("printq", "Range", q.Range[0])
	}

	if q.Match != nil && q.Match[0] != nil {
		log.Info("printq", "Match", q.Match[0])
		printq(q.Match[0].SubQuery)
	}

	if q.MatchOne != nil && q.MatchOne[0] != nil {
		log.Info("printq", "MatchOne", q.MatchOne[0])
		printq(q.MatchOne[0].SubQuery)
	}
	if q.MultiMatch != nil && q.MultiMatch[0] != nil {
		log.Info("printq", "MultiMatch", q.MultiMatch[0])
	}
}

type ConfigDB struct {
}

func (c *ConfigDB) IsHaveProofPermission(addr string) bool {
	return true
}
func (c *ConfigDB) IsHaveDelProofPermission(send, proofOrg, proofOwner string) bool {
	return true
}
func (c *ConfigDB) GetOrganizationName(addr string) (string, error) {
	return "fuzamei", nil
}

type myProofDB struct {
}

func (c *myProofDB) GetProof(id string) (*model.Proof, error) {
	p := &model.Proof{
		Proof: make(map[string]interface{}),
	}
	p.Proof["proof_organization"] = "orgamization1"
	p.Proof["proof_sender"] = "address1"
	p.Proof["proof_deleted"] = ""
	p.Proof["proof_deleted_note"] = ""
	p.Proof["proof_tx_hash"] = "id123"
	p.Proof["basehash"] = "null"
	p.Proof["proof_deleted_flag"] = false
	p.Proof["proof_id"] = "proof-id123"
	return p, nil
}

func (c *myProofDB) ListProof(id string) ([]*model.Proof, error) {
	p1 := &model.Proof{
		Proof: make(map[string]interface{}),
	}
	p1.Proof["proof_organization"] = "orgamization1"
	p1.Proof["proof_sender"] = "address1"
	p1.Proof["proof_deleted"] = ""
	p1.Proof["proof_deleted_note"] = ""
	p1.Proof["proof_tx_hash"] = "id124"
	p1.Proof["proof_id"] = "proof-id124"
	p1.Proof["basehash"] = "id123"
	p1.Proof["proof_deleted_flag"] = false

	p2 := &model.Proof{
		Proof: make(map[string]interface{}),
	}
	p2.Proof["proof_organization"] = "orgamization1"
	p2.Proof["proof_sender"] = "address1"
	p2.Proof["proof_deleted"] = ""
	p2.Proof["proof_deleted_note"] = ""
	p2.Proof["proof_tx_hash"] = "id125"
	p2.Proof["basehash"] = "id123"
	p2.Proof["proof_id"] = "proof-id125"
	p2.Proof["proof_deleted_flag"] = false
	return []*model.Proof{p1, p2}, nil
}

func (c *myProofDB) GetProofLog(id string) (*model.Log, error) {
	return &model.Log{
		ID: id,
	}, nil
}

func (c *myProofDB) GetTemplate(id string) (*model.Template, error) {
	p := &model.Template{
		Template: make(map[string]interface{}),
	}
	p.Template["template_name"] = "test opt"
	p.Template["template_data"] = "[{\"label\":\"相册\",\"key\":\"\",\"type\":3,\"data\":[{\"data\":[{\"type\":\"image\",\"format\":\"hash\",\"value\":\"\"}],\"type\":1,\"key\":\"\",\"label\":\"相册\"},{\"data\":{\"type\":\"text\",\"format\":\"string\",\"value\":\"\"},\"type\":0,\"key\":\"\",\"label\":\"照片描述\"}]},{\"label\":\"ext\",\"key\":\"\",\"type\":3,\"data\":[{\"data\":{\"type\":\"text\",\"format\":\"string\",\"value\":\"\"},\"type\":0,\"key\":\"存证名称\",\"label\":\"存证名称\"},{\"data\":{\"type\":\"text\",\"format\":\"string\",\"value\":\"\"},\"type\":0,\"key\":\"basehash\",\"label\":\"basehash\"},{\"data\":{\"type\":\"text\",\"format\":\"string\",\"value\":\"\"},\"type\":0,\"key\":\"prehash\",\"label\":\"prehash\"},{\"data\":{\"type\":\"text\",\"format\":\"string\",\"value\":\"\"},\"type\":0,\"key\":\"存证类型\",\"label\":\"存证类型\"}]}]"

	p.Template["template_organization"] = "orgamization1"
	p.Template["template_sender"] = "address1"
	p.Template["template_deleted"] = ""
	p.Template["template_deleted_note"] = ""
	p.Template["template_tx_hash"] = "id123"
	p.Template["basehash"] = "null"
	p.Template["template_deleted_flag"] = false
	p.Template["template_id"] = "template-id123"
	return p, nil
}

func (c *myProofDB) GetProofUpdateRecord(updatehash string, version int) (*model.Proof, error) {
	p := &model.Proof{
		Proof: make(map[string]interface{}),
	}
	return p, nil
}

func Test_DeleteProof(t *testing.T) {
	delPayload := "{\"id\": \"id123\",\"note\":\"test delete\"}"
	tx := &types.Transaction{Execer: []byte("user.p.sanhe.proof_delete"), Payload: []byte(delPayload)}
	tx.To = db.ExecAddress("user.p.sanhe.proof_delete")
	priv := HexToPrivkey(Priv)
	if priv != nil {
		tx.Sign(types.SECP256K1, priv)
	}

	//构建区块
	newblock := &types.Block{}
	newblock.Height = 2
	newblock.BlockTime = types.Now().Unix()
	newblock.ParentHash = zeroHash[:]
	newblock.Txs = []*types.Transaction{tx}
	newblock.TxHash = zeroHash[:]

	env := db.TxEnv{
		TxIndex: 0,
		Block:   &types.BlockDetail{},
	}
	env.Block.Block = newblock
	rec := types.ReceiptData{Ty: types.ExecPack}
	env.Block.Receipts = append(env.Block.Receipts, &rec)

	convert := NewConvert("user.p.sanhe.", "bty", "proof")
	c := convert.(*ProofConvert)
	c.RecordGen = NewRecordGen("p1", "p2", "l1", "l2", "t1", "t2", "u1", "u2")
	c.configDB = &proofconfig.None{}
	c.proofDB = &myProofDB{}

	records, err := convert.ConvertTx(&env, db.SeqTypeAdd)

	assert.Nil(t, err)
	assert.Equal(t, 5, len(records))
	for _, r := range records {
		t.Log("Test_DeleteProof", "jsonstr", string(r.Value()))
	}
	r1 := records[0].(*ProofRecord)
	r2 := records[1].(*ProofRecord)
	r3 := records[2].(*ProofRecord)
	l1 := records[3].(*LogRecord)

	assert.Equal(t, true, r1.Proof.Proof["proof_deleted_flag"].(bool))
	assert.Equal(t, true, r2.Proof.Proof["proof_deleted_flag"].(bool))
	assert.Equal(t, true, r3.Proof.Proof["proof_deleted_flag"].(bool))
	assert.Equal(t, "delete", l1.Log.Op)
	assert.Equal(t, db.OpUpdate, r1.Op.OpType())
	assert.Equal(t, db.OpUpdate, r2.Op.OpType())
	assert.Equal(t, db.OpUpdate, r3.Op.OpType())
}

func Test_ForceDeleteProof(t *testing.T) {
	delPayload := "{\"id\": \"id123\",\"note\":\"test delete\", \"force\":true}"
	tx := &types.Transaction{Execer: []byte("user.p.sanhe.proof_delete"), Payload: []byte(delPayload)}
	tx.To = db.ExecAddress("user.p.sanhe.proof_delete")
	priv := HexToPrivkey(Priv)
	if priv != nil {
		tx.Sign(types.SECP256K1, priv)
	}

	//构建区块
	newblock := &types.Block{}
	newblock.Height = 2
	newblock.BlockTime = types.Now().Unix()
	newblock.ParentHash = zeroHash[:]
	newblock.Txs = []*types.Transaction{tx}
	newblock.TxHash = zeroHash[:]

	env := db.TxEnv{
		TxIndex: 0,
		Block:   &types.BlockDetail{},
	}
	env.Block.Block = newblock
	rec := types.ReceiptData{Ty: types.ExecPack}
	env.Block.Receipts = append(env.Block.Receipts, &rec)

	convert := NewConvert("user.p.sanhe.", "bty", "proof")
	c := convert.(*ProofConvert)
	c.RecordGen = NewRecordGen("p1", "p2", "l1", "l2", "t1", "t2", "u1", "u2")
	c.configDB = &proofconfig.None{}
	c.proofDB = &myProofDB{}

	records, err := convert.ConvertTx(&env, db.SeqTypeAdd)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(records))
	for _, r := range records {
		t.Log("Test_ForceDeleteProof", "jsonstr", string(r.Value()))
	}
	r1 := records[0].(*ProofRecord)
	r2 := records[1].(*ProofRecord)
	r3 := records[2].(*ProofRecord)
	l1 := records[3].(*LogRecord)

	assert.Equal(t, true, r1.Proof.Proof["proof_deleted_flag"].(bool))
	assert.Equal(t, true, r2.Proof.Proof["proof_deleted_flag"].(bool))
	assert.Equal(t, true, r3.Proof.Proof["proof_deleted_flag"].(bool))
	assert.Equal(t, "delete", l1.Log.Op)
	assert.Equal(t, db.OpDel, r1.Op.OpType())
	assert.Equal(t, db.OpDel, r2.Op.OpType())
	assert.Equal(t, db.OpDel, r3.Op.OpType())
}
func Test_Encrypt(t *testing.T) {

	jsonstr := "{\"key\": \"user-id-1\",\"data\": [{\"value\": \"0\",\"type\": \"number\",\"format\": \"int32\"}, {\"value\": \"2\",\"type\": \"number\",\"format\": \"int64\"}]}"

	jsonstr1 := "k"
	t.Log("Test_Encrypt", "jsonstr", jsonstr)

	privkey, pubkey := dbcom.CreateKeys()
	t.Log("Test_Encrypt", "privkey", privkey, "pubkey", pubkey)

	rsa, err := dbcom.NewXRsa(privkey, pubkey)
	assert.Nil(t, err)

	encryptstr, err := rsa.PublicEncrypt(jsonstr)
	assert.Nil(t, err)
	t.Log("Test_Encrypt:rsa.PublicEncrypt", "encryptstr", encryptstr)

	decryptstr, err := rsa.PrivateDecrypt(encryptstr)
	assert.Nil(t, err)
	t.Log("Test_Encrypt:rsa.PrivateDecrypt", "decryptstr", decryptstr)

	assert.Equal(t, decryptstr, jsonstr)

	//pub和priv分开解密和加密
	enstr, err := dbcom.PublicEncrypt(pubkey, jsonstr1)
	assert.Nil(t, err)
	t.Log("Test_Encrypt:PublicEncrypt", "enstr", enstr)

	destr, err := dbcom.PrivateDecrypt(privkey, enstr)
	assert.Nil(t, err)
	t.Log("Test_Encrypt:PrivateDecrypt", "destr", destr)

	assert.Equal(t, destr, jsonstr1)

}

func Test_ProofV3(t *testing.T) {
	jsonstr := "{\"ext\":{\"basehash\":\"null\",\"prehash\":\"null\",\"存证名称\":\"相册\",\"存证类型\":\"相册\"},\"相册\":{\"照片描述\":\"\",\"相册\":[\"4bb42ecfe15a99ee586031b7a5a20a5d15c4364780f5ea14d278095d620619f5\",\"b45f6c7728e9b59f9ad381f36c29945317ba646e49ad3b73511ec1731c22dff4\"]}}"
	//jsonstr := "[" + jsonstr5 + "]"

	//base编码
	xx := dbcom.EncodeToString([]byte(jsonstr))

	optionStr := `[{"key":"template","value":"id123"}]`
	enOptionStr := dbcom.EncodeToString([]byte(optionStr))

	notestr := "{\"userName\":\"18701358726\",\"userIcon\":\"\",\"evidenceName\":\"公益捐赠yyh\",\"stepName\":\"\",\"version\":1}"
	enNoteStr := dbcom.EncodeToString([]byte(notestr))

	payload := "{\"version\":\"V3.0.0\",\"option\":\"" + enOptionStr + "\",\"data\":\"" + xx + "\",\"note\":\"" + enNoteStr + "\"}"

	t.Log("Test_Proof", "payload", payload)

	var payload2 api.ProofInfo
	err := json.Unmarshal([]byte(payload), &payload2)
	if err != nil {
		t.Log("Test_Proof1111err", "err", err)

		return
	}

	yy, _ := dbcom.DecodeString(payload2.Data)
	t.Log("Test_Proof1111", "payload.data", string(yy))

	t.Log("Test_Proof1111", "payload.Version", payload2.Version)

	t.Log("Test_Proof1111", "payload.Option", payload2.Option)

	tx := &types.Transaction{Execer: []byte("user.p.testproof.proof"), Payload: []byte(payload)}
	tx.To = db.ExecAddress("user.p.testproof.proof")
	priv := HexToPrivkey(Priv)
	if priv != nil {
		tx.Sign(types.SECP256K1, priv)
	}

	//构建区块
	newblock := &types.Block{}
	newblock.Height = 1
	newblock.BlockTime = types.Now().Unix()
	newblock.ParentHash = zeroHash[:]
	newblock.Txs = []*types.Transaction{tx}
	newblock.TxHash = zeroHash[:]
	env := db.TxEnv{
		TxIndex:   0,
		Block:     &types.BlockDetail{},
		BlockHash: common.ToHex(newblock.HashByForkHeight(0)),
	}
	blockTime := newblock.BlockTime
	blockHash := env.BlockHash
	env.Block.Block = newblock
	rec := types.ReceiptData{Ty: types.ExecPack}
	env.Block.Receipts = append(env.Block.Receipts, &rec)

	convert := NewConvert("user.p.testproof.", "bty", "proof")
	c := convert.(*ProofConvert)
	c.RecordGen = NewRecordGen("p1", "p2", "l1", "l2", "t1", "t2", "u1", "u2")
	c.configDB = &proofconfig.None{}
	c.proofDB = &myProofDB{}
	records, err := convert.ConvertTx(&env, db.SeqTypeAdd)

	assert.Nil(t, err)
	for i, r := range records {
		t.Log(string(r.Value()))
		if i == 0 && !proofUtil.IsParserToKV {
			decodeProof(t, r.Value())
		}
	}

	re := records[0].(*ProofRecord)
	t.Log("Test_Proof", "ikey", re.Key())
	t.Log("Test_Proof", "OpType", re.OpType())

	pInfo := re.Proof.Proof
	t.Log("Test_Proof", "pInfo", pInfo)

	assert.Equal(t, common.ToHex(tx.Hash()), pInfo["proof_tx_hash"])
	assert.Equal(t, fmt.Sprintf("%s-%s", model.ProofID, common.ToHex(tx.Hash())), pInfo["proof_id"])

	assert.Equal(t, int64(1), pInfo["proof_height"])
	assert.Equal(t, blockTime, pInfo["proof_block_time"])
	assert.Equal(t, blockHash, pInfo["proof_block_hash"])

	assert.Equal(t, "fuzamei", pInfo["proof_organization"])
	assert.Equal(t, "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", pInfo["proof_sender"])
	assert.Equal(t, int64(100000), pInfo["proof_height_index"])
	assert.Equal(t, notestr, pInfo["proof_note"])

	//if !proofUtil.IsParserToKV {
	//	assert.Equal(t, int64(1136214245), pInfo["user-id-0"])
	//	assert.Equal(t, 0.123456789, pInfo["user-id-1"].([]interface{})[0])
	//	assert.Equal(t, int64(2), pInfo["user-id-1"].([]interface{})[1])
	//
	//	assert.Equal(t, int64(1), pInfo["user-id-2"].(map[string]interface{})["user-id-3"])
	//
	//	assert.Equal(t, int64(1), pInfo["user-id-4"].([]interface{})[0].(map[string]interface{})["user-id-5"])
	//
	//	assert.Equal(t, int64(2), pInfo["user-id-4"].([]interface{})[1].(map[string]interface{})["user-id-6"])
	//
	//	assert.Equal(t, "0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f", pInfo["ref_hashes"].([]interface{})[0])
	//	assert.Equal(t, "0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56", pInfo["ref_hashes"].([]interface{})[1])
	//} else {
	//	assert.Equal(t, int64(1136214245), pInfo["user-id-0"])
	//	assert.Equal(t, 0.123456789, pInfo["user-id-1"].([]interface{})[0])
	//	assert.Equal(t, int64(2), pInfo["user-id-1"].([]interface{})[1])
	//
	//	assert.Equal(t, int64(1), pInfo["user-id-3"])
	//
	//	assert.Equal(t, int64(1), pInfo["user-id-5"])
	//
	//	assert.Equal(t, int64(2), pInfo["user-id-6"])
	//
	//	assert.Equal(t, "0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f", pInfo["ref_hashes"].([]interface{})[0])
	//	assert.Equal(t, "0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56", pInfo["ref_hashes"].([]interface{})[1])
	//
	//}
}

func TestProofConvert_AddProof(t *testing.T) {
	jsonstr := "{\"ext\":{\"basehash\":\"null\",\"prehash\":\"null\",\"存证名称\":\"相册\",\"存证类型\":\"相册\"},\"相册\":{\"照片描述\":\"\",\"相册\":[\"4bb42ecfe15a99ee586031b7a5a20a5d15c4364780f5ea14d278095d620619f5\",\"b45f6c7728e9b59f9ad381f36c29945317ba646e49ad3b73511ec1731c22dff4\"]}}"
	//base编码
	xx := dbcom.EncodeToString([]byte(jsonstr))

	optionStr := `[{"key":"template","value":"id123"}]`
	enOptionStr := dbcom.EncodeToString([]byte(optionStr))

	noteStr := "{\"userName\":\"18701358726\",\"userIcon\":\"\",\"evidenceName\":\"公益捐赠yyh\",\"stepName\":\"\",\"version\":1}"
	enNoteStr := dbcom.EncodeToString([]byte(noteStr))

	extStr := "{\"update_hash\":\"id123\",\"source_hash\": \"[\\\"0xbff42fdbf6bb6c0461eead765f06d89cedc47d6bc0a06efd068e192d2ddf4e7c\\\",\\\"0xa536fdb9e332d7c1525ce022200f9498c487d0b1effc5e6a37fd3943caef36fa\\\"]\"}"
	enExtStr := dbcom.EncodeToString([]byte(extStr))

	payload := "{\"version\":\"V3.0.0\",\"option\":\"" + enOptionStr + "\",\"data\":\"" + xx + "\",\"note\":\"" + enNoteStr + "\",\"ext\":\"" + enExtStr + "\"}"

	t.Log("Test_Proof", "payload", payload)

	var payload2 api.ProofInfo
	err := json.Unmarshal([]byte(payload), &payload2)
	if err != nil {
		t.Log("Test_Proof1111err", "err", err)

		return
	}

	yy, _ := dbcom.DecodeString(payload2.Data)
	t.Log("Test_Proof1111", "payload.data", string(yy))
	t.Log("Test_Proof1111", "payload.Version", payload2.Version)
	t.Log("Test_Proof1111", "payload.Option", payload2.Option)
	t.Log("Test_Proof1111", "payload.Update", payload2.Ext)

	tx := &types.Transaction{Execer: []byte("user.p.testproof.proof"), Payload: []byte(payload)}
	tx.To = db.ExecAddress("user.p.testproof.proof")
	priv := HexToPrivkey(Priv)
	if priv != nil {
		tx.Sign(types.SECP256K1, priv)
	}

	//构建区块
	newblock := &types.Block{}
	newblock.Height = 1
	newblock.BlockTime = types.Now().Unix()
	newblock.ParentHash = zeroHash[:]
	newblock.Txs = []*types.Transaction{tx}
	newblock.TxHash = zeroHash[:]

	env := db.TxEnv{
		TxIndex:   0,
		Block:     &types.BlockDetail{},
		BlockHash: common.ToHex(newblock.HashByForkHeight(0)),
	}
	env.Block.Block = newblock
	rec := types.ReceiptData{Ty: types.ExecPack}
	env.Block.Receipts = append(env.Block.Receipts, &rec)

	convert := NewConvert("user.p.testproof.", "bty", "proof")
	c := convert.(*ProofConvert)
	c.RecordGen = NewRecordGen("p1", "p2", "l1", "l2", "t1", "t2", "u1", "u2")
	c.configDB = &proofconfig.None{}
	c.proofDB = &myProofDB{}

	db.SetVersion(6)

	records, err := convert.ConvertTx(&env, db.SeqTypeAdd)
	c.AddProof(&env, db.SeqTypeAdd)

	assert.Nil(t, err)
	for i, r := range records {
		t.Log(string(r.Value()))
		if i == 0 && !proofUtil.IsParserToKV {
			decodeProof(t, r.Value())
		}
	}

	re1 := records[0].(*ProofRecord)
	t.Log("Test_Proof", "ikey", re1.Key())
	t.Log("Test_Proof", "OpType", re1.OpType())

	pInfo1 := re1.Proof.Proof
	t.Log("Test_Proof", "pInfo", pInfo1)
	for i, i2 := range pInfo1 {
		t.Log(i, ":", i2)
	}

	re2 := records[1].(*ProofRecord)
	t.Log("Test_Proof", "ikey", re2.Key())
	t.Log("Test_Proof", "OpType", re2.OpType())

	pInfo2 := re2.Proof.Proof
	t.Log("Test_Proof", "pInfo", pInfo2)
	for i, i2 := range pInfo2 {
		t.Log(i, ":", i2)
	}
}
