package db

// convert seq
const (
	//修改Mapping去掉了Type
	SeqMapping = `{
    "mappings":{
        "properties":{
            "sync_seq":{
                "type":"long"
            }
        }
    }
}`

	LastSeqDB   = "last_seq"
	DefaultType = "_doc"
)
