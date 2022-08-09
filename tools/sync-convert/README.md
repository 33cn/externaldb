本服务是接收链推送与解析推送数据(即sync与convert)合一版本，合一服务为sync_convert，查询数据为rpc服务

### 1. 软件包的生成：
##### （1）make build：编译运行程序
##### （2）make pkg-mergeStart：打包本服务及其依赖服务
### 2. 运行程序：
##### （1）在对应服务器上解压程序包
##### （2）修改 .env文件，.env文件中有如下配置:
- ES的版本选择，默认为7
- 展开服务的docker-compose.yaml有关容器名的配置
- 启动ES的docker-compose.yaml相关配置
- externaldb.toml中需根据**物料表**修改的相关配置
- externaldb.toml中es的prefix，节点推送和rpc_host相关配置

##### （3）运行程序：（推荐分开启动，若连接到已部署的ES上则选择分开启动）

**A. 一键启动**
- make quickStart: 快速运行，包括启动es、初始化配置文件、启动展开服务

**B. 分开启动**
- make initES：先启动ES（若连接到已有ES上，则忽略此步骤）
- make initConfig：再初始化相关配置
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

##### （5）重新同步数据：
- 修改externaldb.toml 配置文件中的sync子选项下的pushBind，pushHost，pushName，然后重新启动sync_convert服务即可
