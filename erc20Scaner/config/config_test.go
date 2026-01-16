package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()

	if cfg == nil {
		t.Fatal("GetDefaultConfig returned nil")
	}

	// 验证 Node 配置
	if cfg.Node.URL != "https://mainnet.bityuan.com/eth" {
		t.Errorf("Expected Node.URL to be 'https://mainnet.bityuan.com/eth', got '%s'", cfg.Node.URL)
	}
	if cfg.Node.GRPC != "localhost:8802" {
		t.Errorf("Expected Node.GRPC to be 'localhost:8802', got '%s'", cfg.Node.GRPC)
	}
	if cfg.Node.Symbol != "bty" {
		t.Errorf("Expected Node.Symbol to be 'bty', got '%s'", cfg.Node.Symbol)
	}

	// 验证 Scanner 配置
	if cfg.Scanner.StartBlock != 42399544 {
		t.Errorf("Expected Scanner.StartBlock to be 42399544, got %d", cfg.Scanner.StartBlock)
	}
	if cfg.Scanner.EndBlock != -1 {
		t.Errorf("Expected Scanner.EndBlock to be -1, got %d", cfg.Scanner.EndBlock)
	}

	// 验证 RPC 配置
	if cfg.RPC.Port != "8080" {
		t.Errorf("Expected RPC.Port to be '8080', got '%s'", cfg.RPC.Port)
	}

	// 验证 Database 配置
	if cfg.Database.Enabled != false {
		t.Errorf("Expected Database.Enabled to be false, got %v", cfg.Database.Enabled)
	}
	if cfg.Database.DSN == "" {
		t.Error("Expected Database.DSN to be set, got empty string")
	}

	// 验证 ES 配置
	if cfg.ES.Enabled != false {
		t.Errorf("Expected ES.Enabled to be false, got %v", cfg.ES.Enabled)
	}
	if cfg.ES.Host != "http://localhost:9200/" {
		t.Errorf("Expected ES.Host to be 'http://localhost:9200/', got '%s'", cfg.ES.Host)
	}
	if cfg.ES.Prefix != "seq01_" {
		t.Errorf("Expected ES.Prefix to be 'seq01_', got '%s'", cfg.ES.Prefix)
	}
	if cfg.ES.Version != 7 {
		t.Errorf("Expected ES.Version to be 7, got %d", cfg.ES.Version)
	}

	// 验证 Log 配置
	if cfg.Log.Level != "info" {
		t.Errorf("Expected Log.Level to be 'info', got '%s'", cfg.Log.Level)
	}
	if cfg.Log.File != "./logs/app.log" {
		t.Errorf("Expected Log.File to be './logs/app.log', got '%s'", cfg.Log.File)
	}
}

func TestLoadConfig_FileNotExists(t *testing.T) {
	// 使用一个不存在的文件路径
	nonExistentFile := "/tmp/non_existent_config_12345.yaml"
	cfg, err := LoadConfig(nonExistentFile)

	if err != nil {
		t.Fatalf("LoadConfig should not return error for non-existent file, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadConfig should return default config for non-existent file")
	}

	// 验证返回的是默认配置
	if cfg.Node.URL != "https://mainnet.bityuan.com/eth" {
		t.Errorf("Expected default Node.URL, got '%s'", cfg.Node.URL)
	}
}

