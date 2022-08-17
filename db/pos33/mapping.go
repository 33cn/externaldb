package pos33

// Pos33TicketMapping Pos33TicketMapping
const Pos33TicketMapping = `{
    "mappings":{
        "properties":{
            "close_at":{
                "properties":{
                    "ts":{
                        "type":"long"
                    },
                    "height":{
                        "type":"long"
                    }
                }
            },
            "status":{
                "type":"keyword"
            },
            "miner_at":{
                "properties":{
                    "height":{
                        "type":"long"
                    },
                    "ts":{
                        "type":"long"
                    }
                }
            },
            "account":{
                "type":"long"
            },
            "open_at":{
                "properties":{
                    "height":{
                        "type":"long"
                    },
                    "ts":{
                        "type":"long"
                    }
                }
            },
            "owner":{
                "type":"keyword"
            },
            "miner":{
                "type":"keyword"
            }
        }
    }
}`

// BindMapping BindMapping
const BindMapping = `{
    "mappings":{
        "properties":{
            "return_address":{
                "type":"keyword"
            },
            "old_miner":{
                "type":"keyword"
            },
            "new_miner":{
                "type":"keyword"
            },
            "ts":{
                "type":"long"
            },
            "height":{
                "type":"long"
            }
        }
    }
}`
