package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	opsctx "github.com/Aias00/cloudbase/internal/ops"
	"github.com/Aias00/cloudbase/internal/pkg/logger"
	"github.com/Aias00/cloudbase/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type OpsWSProxyConfig = opsctx.WSProxyConfig

const (
	envOpsWSTrustProxy     = "OPS_WS_TRUST_PROXY"
	envOpsWSTrustedProxies = "OPS_WS_TRUSTED_PROXIES"
	envOpsWSOriginPolicy   = "OPS_WS_ORIGIN_POLICY"
	envOpsWSMaxConns       = "OPS_WS_MAX_CONNS"
	envOpsWSMaxConnsPerIP  = "OPS_WS_MAX_CONNS_PER_IP"
)

const (
	OriginPolicyStrict     = opsctx.WSOriginPolicyStrict
	OriginPolicyPermissive = opsctx.WSOriginPolicyPermissive
)

var opsWSProxyConfig = loadOpsWSProxyConfigFromEnv()

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return isAllowedOpsWSOrigin(r)
	},
	// Subprotocol negotiation:
	// - The frontend passes ["cloudbase-admin", "jwt.<token>"].
	// - We always select "cloudbase-admin" so the token is never echoed back in the handshake response.
	Subprotocols: []string{"cloudbase-admin"},
}

const (
	qpsWSPushInterval       = 2 * time.Second
	qpsWSRefreshInterval    = 5 * time.Second
	qpsWSRequestCountWindow = 1 * time.Minute

	defaultMaxWSConns      = 100
	defaultMaxWSConnsPerIP = 20
)

var wsConnCount atomic.Int32
var wsConnCountByIPMu sync.Mutex
var wsConnCountByIP = make(map[string]int32)

const qpsWSIdleStopDelay = 30 * time.Second

const (
	opsWSCloseRealtimeDisabled = 4001
)

var qpsWSIdleStopMu sync.Mutex
var qpsWSIdleStopTimer *time.Timer

func cancelQPSWSIdleStop() {
	qpsWSIdleStopMu.Lock()
	if qpsWSIdleStopTimer != nil {
		qpsWSIdleStopTimer.Stop()
		qpsWSIdleStopTimer = nil
	}
	qpsWSIdleStopMu.Unlock()
}

func scheduleQPSWSIdleStop() {
	qpsWSIdleStopMu.Lock()
	if qpsWSIdleStopTimer != nil {
		qpsWSIdleStopMu.Unlock()
		return
	}
	qpsWSIdleStopTimer = time.AfterFunc(qpsWSIdleStopDelay, func() {
		// Only stop if truly idle at fire time.
		if wsConnCount.Load() == 0 {
			qpsWSCache.Stop()
		}
		qpsWSIdleStopMu.Lock()
		qpsWSIdleStopTimer = nil
		qpsWSIdleStopMu.Unlock()
	})
	qpsWSIdleStopMu.Unlock()
}

type opsWSRuntimeLimits = opsctx.WSRuntimeLimits

var opsWSLimits = loadOpsWSRuntimeLimitsFromEnv()

const (
	qpsWSWriteTimeout = 10 * time.Second
	qpsWSPongWait     = 60 * time.Second
	qpsWSPingInterval = 30 * time.Second

	// We don't expect clients to send application messages; we only read to process control frames (Pong/Close).
	qpsWSMaxReadBytes = 1024
)

type opsWSQPSCache struct {
	refreshInterval    time.Duration
	requestCountWindow time.Duration

	lastUpdatedUnixNano atomic.Int64
	payload             atomic.Value // []byte

	opsService *service.OpsService
	cancel     context.CancelFunc
	done       chan struct{}

	mu      sync.Mutex
	running bool
}

var qpsWSCache = &opsWSQPSCache{
	refreshInterval:    qpsWSRefreshInterval,
	requestCountWindow: qpsWSRequestCountWindow,
}

func (c *opsWSQPSCache) start(opsService *service.OpsService) {
	if c == nil || opsService == nil {
		return
	}

	for {
		c.mu.Lock()
		if c.running {
			c.mu.Unlock()
			return
		}

		// If a previous refresh loop is currently stopping, wait for it to fully exit.
		done := c.done
		if done != nil {
			c.mu.Unlock()
			<-done

			c.mu.Lock()
			if c.done == done && !c.running {
				c.done = nil
			}
			c.mu.Unlock()
			continue
		}

		c.opsService = opsService
		ctx, cancel := context.WithCancel(context.Background())
		c.cancel = cancel
		c.done = make(chan struct{})
		done = c.done
		c.running = true
		c.mu.Unlock()

		go func() {
			defer close(done)
			c.refreshLoop(ctx)
		}()
		return
	}
}

