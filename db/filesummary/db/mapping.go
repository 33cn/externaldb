package db

// FileSummaryMapping 文件汇总表
const FileSummaryMapping = `{
	 "mappings":{
        "properties":{
            "file_type":{
                "type":"keyword"
            },
            "file_hash":{
                "type":"keyword"
            },
            "file_size":{
                "type":"long"
            },
            "part_type":{
                "type":"long"
            },
            "part_hashs":{
                "type":"binary"
            },
            "file_blacklist":{
                "type":"keyword"
            },
            "file_blacklist_flag":{
                "type":"boolean"
            },
            "file_blacklist_note":{
                "type":"keyword"
            }
        }
    }
}`
