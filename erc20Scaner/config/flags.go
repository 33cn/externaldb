package config

import "flag"

// CommonFlags 通用命令行参数（scanner 和 rpc 共用）
type CommonFlags struct {
	ConfigFile  *string
	NodeURL     *string
	DBEnabled   *bool
	DBDSN       *string
	ESEnabled   *bool
	ESHost      *string
	ESPrefix    *string
	ESVersion   *int
	ESUser      *string
	ESPassword  *string
	ChainGRPC   *string
	ChainSymbol *string
}

// ScannerFlags Scanner 专用参数
type ScannerFlags struct {
	CommonFlags
	StartBlock *int64
	EndBlock   *int64
}

// RPCFlags RPC 专用参数
type RPCFlags struct {
	CommonFlags
	Port *string
}

// DefineCommonFlags 定义通用命令行参数
func DefineCommonFlags() *CommonFlags {
	return &CommonFlags{
		ConfigFile:  flag.String("c", "config.yaml", "config file path"),
		NodeURL:     flag.String("u", "", "node url (overrides config)"),
		DBEnabled:   flag.Bool("db", false, "enable database (overrides config)"),
		DBDSN:       flag.String("dsn", "", "database DSN (overrides config)"),
		ESEnabled:   flag.Bool("es", false, "enable ES mode (overrides config)"),
		ESHost:      flag.String("es-host", "", "ES host (overrides config)"),
		ESPrefix:    flag.String("es-prefix", "", "ES prefix (overrides config)"),
		ESVersion:   flag.Int("es-version", 0, "ES version (6 or 7, overrides config)"),
		ESUser:      flag.String("es-user", "", "ES username (overrides config)"),
		ESPassword:  flag.String("es-pwd", "", "ES password (overrides config)"),
		ChainGRPC:   flag.String("chain-grpc", "", "Chain33 gRPC host (overrides config)"),
		ChainSymbol: flag.String("chain-symbol", "", "Chain33 symbol (overrides config)"),
	}
}

// DefineScannerFlags 定义 Scanner 专用参数
func DefineScannerFlags() *ScannerFlags {
	common := DefineCommonFlags()
	return &ScannerFlags{
		CommonFlags: *common,
		StartBlock:  flag.Int64("s", 0, "start block (overrides config)"),
		EndBlock:    flag.Int64("e", 0, "end block (overrides config, use -1 for unlimited)"),
	}
}

// DefineRPCFlags 定义 RPC 专用参数
func DefineRPCFlags() *RPCFlags {
	common := DefineCommonFlags()
	return &RPCFlags{
		CommonFlags: *common,
		Port: flag.String("port", "8080", "HTTP server port"),
	}
}