// Stop stops the background refresh loop.
// It is safe to call multiple times.
func (c *opsWSQPSCache) Stop() {
	if c == nil {
		return
	}

	c.mu.Lock()
	if !c.running {
		done := c.done
		c.mu.Unlock()
		if done != nil {
			<-done
		}
		return
	}
	cancel := c.cancel
	c.cancel = nil
	c.running = false
	c.opsService = nil
	done := c.done
	c.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}

	c.mu.Lock()
	if c.done == done && !c.running {
		c.done = nil
	}
	c.mu.Unlock()
}

func (c *opsWSQPSCache) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(c.refreshInterval)
	defer ticker.Stop()

	c.refresh(ctx)
	for {
		select {
		case <-ticker.C:
			c.refresh(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (c *opsWSQPSCache) refresh(parentCtx context.Context) {
	if c == nil {
		return
	}

	c.mu.Lock()
	opsService := c.opsService
	c.mu.Unlock()
	if opsService == nil {
		return
	}

	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()

	now := time.Now().UTC()
	stats, err := opsService.GetWindowStats(ctx, now.Add(-c.requestCountWindow), now)
	if err != nil || stats == nil {
		if err != nil {
			logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] refresh: get window stats failed: %v", err)
		}
		return
	}

	requestCount := stats.SuccessCount + stats.ErrorCountTotal
	qps := 0.0
	tps := 0.0
	if c.requestCountWindow > 0 {
		seconds := c.requestCountWindow.Seconds()
		qps = opsctx.RoundTo1DP(float64(requestCount) / seconds)
		tps = opsctx.RoundTo1DP(float64(stats.TokenConsumed) / seconds)
	}

	payload := gin.H{
		"type":      "qps_update",
		"timestamp": now.Format(time.RFC3339),
		"data": gin.H{
			"qps":           qps,
			"tps":           tps,
			"request_count": requestCount,
		},
	}

	msg, err := json.Marshal(payload)
	if err != nil {
		logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] refresh: marshal payload failed: %v", err)
		return
	}

	c.payload.Store(msg)
	c.lastUpdatedUnixNano.Store(now.UnixNano())
}

func (c *opsWSQPSCache) getPayload() []byte {
	if c == nil {
		return nil
	}
	if cached, ok := c.payload.Load().([]byte); ok && cached != nil {
		return cached
	}
	return nil
}

func closeWS(conn *websocket.Conn, code int, reason string) {
	if conn == nil {
		return
	}
	msg := websocket.FormatCloseMessage(code, reason)
	_ = conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(qpsWSWriteTimeout))
	_ = conn.Close()
}

// QPSWSHandler handles realtime QPS push via WebSocket.
// GET /api/v1/admin/ops/ws/qps
func (h *OpsHandler) QPSWSHandler(c *gin.Context) {
	clientIP := requestClientIP(c.Request)

	if h == nil || h.opsService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ops service not initialized"})
		return
	}

	// If realtime monitoring is disabled, prefer a successful WS upgrade followed by a clean close
	// with a deterministic close code. This prevents clients from spinning on 404/1006 reconnect loops.
	if !h.opsService.IsRealtimeMonitoringEnabled(c.Request.Context()) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ops realtime monitoring is disabled"})
			return
		}
		closeWS(conn, opsWSCloseRealtimeDisabled, "realtime_disabled")
		return
	}

	cancelQPSWSIdleStop()
	// Lazily start the background refresh loop so unit tests that never hit the
	// websocket route don't spawn goroutines that depend on DB/Redis stubs.
	qpsWSCache.start(h.opsService)

	// Reserve a global slot before upgrading the connection to keep the limit strict.
	if !tryAcquireOpsWSTotalSlot(opsWSLimits.MaxConns) {
		logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] connection limit reached: %d/%d", wsConnCount.Load(), opsWSLimits.MaxConns)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "too many connections"})
		return
	}
	defer func() {
		if wsConnCount.Add(-1) == 0 {
			scheduleQPSWSIdleStop()
		}
	}()

	if opsWSLimits.MaxConnsPerIP > 0 && clientIP != "" {
		if !tryAcquireOpsWSIPSlot(clientIP, opsWSLimits.MaxConnsPerIP) {
			logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] per-ip connection limit reached: ip=%s limit=%d", clientIP, opsWSLimits.MaxConnsPerIP)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "too many connections"})
			return
		}
		defer releaseOpsWSIPSlot(clientIP)
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] upgrade failed: %v", err)
		return
	}

	defer func() {
		_ = conn.Close()
	}()

	handleQPSWebSocket(c.Request.Context(), conn)
}

