package block

// mapping
const (
	StatMapping = `{
    	"mappings":{
        	"properties":{
				"coins":{
             	   "type":"long"
           		},
           		"fee":{
					"type":"long"
				},
            	"height":{
                	"type":"long"
            	},
            	"mine":{
                	"type":"long"
            	},
            	"time":{
                	"type":"long"
            	},
            	"tx_count":{
                	"type":"long"
            	}
        	}
    	}
	}`
)
