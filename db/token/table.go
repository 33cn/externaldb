package token

// mapping
const (
	TxRecordMapping = `{
    "mappings":{
        "properties":{
            "from":{
                "type":"keyword"
            },
            "to":{
                "type":"keyword"
            },
            "amount":{
                "type":"long"
            },
            "symbol":{
                "type":"keyword"
            },
            "action_type":{
                "type":"keyword"
            },
            "tx_hash":{
                "type":"keyword"
            },
            "height":{
                "type":"long"
            },
            "index":{
                "type":"long"
            }
        }
    }
}`

	BalanceRecordMapping = `{
    "mappings":{
        "properties":{
            "balance":{
                "type":"long"
            },
            "frozen":{
                "type":"long"
            }
        }
    }
}`

	InfoRecordMapping = `{
    "mappings":{
        "properties":{
            "name":{
                "type":"keyword"
            },
            "symbol":{
                "type":"keyword"
            },
            "amount":{
                "type":"long"
            },
            "owner":{
                "type":"keyword"
            },
            "creator":{
                "type":"keyword"
            },
            "introduction":{
                "type":"keyword"
            },
            "price":{
                "type":"long"
            },
            "category":{
                "type":"long"
            },
            "status":{
                "type":"long"
            },
            "prepare_height":{
                "type":"long"
            },
            "create_height":{
                "type":"long"
            },
            "revoke_height":{
                "type":"long"
            }
        }
    }
}`
)