func tryAcquireOpsWSTotalSlot(limit int32) bool {
	if limit <= 0 {
		return true
	}
	for {
		current := wsConnCount.Load()
		if current >= limit {
			return false
		}
		if wsConnCount.CompareAndSwap(current, current+1) {
			return true
		}
	}
}

func tryAcquireOpsWSIPSlot(clientIP string, limit int32) bool {
	slotKey, ok := opsctx.WSClientIPSlotKey(clientIP)
	if !ok || limit <= 0 {
		return true
	}
	wsConnCountByIPMu.Lock()
	defer wsConnCountByIPMu.Unlock()
	current := wsConnCountByIP[slotKey]
	if current >= limit {
		return false
	}
	wsConnCountByIP[slotKey] = current + 1
	return true
}

func releaseOpsWSIPSlot(clientIP string) {
	slotKey, ok := opsctx.WSClientIPSlotKey(clientIP)
	if !ok {
		return
	}
	wsConnCountByIPMu.Lock()
	defer wsConnCountByIPMu.Unlock()
	current, ok := wsConnCountByIP[slotKey]
	if !ok {
		return
	}
	if current <= 1 {
		delete(wsConnCountByIP, slotKey)
		return
	}
	wsConnCountByIP[slotKey] = current - 1
}

func handleQPSWebSocket(parentCtx context.Context, conn *websocket.Conn) {
	if conn == nil {
		return
	}

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	var closeOnce sync.Once
	closeConn := func() {
		closeOnce.Do(func() {
			_ = conn.Close()
		})
	}

	closeFrameCh := make(chan []byte, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		conn.SetReadLimit(qpsWSMaxReadBytes)
		if err := conn.SetReadDeadline(time.Now().Add(qpsWSPongWait)); err != nil {
			logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] set read deadline failed: %v", err)
			return
		}
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(qpsWSPongWait))
		})
		conn.SetCloseHandler(func(code int, text string) error {
			select {
			case closeFrameCh <- websocket.FormatCloseMessage(code, text):
			default:
			}
			cancel()
			return nil
		})

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
					logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] read failed: %v", err)
				}
				return
			}
		}
	}()

	// Push QPS data every 2 seconds (values are globally cached and refreshed at most once per qpsWSRefreshInterval).
	pushTicker := time.NewTicker(qpsWSPushInterval)
	defer pushTicker.Stop()

	// Heartbeat ping every 30 seconds.
	pingTicker := time.NewTicker(qpsWSPingInterval)
	defer pingTicker.Stop()

	writeWithTimeout := func(messageType int, data []byte) error {
		if err := conn.SetWriteDeadline(time.Now().Add(qpsWSWriteTimeout)); err != nil {
			return err
		}
		return conn.WriteMessage(messageType, data)
	}

	sendClose := func(closeFrame []byte) {
		if closeFrame == nil {
			closeFrame = websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		}
		_ = writeWithTimeout(websocket.CloseMessage, closeFrame)
	}

	for {
		select {
		case <-pushTicker.C:
			msg := qpsWSCache.getPayload()
			if msg == nil {
				continue
			}
			if err := writeWithTimeout(websocket.TextMessage, msg); err != nil {
				logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] write failed: %v", err)
				cancel()
				closeConn()
				wg.Wait()
				return
			}

		case <-pingTicker.C:
			if err := writeWithTimeout(websocket.PingMessage, nil); err != nil {
				logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] ping failed: %v", err)
				cancel()
				closeConn()
				wg.Wait()
				return
			}

		case closeFrame := <-closeFrameCh:
			sendClose(closeFrame)
			closeConn()
			wg.Wait()
			return

		case <-ctx.Done():
			var closeFrame []byte
			select {
			case closeFrame = <-closeFrameCh:
			default:
			}
			sendClose(closeFrame)

			closeConn()
			wg.Wait()
			return
		}
	}
}

