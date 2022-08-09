
 * url:  http://Host:Port/Title
 * request

```

{
        "id" : 1 ,
        "method" : "Account.ListAsset", 
        "params":[{
   		"match" : [
   		   {
   		      "value" : "145HxiUdRSzt49kKA2WoDyjBcYAJB67M46",
   		      "key" : "address"
   		   },
   		   {
   		      "value" : "1BXvgjmBw1aBgmGn1hjfGyRkmN3krWpFP4",
   		      "key" : "exec"
   		   }
   		]
	}
		
        ]
}

```

 * response 

```
{
   "result" : [
      {
         "asset_symbol" : "bty",
         "frozen" : 0,
         "total" : 0,
         "balance" : 0,
         "height_index" : 490600002,
         "exec" : "1BXvgjmBw1aBgmGn1hjfGyRkmN3krWpFP4",
         "address" : "145HxiUdRSzt49kKA2WoDyjBcYAJB67M46",
         "type" : "contractInternal",
         "asset_exec" : "coins"
      },
      {
         "type" : "contractInternal",
         "asset_exec" : "token",
         "address" : "145HxiUdRSzt49kKA2WoDyjBcYAJB67M46",
         "height_index" : 206500002,
         "exec" : "1BXvgjmBw1aBgmGn1hjfGyRkmN3krWpFP4",
         "total" : 675000000,
         "balance" : 675000000,
         "asset_symbol" : "TEST",
         "frozen" : 0
      }
   ],
   "id" : 1,
   "error" : null
}


```
