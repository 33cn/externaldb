package db

// TokenMapping mapping
// nolint
const TokenMapping = `{
    "mappings":{
        "properties":{
            "goods_id":{
                "type":"long"
            },
            "name":{
                "type":"text"
            },
            "owner":{
                "type":"keyword"
            },
            "publisher":{
                "type":"keyword"
            },
            "label_id":{
                "type":"keyword"
            },
            "goods_type":{
                "type":"integer"
            },
            "amount":{
                "type":"long"
            },
            "trace_hash":{
                "type":"text"
            },
            "publish_time":{
                "type":"long"
            },
            "contract_addr":{
               	"type":"keyword"
            },
            "source_hash":{
                "type":"text"
            },
            "remark":{
                "type":"text"
            },
            "call_func_name":{
                "type":"keyword"
            },
			"evm_height_index":{
                "type":"long"
            },
            "evm_tx_hash":{
                "type":"keyword"
            },
            "evm_height":{
                "type":"long"
            },
            "evm_block_time":{
                "type":"long"
            },
            "evm_block_hash":{
                "type":"keyword"
            }
        }
    }
}`

const TransferMapping = `{
    "mappings":{
        "properties":{
            "from":{
                "type":"keyword"
            },
            "to":{
                "type":"keyword"
            },
            "goods_id":{
                "type":"keyword"
            },
            "amount":{
                "type":"long"
            },
            "data":{
                "type":"text"
            }
        }
    }
}`

const AccountMapping = `{
	"mappings":{
		"properties":{
			"owner_addr":{
                "type": "keyword"
            },
            "contract_addr":{
                "type": "keyword"
            },
            "label_id":{
                "type":"keyword"
            },
            "goods_id":{
                "type": "long"
            },
            "balance":{
                "type": "long"
            }
		}
	}
}`
