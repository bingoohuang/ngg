package braft

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bingoohuang/ngg/braft/fsm"
	"github.com/bingoohuang/ngg/braft/util"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/ver"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
	"github.com/imroc/req/v3"
	"github.com/samber/lo"
	"github.com/sqids/sqids-go"
)

// HTTPConfig is configuration for HTTP service.
type HTTPConfig struct {
	Handlers []pathHandler
	EnableKv bool
}

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc func(ctx *gin.Context, n *Node)

type pathHandler struct {
	handler HandlerFunc
	method  string
	path    string
}

// HTTPConfigFn is function options for HTTPConfig.
type HTTPConfigFn func(*HTTPConfig)

// WithEnableKV enables or disables KV service on HTTP.
func WithEnableKV(b bool) HTTPConfigFn { return func(c *HTTPConfig) { c.EnableKv = b } }

// WithHandler defines the http handler.
func WithHandler(method, path string, handler HandlerFunc) HTTPConfigFn {
	return func(c *HTTPConfig) {
		c.Handlers = append(c.Handlers, pathHandler{method: method, path: path, handler: handler})
	}
}

// runHTTP run http service on block.
func (n *Node) runHTTP(fs ...HTTPConfigFn) {
	c := &HTTPConfig{EnableKv: true}
	for _, f := range fs {
		f(c)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(util.Logger(true), gin.Recovery())
	r.GET("/raft", n.ServeRaft)
	r.GET("/distribute/:key", n.ServeDistribute)

	if c.EnableKv {
		n.RegisterServeKV(r, "/kv")
	}

	for _, h := range c.Handlers {
		hh := h.handler
		log.Printf("register method: %s path: %s handler: %s", h.method, h.path, ss.GetFuncName(hh))
		r.Handle(h.method, h.path, func(ctx *gin.Context) {
			hh(ctx, n)
		})
	}

	n.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", n.Conf.Hport),
		Handler: r,
	}

	go func() {
		if err := n.httpServer.ListenAndServe(); err != nil {
			log.Printf("E! listen: %s\n", err)
		}
	}()
}

func getQuery(ctx *gin.Context, k ...string) string {
	for _, _k := range k {
		if q, _ := ctx.GetQuery(_k); q != "" {
			return q
		}
	}

	return ""
}

// RaftNode is a node info of raft cluster.
type RaftNode struct {
	ServerID  string `json:"serverID"`
	BuildTime string `json:"buildTime"`
	Duration  string `json:"duration"`

	Address    string `json:"address"`
	RaftState  string `json:"raftState"`
	Leader     string `json:"leader"`
	AppVersion string `json:"appVersion"`
	StartTime  string `json:"startTime"`
	Error      string `json:"error,omitempty"`
	GoVersion  string `json:"goVersion"`

	GitCommit      string   `json:"gitCommit"`
	DiscoveryNodes []string `json:"discoveryNodes"`

	Addr []string `json:"addr"`

	BizData json.RawMessage `json:"bizData,omitempty"`

	RaftID RaftID `json:"raftID"`

	RaftLogSum uint64 `json:"raftLogSum"`
	Pid        uint64 `json:"pid"`

	Rss  uint64  `json:"rss"`
	Pcpu float32 `json:"pcpu"`

	RaftPort      int `json:"raftPort"`
	DiscoveryPort int `json:"discoveryPort"`
	HttpPort      int `json:"httpPort"`

	NodeIds []string `json:"nodeIds"`
}

// raftServer tracks the information about a single server in a configuration.
type raftServer struct {
	// Suffrage determines whether the server gets a vote.
	Suffrage raft.ServerSuffrage `json:"suffrage"`
	// ID is a unique string identifying this server for all time.
	ID raft.ServerID `json:"id"`
	// Address is its network address that a transport can contact.
	Address raft.ServerAddress `json:"address"`
}

var httpClient = req.C() // Use C() to create a client.

func (n *Node) GetDistribute(key string, result any) (nodeID string, err error) {
	r, err := n.Leader()
	if err != nil {
		return "", err
	}

	ports := ss.Pick1(sqids.New()).Decode(r.Sqid)
	httpPort := int(ports[2])
	addr := fmt.Sprintf("http://%s:%d/distribute/%s", r.IP, httpPort, key)
	if _, err := httpClient.R().SetSuccessResult(result).Get(addr); err != nil {
		return "", err
	}

	return n.RaftID.NodeID(), nil
}
func (n *Node) ServeDistribute(ctx *gin.Context) {
	key := ctx.Param("key")
	value, ok := n.DistributeCache.Load(key)
	if !ok {
		ctx.AbortWithStatus(410)
		return
	}

	bean, ok := value.(fsm.Distributable)
	if !ok {
		ctx.AbortWithStatus(411)
		return
	}

	items := bean.GetDistributableItems()
	dataLen := n.distributor.Distribute(n.ShortNodeIds(), items)
	log.Printf("ServeDistribute /distribute/%s %d items: %s", key, dataLen, ss.Json(bean))
	ctx.JSON(200, bean)
}

