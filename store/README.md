# seq 存取

原来用es存取seq, 发现seq的使用基本是根据 seq-number 去存取 seq, 更新和读取 last_seq, 没有搜索的要求, 
换成可以kv存储更为合适. 

分步骤实现: 
 1. 操作做成接口, 修改现有实现满足接口
 1. 添加新的存储
 1. 现在是轮训数据库实现, 改成订阅last_seq更新

## 需求

sync 需求: 
 1. 读last_seq  (重启)
 1. 写last_seq (保存新的seq后)
 1. 写seq (保存seq)

convert:
 1. 读last_seq (重启,
 1. 订阅last_seq: 看seq是否更新
 1. 读seq

rpc:
 1. 获得: 读last_seq

## 