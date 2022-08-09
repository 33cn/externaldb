package blockinfo

// Mapping block mapping
const Mapping = `{
    "mappings":{
        "properties":{
            "hash":{
                "type":"keyword"
            },
            "height":{
                "type":"long"
            },
            "tx_count":{
                "type":"long"
            },
            "block_time":{
                "type":"long"
            },
            "from":{
                "type":"keyword"
            }
        }
    }
}`
