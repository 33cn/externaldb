package contractverify

// Mapping 合约验证表
const Mapping = `{
	 "mappings":{
        "properties":{
            "contract_bin_hash":{
                "type":"keyword"
            },
            "contract_bin":{
                "type":"text"
            },
            "contract_abi":{
                "type":"text"
            },
            "contract_type":{
                "type":"text"
            },
            "compile_type":{
                "type":"text"
            },
            "compile_version":{
                "type":"text"
            }
        }
    }
}`
