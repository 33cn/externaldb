package multisig

// db
const (
	msMapping = `{
    "mappings":{
        "properties":{
            "address":{
                "type":"keyword"
            },
            "create_address":{
                "type":"keyword"
            },
            "daily_limit":{
                "type":"long"
            },
            "execer":{
                "type":"keyword"
            },
            "last_day":{
                "type":"long"
            },
            "multi_signature_address":{
                "type":"keyword"
            },
            "required_weight":{
                "type":"long"
            },
            "spent_today":{
                "type":"long"
            },
            "symbol":{
                "type":"keyword"
            },
            "tx_count":{
                "type":"long"
            },
            "weight":{
                "type":"long"
            }
        }
    }
}`

	txMapping = `{
    "mappings":{
        "properties":{
            "multi_signature_address":{
                "type":"keyword"
            },
            "tx_hash":{
                "type":"keyword"
            },
            "tx_id":{
                "type":"long"
            },
            "type":{
                "type":"keyword"
            }
        }
    }
}`

	listMapping = `{
    "mappings":{
        "properties":{
            "address":{
                "type":"keyword"
            },
            "creator":{
                "type":"boolean"
            },
            "executed":{
                "type":"boolean"
            },
            "multi_signature_address":{
                "type":"keyword"
            },
            "tx_hash":{
                "type":"keyword"
            },
            "tx_id":{
                "type":"long"
            },
            "weight":{
                "type":"long"
            }
        }
    }
}`
)
