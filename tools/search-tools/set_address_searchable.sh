#!/usr/bin/env bash
curl -XPUT 'http://localhost:9200/coins-tx/_mapping/coins-tx' --header "Content-Type:application/json" -d '
{       
  "properties": {
        "from": {  
            "type": "text",
            "fielddata": true
        },     
        "to": {  
            "type": "text",
            "fielddata": true
        }       
    }         
}'
