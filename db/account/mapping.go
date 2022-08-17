package account

// AccountMapping mapping
const AccountMapping = `{
    "mappings":{
        "properties":{
            "type":{
                "type":"keyword"
            },
            "frozen":{
                "type":"long"
            },
            "total":{
                "type":"long"
            },
            "address":{
                "type":"keyword"
            },
            "asset_symbol":{
                "type":"keyword"
            },
            "exec":{
                "type":"keyword"
            },
            "balance":{
                "type":"long"
            },
            "height_index":{
                "type":"long"
            },
            "asset_exec":{
                "type":"keyword"
            },
            "height":{
                "type":"long"
            },
            "block_time":{
                "type":"long"
            },
            "addr_type":{
                "type":"keyword"
            }
        }
    }
}`
