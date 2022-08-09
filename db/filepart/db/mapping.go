package db

// FilePartMapping 文件分片表
const FilePartMapping = `{
	 "mappings":{
        "properties":{
  			"data":{
                "type":"binary"
            },
			"tx_hash":{
                "type":"keyword"
            }
        }
    }
}`