// ServeRaft services the raft http api.
func (n *Node) ServeRaft(ctx *gin.Context) {
	raftServers := n.GetRaftServers()
	nodes := n.GetRaftNodes(raftServers)
	leaderAddr, leaderID := n.Raft.LeaderWithID()
	ctx.JSON(http.StatusOK, gin.H{
		"leaderAddr":    leaderAddr,
		"leaderID":      leaderID,
		"nodes":         nodes,
		"nodeNum":       len(nodes),
		"discovery":     n.DiscoveryName(),
		"currentLeader": n.IsLeader(),
		"raftServers": lo.Map(raftServers, func(r raft.Server, index int) raftServer {
			return raftServer{
				Suffrage: r.Suffrage,
				ID:       r.ID,
				Address:  r.Address,
			}
		}),
		"memberList": lo.Map(n.mList.Members(), func(m *memberlist.Node, index int) any {
			return map[string]any{
				"name": m.Name,
				"addr": m.Addr,
				"port": m.Port,
			}
		}),
	})
}

// GetRaftNodesInfo return the raft nodes information.
func (n *Node) GetRaftNodesInfo() (nodes []RaftNode) {
	raftServers := n.GetRaftServers()
	return n.GetRaftNodes(raftServers)
}

// GetRaftNodes return the raft nodes information.
func (n *Node) GetRaftNodes(raftServers []raft.Server) (nodes []RaftNode) {
	for _, server := range raftServers {
		rsp, err := GetPeerDetails(string(server.Address), 3*time.Second)
		if err != nil {
			log.Printf("E! GetPeerDetails error: %v", err)
			nodes = append(nodes, RaftNode{Address: string(server.Address), Error: err.Error()})
			continue
		}

		rid := ParseRaftID(rsp.ServerId)
		ports := ss.Pick1(sqids.New()).Decode(rid.Sqid)

		nodes = append(nodes, RaftNode{
			RaftID:  rid,
			Address: string(server.Address), Leader: rsp.Leader,
			ServerID: rsp.ServerId, RaftState: rsp.RaftState,
			Error:          rsp.Error,
			DiscoveryNodes: rsp.DiscoveryNodes,
			StartTime:      rsp.StartTime,
			Duration:       rsp.Duration,

			Rss:        rsp.Rss,
			Pcpu:       rsp.Pcpu,
			RaftLogSum: rsp.RaftLogSum,
			Pid:        rsp.Pid,

			GitCommit:  ver.GitCommit,
			BuildTime:  ver.BuildTime,
			GoVersion:  ver.GoVersion,
			AppVersion: ver.AppVersion,

			BizData: json.RawMessage(rsp.BizData),

			Addr: rsp.Addr,

			RaftPort:      int(ports[0]),
			DiscoveryPort: int(ports[1]),
			HttpPort:      int(ports[2]),

			NodeIds: rsp.NodeIds,
		})
	}
	return
}

// GetRaftServers 获得 Raft 节点服务器列表.
func (n *Node) GetRaftServers() []raft.Server {
	return n.Raft.GetConfiguration().Configuration().Servers
}

// RegisterServeKV register kv service for the gin route.
func (n *Node) RegisterServeKV(r gin.IRoutes, path string) {
	r.GET(path, n.ServeKV)
	r.POST(path, n.ServeKV)
	r.DELETE(path, n.ServeKV)
}

// ServeKV services the kv set/get http api.
func (n *Node) ServeKV(ctx *gin.Context) {
	req := fsm.KvRequest{
		MapName: ss.Or(getQuery(ctx, "map", "m"), "default"),
		Key:     ss.Or(getQuery(ctx, "key", "k"), "default"),
	}
	ctx.Header("Braft-IP", n.RaftID.IP)
	ctx.Header("Braft-ID", n.RaftID.ID)
	ctx.Header("Braft-Host", n.RaftID.Hostname)
	switch ctx.Request.Method {
	case http.MethodPost:
		req.KvOperate = fsm.KvSet
		req.Value = getQuery(ctx, "value", "v")
	case http.MethodGet:
		req.KvOperate = fsm.KvGet
		for _, service := range n.Conf.Services {
			if m, ok := service.(fsm.KvExecutable); ok {
				ctx.JSON(http.StatusOK, m.Exec(req))
				return
			}
		}
	case http.MethodDelete:
		req.KvOperate = fsm.KvDel
	}

	result, err := n.RaftApply(req, time.Second)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	} else {
		ctx.JSON(http.StatusOK, result)
	}
}
