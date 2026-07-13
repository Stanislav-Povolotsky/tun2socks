package engine

import "time"

type Key struct {
	MTU                      int           `yaml:"mtu"`
	Mark                     int           `yaml:"fwmark"`
	Proxy                    string        `yaml:"proxy"`
	RestAPI                  string        `yaml:"restapi"`
	Device                   string        `yaml:"device"`
	LogLevel                 string        `yaml:"loglevel"`
	Interface                string        `yaml:"interface"`
	TCPKeepaliveIdleTime     time.Duration `yaml:"tcp-keepalive-idle-time"`
	TCPKeepaliveInterval     time.Duration `yaml:"tcp-keepalive-interval"`
	TCPKeepaliveCount        int           `yaml:"tcp-keepalive-count"`
	TCPModerateReceiveBuffer bool          `yaml:"tcp-moderate-receive-buffer"`
	TCPSendBufferSize        string        `yaml:"tcp-send-buffer-size"`
	TCPReceiveBufferSize     string        `yaml:"tcp-receive-buffer-size"`
	MulticastGroups          []string      `yaml:"multicast-groups"`
	TUNPreUp                 string        `yaml:"tun-pre-up"`
	TUNPostUp                string        `yaml:"tun-post-up"`
	UDPTimeout               time.Duration `yaml:"udp-timeout"`
	FakeDNS                  bool          `yaml:"fakedns"`
	FakeDNSNetIPv4           string        `yaml:"fakedns-net-ipv4"`
	FakeDNSListenAddress     string        `yaml:"fakedns-listen-addr"`
	DNSAddr                  string        `yaml:"dns-addr"`
}
