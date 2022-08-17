package block

// Mapping block mapping
const Mapping = `{
    "mappings":{
        "properties":{
            "from":{
                "type":"keyword"
            },
            "block_detall":{
                "type":"binary"
            },
            "hash":{
                "type":"keyword"
            },
            "sync_seq":{
                "type":"long"
            },
            "number":{
                "type":"long"
            },
            "type":{
                "type":"long"
            }
        }
    }
}`
