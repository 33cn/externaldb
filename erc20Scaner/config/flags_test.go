package config

import (
	"flag"
	"testing"
)

func TestDefineCommonFlags(t *testing.T) {
	// 重置 flag 包
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	flags := DefineCommonFlags()

	if flags == nil {
		t.Fatal("DefineCommonFlags returned nil")
	}

	// 验证所有字段都已初始化
	if flags.ConfigFile == nil {
		t.Error("ConfigFile should not be nil")
	}
	if flags.NodeURL == nil {
		t.Error("NodeURL should not be nil")
	}
	if flags.DBEnabled == nil {
		t.Error("DBEnabled should not be nil")
	}
	if flags.DBDSN == nil {
		t.Error("DBDSN should not be nil")
	}
	if flags.ESEnabled == nil {
		t.Error("ESEnabled should not be nil")
	}
	if flags.ESHost == nil {
		t.Error("ESHost should not be nil")
	}
	if flags.ESPrefix == nil {
		t.Error("ESPrefix should not be nil")
	}
	if flags.ESVersion == nil {
		t.Error("ESVersion should not be nil")
	}
	if flags.ESUser == nil {
		t.Error("ESUser should not be nil")
	}
	if flags.ESPassword == nil {
		t.Error("ESPassword should not be nil")
	}
	if flags.ChainGRPC == nil {
		t.Error("ChainGRPC should not be nil")
	}
	if flags.ChainSymbol == nil {
		t.Error("ChainSymbol should not be nil")
	}
}

func TestDefineScannerFlags(t *testing.T) {
	// 重置 flag 包
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	flags := DefineScannerFlags()

	if flags == nil {
		t.Fatal("DefineScannerFlags returned nil")
	}

	// 验证 CommonFlags 已初始化
	if flags.CommonFlags.ConfigFile == nil {
		t.Error("CommonFlags.ConfigFile should not be nil")
	}

	// 验证 Scanner 专用字段
	if flags.StartBlock == nil {
		t.Error("StartBlock should not be nil")
	}
	if flags.EndBlock == nil {
		t.Error("EndBlock should not be nil")
	}
}

func TestDefineRPCFlags(t *testing.T) {
	// 重置 flag 包
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	flags := DefineRPCFlags()

	if flags == nil {
		t.Fatal("DefineRPCFlags returned nil")
	}

	// 验证 CommonFlags 已初始化
	if flags.CommonFlags.ConfigFile == nil {
		t.Error("CommonFlags.ConfigFile should not be nil")
	}

	// 验证 RPC 专用字段
	if flags.Port == nil {
		t.Error("Port should not be nil")
	}
}

func TestFlags_DefaultValues(t *testing.T) {
	// 注意：由于 flag 包使用全局变量，我们不能在同一个测试中多次调用 DefineCommonFlags
	// 这个测试只验证一次调用的默认值
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	commonFlags := DefineCommonFlags()

	// 验证默认值
	if *commonFlags.ConfigFile != "config.yaml" {
		t.Errorf("Expected ConfigFile default to be 'config.yaml', got '%s'", *commonFlags.ConfigFile)
	}
	if *commonFlags.NodeURL != "" {
		t.Errorf("Expected NodeURL default to be empty, got '%s'", *commonFlags.NodeURL)
	}
	if *commonFlags.DBEnabled != false {
		t.Errorf("Expected DBEnabled default to be false, got %v", *commonFlags.DBEnabled)
	}
}

func TestFlags_ScannerDefaultValues(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	scannerFlags := DefineScannerFlags()

	if *scannerFlags.StartBlock != 0 {
		t.Errorf("Expected StartBlock default to be 0, got %d", *scannerFlags.StartBlock)
	}
	if *scannerFlags.EndBlock != 0 {
		t.Errorf("Expected EndBlock default to be 0, got %d", *scannerFlags.EndBlock)
	}
}

func TestFlags_RPCDefaultValues(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	rpcFlags := DefineRPCFlags()

	if *rpcFlags.Port != "8080" {
		t.Errorf("Expected Port default to be '8080', got '%s'", *rpcFlags.Port)
	}
}

