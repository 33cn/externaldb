package evm

// EVMMapping evm执行记录表
const EVMMapping = `{
	"mappings":{
        "properties":{
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
            },
			"evm_events":{
				"type":"text"
			}
        }
    }
}`

// TokenMapping mapping
// nolint
const TokenMapping = `{
    "mappings":{
        "properties":{
            "token_id":{
                "type":"keyword"
            },
            "owner":{
                "type":"keyword"
            },
            "token_type":{
                "type":"text"
            },
            "amount":{
                "type":"long"
            },
            "contract_addr":{
               	"type":"keyword"
            },
 			"publish_address":{
                "type":"keyword"
            },
 			"publish_tx_hash":{
                "type":"keyword"
            },
       		"publish_height":{
                "type":"long"
            },
			"publish_height_index":{
                "type":"long"
            },
 			"publish_block_hash":{
                "type":"keyword"
            },
            "publish_block_time":{
                "type":"long"
            }
        }
    }
}`

const EVMTransferMapping = `{
    "mappings":{
        "properties":{
            "token_id":{
                "type":"keyword"
            },
            "from":{
                "type":"keyword"
            },
  			"to":{
                "type":"keyword"
            },
			"operator":{
                "type":"keyword"
            },
            "token_type":{
                "type":"text"
            },
            "amount":{
                "type":"long"
            },
            "contract_addr":{
               	"type":"keyword"
            },
 			"tx_hash":{
                "type":"keyword"
            },
       		"height":{
                "type":"long"
            },
			"height_index":{
                "type":"long"
            },
 			"block_hash":{
                "type":"keyword"
            },
            "block_time":{
                "type":"long"
            }
        }
    }
}`
