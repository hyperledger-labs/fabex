**Block explorer for Hyperledger Fabric**

Usage scenario example:                        
(_you can skip any step)_

build:                  
`go build`

enroll user:                         
`./fabex --enrolluser true`

create db table:                                                         
`./fabex --task initdb --config ./config.yaml`

get transactions of specific block (chain operation):                                
`./fabex --task getblock --blocknum 14 --config ./config.yaml`

get all transactions (db operation):                                                                   
`./fabex --task getblock --blocknum 15 --config ./config.yaml`