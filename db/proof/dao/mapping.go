package dao

import (
	"github.com/33cn/externaldb/db"
	"github.com/33cn/externaldb/util"
)

// InitDB init db
func InitDB(cli db.DBCreator, ProofDBX, ProofTableX, LogDBX, LogTableX, TemplateDBX, TemplateTableX, ProofUpdateDBX, ProofUpdateTableX string) error {
	err := util.InitIndex(cli, ProofDBX, ProofTableX, ProofMapping)
	if err != nil {
		return err
	}
	//TODO template mapping
	err = util.InitIndex(cli, TemplateDBX, TemplateTableX, ProofTemplateMapping)
	if err != nil {
		return err
	}

	err = util.InitIndex(cli, ProofUpdateDBX, ProofUpdateTableX, ProofMapping)
	if err != nil {
		return err
	}

	return util.InitIndex(cli, LogDBX, LogTableX, ProofLogMapping)
}

// ProofMapping es proof
//sender: 存证的上传者
//organization：存证属于哪个组织或者公司
//height_index：存证所在的区块高度
//tx_hash：存证所在的交易hash值
//ref_hashes:与本存证相关联的交易hash
//data：存证类容
//delete: 此存证被删除，标记删除此存证的hash
// ProofMapping mapping
const ProofMapping = `{
    "mappings":{
        "date_detection": false,
        "properties":{
            "proof_sender":{
                "type":"keyword"
            },
            "proof_organization":{
                "type":"keyword"
            },
            "proof_height_index":{
                "type":"long"
            },
            "proof_tx_hash":{
                "type":"keyword"
            },
            "proof_ref_hashes":{
                "type":"keyword"
            },
            "proof_data":{
                "type":"binary"
            },
            "proof_deleted":{
                "type":"keyword"
            },
            "proof_deleted_note":{
                "type":"text"
            },
            "proof_id":{
                "type":"keyword"
            },
            "proof_height":{
                "type":"long"
            },
            "proof_block_time":{
                "type":"long"
            },
            "proof_block_hash":{
                "type":"keyword"
            },
            "proof_note":{
                "type":"text"
            },
            "proof_deleted_flag":{
                "type":"boolean"
            }
        }
    }
}`

// ProofLogMapping Proof 删除和恢复的日志
const ProofLogMapping = `{
    "mappings":{
        "properties":{
            "id":{
                "type":"keyword"
            },
            "height":{
                "type":"long"
            },
            "index":{
                "type":"long"
            },
            "proof_hash":{
                "type":"keyword"
            },
            "op":{
                "type":"keyword"
            },
            "note":{
                "type":"text"
            },
            "force":{
                "type":"boolean"
            },
            "address":{
                "type":"keyword"
            },
            "block_time":{
                "type":"long"
            },
            "block_hash":{
                "type":"keyword"
            }
        }
    }
}`

const ProofTemplateMapping = `{
    "mappings":{
        "properties":{
            "template_block_hash":{
                "type":"keyword"
            },
            "template_block_time":{
                "type":"long"
            },
            "template_data":{
                "type":"keyword"
            },
            "template_deleted":{
                "type":"keyword"
            },
            "proof_ref_hashes":{
                "type":"keyword"
            },
            "template_deleted_flag":{
                "type":"boolean"
            },
            "template_deleted_note":{
                "type":"keyword"
            },
            "template_height":{
                "type":"long"
            },
            "template_height_index":{
                "type":"long"
            },
            "template_id":{
                "type":"keyword"
            },
            "template_name":{
                "type":"keyword"
            },
            "template_organization":{
                "type":"keyword"
            },
            "template_sender":{
                "type":"keyword"
            },
            "template_tx_hash":{
                "type":"keyword"
            }
        }
    }
}`
