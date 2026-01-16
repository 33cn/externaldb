# Config 包单元测试说明

## 测试覆盖率

当前测试覆盖率为 **97.9%**。

## 测试文件结构

```
config/
├── config_test.go      # 配置加载和默认配置测试
├── flags_test.go       # Flag 定义测试
└── loader_test.go      # 配置合并和加载逻辑测试
```

## 测试用例概览

### config_test.go

测试 `config.go` 中的功能：

1. **TestGetDefaultConfig** - 验证默认配置的所有字段
2. **TestLoadConfig_FileNotExists** - 测试配置文件不存在的情况
3. **TestLoadConfig_ValidYAML** - 测试加载有效的 YAML 配置文件
4. **TestLoadConfig_InvalidYAML** - 测试加载无效的 YAML 文件
5. **TestLoadConfig_PartialConfig** - 测试部分配置（只设置部分字段）
6. **TestLoadConfig_EmptyFile** - 测试空配置文件

### flags_test.go

测试 `flags.go` 中的功能：

1. **TestDefineCommonFlags** - 验证通用 flags 的定义
2. **TestDefineScannerFlags** - 验证 Scanner flags 的定义
3. **TestDefineRPCFlags** - 验证 RPC flags 的定义
4. **TestFlags_DefaultValues** - 验证通用 flags 的默认值
5. **TestFlags_ScannerDefaultValues** - 验证 Scanner flags 的默认值
6. **TestFlags_RPCDefaultValues** - 验证 RPC flags 的默认值

### loader_test.go

测试 `loader.go` 中的功能：

1. **TestDetectFlagSet_NoFlags** - 测试没有设置任何 flag 的情况
2. **TestDetectFlagSet_SomeFlags** - 测试设置部分 flags 的情况
3. **TestDetectFlagSet_AllFlags** - 测试设置所有 flags 的情况
4. **TestDetectFlagSet_AlternativeFlagNames** - 测试替代 flag 名称（如 `es-password` vs `es-pwd`）
5. **TestMergeConfigWithFlagsFromStruct** - 测试配置合并功能
6. **TestMergeConfigWithFlagsFromStruct_FlagNotSet** - 测试 flag 未设置时不覆盖配置
7. **TestMergeConfigWithFlagsFromStruct_AllFields** - 测试所有字段的合并
8. **TestMergeConfigWithFlagsFromStruct_EmptyStringFlags** - 测试空字符串 flag 的处理
9. **TestMergeConfigWithFlagsFromStruct_NilFlags** - 测试 nil flag 的处理
10. **TestGetFinalValue** - 测试获取最终配置值（字符串类型）
11. **TestGetFinalIntValue** - 测试获取最终配置值（整数类型）
12. **TestGetFinalBoolValue** - 测试获取最终配置值（布尔类型）
13. **TestLoadAndMergeForScanner_WithConfigFile** - 测试 Scanner 配置加载（使用配置文件）
14. **TestLoadAndMergeForScanner_WithFlags** - 测试 Scanner 配置加载（使用 flags）
15. **TestLoadAndMergeForScanner_ConfigFileError** - 测试 Scanner 配置加载错误处理
16. **TestLoadAndMergeForRPC_WithConfigFile** - 测试 RPC 配置加载（使用配置文件）
17. **TestLoadAndMergeForRPC_WithFlags** - 测试 RPC 配置加载（使用 flags）
18. **TestLoadAndMergeForRPC_DefaultPort** - 测试 RPC 默认端口
19. **TestLoadAndMergeForRPC_EmptyPortInConfig** - 测试配置文件中 port 为空的情况

## 运行测试

### 运行所有测试

```bash
cd erc20Scaner
go test ./config -v
```

### 运行测试并查看覆盖率

```bash
go test ./config -cover
```

### 生成覆盖率报告

```bash
go test ./config -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### 生成 HTML 覆盖率报告

```bash
go test ./config -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## 测试策略

### 1. 配置加载测试

- ✅ 测试配置文件存在和不存在的情况
- ✅ 测试有效和无效的 YAML 格式
- ✅ 测试部分配置和完整配置
- ✅ 测试空配置文件

### 2. Flag 定义测试

- ✅ 验证所有 flag 结构体字段已初始化
- ✅ 验证默认值正确
- ✅ 测试通用 flags、Scanner flags 和 RPC flags

### 3. 配置合并测试

- ✅ 测试 flag 覆盖配置文件
- ✅ 测试 flag 未设置时使用配置文件
- ✅ 测试空字符串和 nil flag 的处理
- ✅ 测试所有字段的合并

### 4. 辅助函数测试

- ✅ 测试 `GetFinalValue` 的所有分支
- ✅ 测试 `GetFinalIntValue` 的所有分支
- ✅ 测试 `GetFinalBoolValue` 的所有分支

### 5. 集成测试

- ✅ 测试 `LoadAndMergeForScanner` 的完整流程
- ✅ 测试 `LoadAndMergeForRPC` 的完整流程
- ✅ 测试错误处理和默认值回退

## 边界情况测试

1. **空字符串 flag** - 确保空字符串不会覆盖配置值
2. **nil flag** - 确保 nil flag 不会导致 panic
3. **零值 flag** - 确保整数类型的零值正确处理
4. **配置文件错误** - 确保读取错误时使用默认配置
5. **替代 flag 名称** - 测试 `es-password` 和 `chain_grpc` 等替代名称

## 注意事项

1. **Flag 包全局变量**：由于 Go 的 `flag` 包使用全局变量，在测试中需要重置 `flag.CommandLine` 以避免 flag 重复定义错误。

2. **临时文件**：使用 `t.TempDir()` 创建临时目录和文件，测试结束后自动清理。

3. **测试隔离**：每个测试都应该独立运行，不依赖其他测试的状态。

## 持续改进

- 定期运行测试确保代码质量
- 添加新功能时同步添加测试用例
- 保持测试覆盖率在 95% 以上
- 关注边界情况和错误处理

## 示例：添加新测试

当添加新的配置项时，应该添加相应的测试：

```go
func TestNewConfigField(t *testing.T) {
    cfg := GetDefaultConfig()
    
    // 验证新字段的默认值
    if cfg.NewField != expectedDefault {
        t.Errorf("Expected NewField to be %v, got %v", expectedDefault, cfg.NewField)
    }
}
```

