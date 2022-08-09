用于快速部署，在修改了物料表上的相关配置后在可一键启动程序，目前适用于用es方式进行同步，可选择es版本为6或者7的方式，选择mq存证同步数据后期陆续再添加上

### 1. 软件包的生成：
##### （1）make build：编译运行程序
##### （2）make pkg-quickStart：生成用于快速启动的程序包
### 2. 运行程序：
##### （1）在对应服务器上解压程序包
##### （2）修改 .env文件，.env文件中有如下配置:
- 展开服务类型选择：分为存证”proof“和区块链浏览器”browser“和通用"common"
- ES的版本选择，默认为7
- 展开服务的docker-compose.yaml有关容器名的配置
- ES的docker-compose.yaml相关配置，对配置es的密码
- externaldb.toml和proof-init.bash文件中需根据**物料表**修改的相关配置
- externaldb.toml中es的prefix，节点推送和rpc_host相关配置

##### （3）运行程序：（推荐分开启动，若连接到已部署的ES上则选择分开启动）

**A. 一键启动**
- make quickStart: 快速运行，包括启动es、初始化配置文件、启动展开服务

**B. 分开启动**
- make initES：先启动ES（若连接到已有ES上，则忽略此步骤）
- make initConfig：再初始化相关配置，包括对配置文件的修改和运行proof-init.sh将初始数据写入ES
- make run：最后，运行展开服务
- 其他：
    - make down：关闭并移除展开服务；
    - make downES：关闭并移除es容器；
    - make restart：重启展开服务；
    - make stop：展开服务停止运行；
    - make rm：删除展开服务容器；
    - make logs：查看日志；
    - make logs-rpc：查看rpc日志；
    - make clean：清除展开服务运行日志

### 3. 更新sync和convert模块--重新同步区块数据
（1）修改.env文件中的`PUSH_BIND`、`PUSH_HOST`、`PUSH_NAME`，用于重新注册同步程序；若要修改储存index，则需修改`SYNC_PREFIX`、`CONVERT_PREFIX`，用于生成新的sync index和convert index

（2）执行make initConfig命令：初始化管理员地址；将修改的配置项加载进externaldb.toml文件中

（3）重启sync、convert和rpc

### 4. 只更新convert模块--重新解析区块数据
（1）修改.env文件中的`CONVERT_PREFIX`，用于生成新的convert index储存存证数据

（2）执行make initConfig命令：初始化管理员地址；将修改的配置项加载进externaldb.toml文件中

（3）重启convert和rpc

### 5. 若部署的是浏览器的展开服务
（1）在配置.env时，需要取消对`JRPC_NAME`的注释

（2）取消docker-compose.yaml文件中对jrpc服务的注释，并注释掉rpc服务

（3）其他启动步骤不变