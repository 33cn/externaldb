
## ci 

```
   1. es docker 先启动
   2. 数据流
   
   dummy node | ----------> | sync | ----> | es | -> | convert | -> |  es | -> | rpcs | <--- test rpc 读到预期的数据
  ------------- push block  --------       ------    -----------               --------                              
   dummy-block
   各种类型的交易

```


## 步骤

 1. make docker 做docker image
 1. docker-composer 脚本组织需要的脚本: 
   * makefile和脚本的界限按: 准备文件分, 在makefile中准备好把文件放到对应位置, 其余交给脚本
```
build/ci
├── ci-setup-config.sh # 各种脚本
├── ci-start-sync.sh
├── configs # 目录 镜像中程序需要的配置
│   ├── chain33.para+(3).toml
│   ├── chain33.toml
│   ├── convert.toml
│   ├── jrpc.toml
│   └── sync.toml
├── datas # ES 数据目录
└── docker-compose.yaml
```
   * tools: dummy-node: 假节点, 做注册和推送区块
   * TODO3: 一些测试判断的脚本
 1. 脚本, 整个ci测试流程的脚本, 配置等
 1. TODO4 本地测试
 1. TODO5 配置到jenkins
 