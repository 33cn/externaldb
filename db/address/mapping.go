package address

// Mapping 合约验证表
const Mapping = `{
	 "mappings":{
        "properties":{
            "address":{
                "type":"keyword"
            },
            "addr_type":{
                "type":"text"
            },
			"tx_count":{
                "type":"long"
            }
        }
    }
}`
