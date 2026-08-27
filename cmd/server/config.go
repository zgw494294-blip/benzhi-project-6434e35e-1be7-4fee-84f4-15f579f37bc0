package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type serverConfig struct {
	Addr      string
	Selfcheck bool
	Database  string
}

func defaultConfig() serverConfig {
	return serverConfig{Addr: "127.0.0.1:19081", Database: "archive.db"}
}
func parseConfig(fs *flag.FlagSet) serverConfig {
	c := defaultConfig()
	addr := fs.String("addr", c.Addr, "监听地址")
	self := fs.Bool("selfcheck", false, "运行自检")
	db := fs.String("database", c.Database, "数据文件")
	fs.Parse(os.Args[1:])
	c.Addr = *addr
	c.Selfcheck = *self
	c.Database = *db
	if env := strings.TrimSpace(os.Getenv("PORT")); env != "" && c.Addr == defaultConfig().Addr {
		if n, e := strconv.Atoi(env); e == nil && n > 0 && n < 65536 {
			c.Addr = "127.0.0.1:" + env
		}
	}
	return c
}
func validateConfig(c serverConfig) error {
	if c.Addr == "" {
		return fmt.Errorf("监听地址为空")
	}
	host, port, e := net.SplitHostPort(c.Addr)
	if e != nil || host != "127.0.0.1" {
		return fmt.Errorf("地址必须绑定回环接口")
	}
	n, e := strconv.Atoi(port)
	if e != nil || n < 1024 || n > 65535 {
		return fmt.Errorf("端口必须在 1024-65535")
	}
	if c.Database == "" {
		return fmt.Errorf("数据文件为空")
	}
	return nil
}
func selfcheckConfig() serverConfig {
	c := defaultConfig()
	c.Selfcheck = true
	c.Database = "file:selfcheck?mode=memory&cache=shared"
	return c
}
func isHighPort(addr string) bool {
	_, p, e := net.SplitHostPort(addr)
	if e != nil {
		return false
	}
	n, _ := strconv.Atoi(p)
	return n >= 1024
}

func configDescription(c serverConfig) string {
	return fmt.Sprintf("监听 %s，数据文件 %s，自检=%t", c.Addr, c.Database, c.Selfcheck)
}

func normalizeAddress(addr string) string {
	if strings.TrimSpace(addr) == "" {
		return defaultConfig().Addr
	}
	return strings.TrimSpace(addr)
}
