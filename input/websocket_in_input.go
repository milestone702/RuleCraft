package input

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"rulecraft/plugin"
)

// WebsocketInInput WebSocket 入站插件。
type WebsocketInInput struct {
	mu       sync.Mutex
	server   *http.Server
	started  bool
	port     int
	authKey  string
	lastData string
	lastTime int64
}

func NewWebsocketInInput() *WebsocketInInput {
	return &WebsocketInInput{port: 19533}
}

func (w *WebsocketInInput) ID() string        { return "websocket_in" }
func (w *WebsocketInInput) Name() string      { return "WebSocket 入站传感器" }
func (w *WebsocketInInput) IsAvailable() bool { return true }

func (w *WebsocketInInput) Configure(port int, authKey string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if port > 0 && port != w.port {
		w.stop()
		w.port = port
	}
	if authKey != "" {
		w.authKey = authKey
	}
}

func (w *WebsocketInInput) Collect(ctx *plugin.SystemContext) error {
	w.mu.Lock()
	if !w.started {
		w.started = true
		w.start()
	}
	data := w.lastData
	ts := w.lastTime
	w.mu.Unlock()

	if data != "" {
		ctx.SetState("websocket_in.data", data)
		ctx.SetState("websocket_in.timestamp", ts)
		ctx.SetState("websocket_in.port", w.port)
		parseIncomingData(ctx, "websocket_in", data, "auto")
	}
	return nil
}

func (w *WebsocketInInput) start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", w.handleWS)

	w.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", w.port),
		Handler: mux,
	}
	go func() {
		w.server.ListenAndServe()
	}()
}

const wsGUID = "258EAFA5-E914-47DA-95CA-5AB5DC11B735"

func (w *WebsocketInInput) handleWS(rw http.ResponseWriter, r *http.Request) {
	if w.authKey != "" {
		if r.Header.Get("X-API-Key") != w.authKey && r.URL.Query().Get("key") != w.authKey {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	hj, ok := rw.(http.Hijacker)
	if !ok {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		return
	}
	defer conn.Close()

	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
	conn.Write([]byte(resp))
	buf.Flush()

	reader := bufio.NewReader(conn)
	for {
		b0, err := reader.ReadByte()
		if err != nil {
			return
		}
		b1, err := reader.ReadByte()
		if err != nil {
			return
		}
		opcode := b0 & 0x0F
		masked := (b1 & 0x80) != 0
		payloadLen := int64(b1 & 0x7F)

		if payloadLen == 126 {
			var len16 uint16
			if err := binary.Read(reader, binary.BigEndian, &len16); err != nil {
				return
			}
			payloadLen = int64(len16)
		} else if payloadLen == 127 {
			var len64 uint64
			if err := binary.Read(reader, binary.BigEndian, &len64); err != nil {
				return
			}
			payloadLen = int64(len64)
		}

		var maskKey [4]byte
		if masked {
			if _, err := io.ReadFull(reader, maskKey[:]); err != nil {
				return
			}
		}

		payload := make([]byte, payloadLen)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return
		}

		if masked {
			for i := range payload {
				payload[i] ^= maskKey[i%4]
			}
		}

		if opcode == 1 { // text frame
			w.mu.Lock()
			w.lastData = string(payload)
			w.lastTime = time.Now().Unix()
			w.mu.Unlock()
		} else if opcode == 8 || opcode == 9 { // close or ping
			return
		}
	}
}

func (w *WebsocketInInput) stop() {
	if w.server != nil {
		w.server.Close()
		w.server = nil
	}
	w.started = false
}
