package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

var cnfg = viper.New()

// InitConfig initializes configs
func InitConfig(configFile string) {
	setDefaults()

	cnfg.SetConfigFile(configFile)

	err := cnfg.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
}

func setDefaults() {
	cnfg.SetDefault("server_read_timeout", 10*time.Second)
	cnfg.SetDefault("server_write_timeout", 10*time.Second)

	cnfg.SetDefault("log.level", "info")
	cnfg.SetDefault("log.output", []string{"app.log"})
	cnfg.SetDefault("log.caller", false)
	cnfg.SetDefault("log.stacktrace", true)

	cnfg.SetDefault("socket.handshake_timeout", 8*time.Second)
	cnfg.SetDefault("socket.read_buffer_size", 4096)
	cnfg.SetDefault("socket.write_buffer_size", 4096)
}

// Listen returns address service should run on (e.g. localhost:8000)
func Listen() string {
	return cnfg.GetString("listen")
}

// ServerReadTimeout returns server read timeout
func ServerReadTimeout() time.Duration {
	return cnfg.GetDuration("server_read_timeout")
}

// ServerWriteTimeout returns server write timeout
func ServerWriteTimeout() time.Duration {
	return cnfg.GetDuration("server_write_timeout")
}

// LogLevel returns current logging level
func LogLevel() string {
	return cnfg.GetString("log.level")
}

// LogOutput returns files where logs should go
func LogOutput() []string {
	return cnfg.GetStringSlice("log.output")
}

// ShouldLogCaller returns whether we should caller
func ShouldLogCaller() bool {
	return cnfg.GetBool("log.caller")
}

// ShouldLogStacktrace returns whether we should log stacktrace
func ShouldLogStacktrace() bool {
	return cnfg.GetBool("log.stacktrace")
}

// HandshakeTimeout returns timeout in which the handshake should be made for websocket upgrade
func HandshakeTimeout() time.Duration {
	return cnfg.GetDuration("socket.handshake_timeout")
}

// ReadBufferSize returns size of buffer for websocket messages on read
func ReadBufferSize() int {
	return cnfg.GetInt("socket.read_buffer_size")
}

// WriteBufferSize returns size of buffer for websocket messages on write
func WriteBufferSize() int {
	return cnfg.GetInt("socket.write_buffer_size")
}
