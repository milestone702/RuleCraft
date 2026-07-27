package input

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"

	"rulecraft/plugin"
)

// TCPUDPInInput TCP/UDP 入站插件。
type TCPUDPInInput struct {
	mu       sync.Mutex
	started  bool
	port     int
	proto    string // "tcp" 或 "udp"
	authKey  string
	lastData string
	lastAddr string
	lastTime int64
	listener net.Listener
	udpConn  *net.UDPConn
}

func NewTCPUDPInInput() *TCPUDPInInput {
	return &TCPUDPInInput{port: 19532, proto: "tcp"}
}

func (t *TCPUDPInInput) ID() string        { return "tcp_udp_in" }
func (t *TCPUDPInInput) Name() string      { return "TCP/UDP 入站传感器" }
func (t *TCPUDPInInput) IsAvailable() bool { return true }

// Configure 设置端口、协议和鉴权密钥。
func (t *TCPUDPInInput) Configure(port int, proto, authKey string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if port > 0 && (port != t.port || proto != t.proto) {
		t.stop()
		t.port = port
		t.proto = proto
	}
	if authKey != "" {
		t.authKey = authKey
	}
}

func (t *TCPUDPInInput) Collect(ctx *plugin.SystemContext) error {
	t.mu.Lock()
	if !t.started {
		t.started = true
		go t.listen()
	}
	data := t.lastData
	addr := t.lastAddr
	ts := t.lastTime
	t.mu.Unlock()

	if data != "" {
		ctx.SetState("tcp_udp_in.data", data)
		ctx.SetState("tcp_udp_in.addr", addr)
		ctx.SetState("tcp_udp_in.timestamp", ts)
		ctx.SetState("tcp_udp_in.port", t.port)
		ctx.SetState("tcp_udp_in.proto", t.proto)
		parseIncomingData(ctx, "tcp_udp_in", data, "auto")
	}
	return nil
}

func (t *TCPUDPInInput) listen() {
	addr := fmt.Sprintf(":%d", t.port)

	if t.proto == "udp" {
		udpAddr, _ := net.ResolveUDPAddr("udp", addr)
		conn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			return
		}
		t.udpConn = conn
		buf := make([]byte, 65535)
		for {
			n, remoteAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			t.mu.Lock()
			t.lastData = string(buf[:n])
			t.lastAddr = remoteAddr.String()
			t.lastTime = time.Now().Unix()
			t.mu.Unlock()
		}
	} else {
		var err error
		t.listener, err = net.Listen("tcp", addr)
		if err != nil {
			return
		}
		for {
			conn, err := t.listener.Accept()
			if err != nil {
				return
			}
			go t.handleTCP(conn)
		}
	}
}

func (t *TCPUDPInInput) handleTCP(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		t.mu.Lock()
		t.lastData = line
		t.lastAddr = conn.RemoteAddr().String()
		t.lastTime = time.Now().Unix()
		t.mu.Unlock()
	}
}

func (t *TCPUDPInInput) stop() {
	if t.listener != nil {
		t.listener.Close()
		t.listener = nil
	}
	if t.udpConn != nil {
		t.udpConn.Close()
		t.udpConn = nil
	}
	t.started = false
}