func isAllowedOpsWSOrigin(r *http.Request) bool {
	if r == nil {
		return false
	}
	trustProxyHeaders := shouldTrustOpsWSProxyHeaders(r)
	reqHost := r.Host
	if trustProxyHeaders {
		if xfHost, ok := opsctx.ParseWSForwardedHost(r.Header.Get("X-Forwarded-Host")); ok {
			reqHost = xfHost
		}
	}
	return opsctx.IsWSOriginAllowed(r.Header.Get("Origin"), reqHost, opsWSProxyConfig.OriginPolicy)
}

func shouldTrustOpsWSProxyHeaders(r *http.Request) bool {
	if r == nil {
		return false
	}
	if !opsWSProxyConfig.TrustProxy {
		return false
	}
	peerIP, ok := requestPeerIP(r)
	if !ok {
		return false
	}
	return opsctx.WSAddrInTrustedProxies(peerIP, opsWSProxyConfig.TrustedProxies)
}

func requestPeerIP(r *http.Request) (netip.Addr, bool) {
	if r == nil {
		return netip.Addr{}, false
	}
	return opsctx.ParseWSPeerAddr(r.RemoteAddr)
}

func requestClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	trustProxyHeaders := shouldTrustOpsWSProxyHeaders(r)
	if trustProxyHeaders {
		if clientIP, ok := opsctx.ParseWSForwardedForClientIP(r.Header.Get("X-Forwarded-For")); ok {
			return clientIP
		}
	}

	if peer, ok := requestPeerIP(r); ok && peer.IsValid() {
		return peer.String()
	}
	return ""
}

func loadOpsWSProxyConfigFromEnv() OpsWSProxyConfig {
	cfg, invalid := opsctx.ParseWSProxyConfig(opsctx.WSProxyConfigInput{
		TrustProxyRaw:     os.Getenv(envOpsWSTrustProxy),
		TrustedProxiesRaw: os.Getenv(envOpsWSTrustedProxies),
		OriginPolicyRaw:   os.Getenv(envOpsWSOriginPolicy),
		DefaultTrustProxy: true,
		DefaultProxies:    defaultTrustedProxies(),
		DefaultPolicy:     OriginPolicyPermissive,
	})

	if invalid.TrustProxy {
		logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] invalid %s=%q (expected bool); using default=%v", envOpsWSTrustProxy, os.Getenv(envOpsWSTrustProxy), cfg.TrustProxy)
	}
	if len(invalid.TrustedProxies) > 0 {
		logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] invalid %s entries ignored: %s", envOpsWSTrustedProxies, strings.Join(invalid.TrustedProxies, ", "))
	}
	if invalid.OriginPolicy {
		logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] invalid %s=%q (expected %q or %q); using default=%q", envOpsWSOriginPolicy, os.Getenv(envOpsWSOriginPolicy), OriginPolicyStrict, OriginPolicyPermissive, cfg.OriginPolicy)
	}

	return cfg
}

func loadOpsWSRuntimeLimitsFromEnv() opsWSRuntimeLimits {
	cfg, invalid := opsctx.ParseWSRuntimeLimits(opsctx.WSRuntimeLimitsInput{
		MaxConnsRaw:          os.Getenv(envOpsWSMaxConns),
		MaxConnsPerIPRaw:     os.Getenv(envOpsWSMaxConnsPerIP),
		DefaultMaxConns:      defaultMaxWSConns,
		DefaultMaxConnsPerIP: defaultMaxWSConnsPerIP,
	})

	if invalid.MaxConns {
		logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] invalid %s=%q (expected int>0); using default=%d", envOpsWSMaxConns, os.Getenv(envOpsWSMaxConns), cfg.MaxConns)
	}
	if invalid.MaxConnsPerIP {
		logger.LegacyPrintf("handler.admin.ops_ws", "[OpsWS] invalid %s=%q (expected int>=0); using default=%d", envOpsWSMaxConnsPerIP, os.Getenv(envOpsWSMaxConnsPerIP), cfg.MaxConnsPerIP)
	}
	return cfg
}

func defaultTrustedProxies() []netip.Prefix {
	prefixes, _ := opsctx.ParseWSTrustedProxyList("127.0.0.0/8,::1/128")
	return prefixes
}
