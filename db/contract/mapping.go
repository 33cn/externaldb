package contract

// Mapping 合约验证表
const Mapping = `{
	 "mappings":{
        "properties":{
            "contract_address":{
                "type":"keyword"
            },
            "creator":{
                "type":"keyword"
            },
            "deploy_block_hash":{
                "type":"keyword"
            },
			"deploy_block_time":{
                "type":"long"
            },
			"deploy_height": {
				"type":"long"
			},
			"deploy_height_index": {
				"type":"long"
			},
			"deploy_tx_hash":{
                "type":"keyword"
            },
            "contract_type":{
                "type":"text"
            },
            "contract_abi":{
                "type":"text"
            },
            "contract_bin":{
                "type":"text"
            },
            "contract_bin_hash":{
                "type":"keyword"
            },
			"tx_count":{
                "type":"long"
            },
            "publish_count":{
                "type":"long"
            }
        }
    }
}`
