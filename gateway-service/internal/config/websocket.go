package config

import "time"

type Websocket struct {
	MsgMaxSize        int           `yaml:"msg_max_size"`
	MsgMaxLength      int           `yaml:"msg_max_length"`
	WriteBufSize      int           `yaml:"write_buf_size"`
	ReadBufSize       int           `yaml:"read_buf_size"`
	EnableCompression bool          `yaml:"enable_compression"`
	WriteWait         time.Duration `yaml:"write_wait"`
	PongWait          time.Duration `yaml:"pong_wait"`
	PingPeriod        time.Duration `yaml:"ping_period"`
	MaxFailedPings    int           `yaml:"max_failed_pings"`
	CheckOrigin       bool          `yaml:"check_origin"`
	AllowedOrigins    []string      `yaml:"allowed_origins"`
}

func DefaultWebsocket() Websocket {
	return Websocket{
		WriteBufSize:      4096,
		ReadBufSize:       4096,
		MsgMaxSize:        16384,
		MsgMaxLength:      2000,
		WriteWait:         10 * time.Second,
		PongWait:          60 * time.Second,
		PingPeriod:        54 * time.Second,
		MaxFailedPings:    3,
		EnableCompression: true,
		CheckOrigin:       false,
	}
}
