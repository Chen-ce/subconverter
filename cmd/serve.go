package cmd

import (
	"fmt"
	"os"
	
	"github.com/Chen-ce/subconverter/config"
	"github.com/Chen-ce/subconverter/server"
	"github.com/spf13/cobra"
)

var (
	servePort       int
	serveHost       string
	serveConfigFile string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP API server",
	Long: `Start an HTTP API server for subscription conversion.
The server requires API key authentication for all endpoints except /health.`,
	Example: `  # Start server with default config
  subconverter serve
  
  # Start server with custom config
  subconverter serve --config config.yaml
  
  # Start server on custom port
  subconverter serve --port 9090`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 0, "Server port (overrides config file)")
	serveCmd.Flags().StringVarP(&serveHost, "host", "H", "", "Server host (overrides config file)")
	serveCmd.Flags().StringVarP(&serveConfigFile, "config", "c", "config.yaml", "Config file path")
}

func runServe(cmd *cobra.Command, args []string) error {
	// 加载配置
	var cfg *config.Config
	
	if _, statErr := os.Stat(serveConfigFile); statErr == nil {
		var err error
		cfg, err = config.Load(serveConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		fmt.Printf("Loaded config from %s\n", serveConfigFile)
	} else {
		fmt.Println("Config file not found, using default config")
		cfg = config.DefaultConfig()
	}
	
	// 命令行参数覆盖配置文件
	if servePort > 0 {
		cfg.Server.Port = servePort
	}
	if serveHost != "" {
		cfg.Server.Host = serveHost
	}
	
	// 设置全局配置
	config.Set(cfg)

	
	// 显示配置信息
	fmt.Printf("Server configuration:\n")
	fmt.Printf("  Host: %s\n", cfg.Server.Host)
	fmt.Printf("  Port: %d\n", cfg.Server.Port)
	fmt.Printf("  Auth: %v\n", cfg.Auth.Enabled)
	if cfg.Auth.Enabled {
		fmt.Printf("  API Keys: %d configured\n", len(cfg.Auth.APIKeys))
	}
	fmt.Printf("  Templates: %s\n", cfg.Templates.Dir)
	fmt.Println()
	
	// 创建并启动服务器
	srv := server.New(cfg)
	return srv.Start()
}