func TestLoadConfig_ValidYAML(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test_config.yaml")

	yamlContent := `
node:
  url: "https://test.example.com/eth"
  grpc: "test:8802"
  symbol: "test"

scanner:
  start_block: 1000000
  end_block: 2000000

rpc:
  port: "9090"

database:
  enabled: true
  dsn: "user:pass@tcp(localhost:3306)/testdb"

es:
  enabled: true
  host: "http://test-es:9200/"
  prefix: "test_"
  version: 6
  user: "testuser"
  password: "testpass"

log:
  level: "debug"
  file: "/tmp/test.log"
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(configFile)

	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadConfig returned nil config")
	}

	// 验证加载的配置
	if cfg.Node.URL != "https://test.example.com/eth" {
		t.Errorf("Expected Node.URL to be 'https://test.example.com/eth', got '%s'", cfg.Node.URL)
	}
	if cfg.Node.GRPC != "test:8802" {
		t.Errorf("Expected Node.GRPC to be 'test:8802', got '%s'", cfg.Node.GRPC)
	}
	if cfg.Node.Symbol != "test" {
		t.Errorf("Expected Node.Symbol to be 'test', got '%s'", cfg.Node.Symbol)
	}

	if cfg.Scanner.StartBlock != 1000000 {
		t.Errorf("Expected Scanner.StartBlock to be 1000000, got %d", cfg.Scanner.StartBlock)
	}
	if cfg.Scanner.EndBlock != 2000000 {
		t.Errorf("Expected Scanner.EndBlock to be 2000000, got %d", cfg.Scanner.EndBlock)
	}

	if cfg.RPC.Port != "9090" {
		t.Errorf("Expected RPC.Port to be '9090', got '%s'", cfg.RPC.Port)
	}

	if cfg.Database.Enabled != true {
		t.Errorf("Expected Database.Enabled to be true, got %v", cfg.Database.Enabled)
	}
	if cfg.Database.DSN != "user:pass@tcp(localhost:3306)/testdb" {
		t.Errorf("Expected Database.DSN to be 'user:pass@tcp(localhost:3306)/testdb', got '%s'", cfg.Database.DSN)
	}

	if cfg.ES.Enabled != true {
		t.Errorf("Expected ES.Enabled to be true, got %v", cfg.ES.Enabled)
	}
	if cfg.ES.Host != "http://test-es:9200/" {
		t.Errorf("Expected ES.Host to be 'http://test-es:9200/', got '%s'", cfg.ES.Host)
	}
	if cfg.ES.Prefix != "test_" {
		t.Errorf("Expected ES.Prefix to be 'test_', got '%s'", cfg.ES.Prefix)
	}
	if cfg.ES.Version != 6 {
		t.Errorf("Expected ES.Version to be 6, got %d", cfg.ES.Version)
	}
	if cfg.ES.User != "testuser" {
		t.Errorf("Expected ES.User to be 'testuser', got '%s'", cfg.ES.User)
	}
	if cfg.ES.Password != "testpass" {
		t.Errorf("Expected ES.Password to be 'testpass', got '%s'", cfg.ES.Password)
	}

	if cfg.Log.Level != "debug" {
		t.Errorf("Expected Log.Level to be 'debug', got '%s'", cfg.Log.Level)
	}
	if cfg.Log.File != "/tmp/test.log" {
		t.Errorf("Expected Log.File to be '/tmp/test.log', got '%s'", cfg.Log.File)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// 创建临时配置文件（无效的 YAML）
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid_config.yaml")

	invalidYAML := `
node:
  url: "https://test.example.com/eth"
  grpc: "test:8802"
    invalid_indent: "error"
`

	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(configFile)

	cfg, err := LoadConfig(configFile)
	if err == nil {
		t.Error("LoadConfig should return error for invalid YAML")
	}

	if cfg != nil {
		t.Error("LoadConfig should return nil config for invalid YAML")
	}
}

func TestLoadConfig_PartialConfig(t *testing.T) {
	// 测试部分配置（只设置部分字段）
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "partial_config.yaml")

	yamlContent := `
node:
  url: "https://partial.example.com/eth"

scanner:
  start_block: 5000000
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(configFile)

	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadConfig returned nil config")
	}

	// 验证设置的字段
	if cfg.Node.URL != "https://partial.example.com/eth" {
		t.Errorf("Expected Node.URL to be 'https://partial.example.com/eth', got '%s'", cfg.Node.URL)
	}
	if cfg.Scanner.StartBlock != 5000000 {
		t.Errorf("Expected Scanner.StartBlock to be 5000000, got %d", cfg.Scanner.StartBlock)
	}

	// 验证未设置的字段使用零值
	if cfg.Node.GRPC != "" {
		t.Errorf("Expected Node.GRPC to be empty, got '%s'", cfg.Node.GRPC)
	}
	if cfg.Scanner.EndBlock != 0 {
		t.Errorf("Expected Scanner.EndBlock to be 0, got %d", cfg.Scanner.EndBlock)
	}
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	// 测试空文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "empty_config.yaml")

	err := os.WriteFile(configFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(configFile)

	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadConfig returned nil config")
	}

	// 空文件应该返回零值配置
	if cfg.Node.URL != "" {
		t.Errorf("Expected Node.URL to be empty, got '%s'", cfg.Node.URL)
	}
}
