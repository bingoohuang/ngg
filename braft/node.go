package braft

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	transport "github.com/Jille/raft-grpc-transport"
	"github.com/bingoohuang/ngg/braft/discovery"
	"github.com/bingoohuang/ngg/braft/fsm"
	"github.com/bingoohuang/ngg/braft/marshal"
	"github.com/bingoohuang/ngg/braft/proto"
	"github.com/bingoohuang/ngg/braft/util"
	"github.com/bingoohuang/ngg/gnet"
	"github.com/bingoohuang/ngg/rotatefile/stdlog"
	"github.com/bingoohuang/ngg/ss"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-sockaddr"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"github.com/sqids/sqids-go"
	"github.com/vishal-bihani/go-tsid"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Node is the raft cluster node.
type Node struct {
	StartTime time.Time
	ctx       context.Context

	wg         *sync.WaitGroup
	Raft       *raft.Raft
	GrpcServer *grpc.Server

	httpServer   *http.Server
	memberConfig *memberlist.Config
	mList        *memberlist.Memberlist
	cancelFunc   context.CancelFunc

	Conf *Config

	TransportManager *transport.Manager
	distributor      *fsm.Distributor
	raftLogSum       *uint64

	addrQueue  *util.UniqueQueue
	notifyCh   chan NotifyEvent
	addr       string
	ID         string
	fns        []ConfigFn
	RaftID     RaftID
	stopped    uint32
	GrpcListen net.Listener

	DistributeCache sync.Map // map[string]fsm.Distributable
}

// Config is the configuration of the node.
type Config struct {
	TypeRegister    *marshal.TypeRegister
	DataDir         string
	Discovery       discovery.Discovery
	Services        []fsm.Service
	LeaderChange    NodeStateChanger
	BizData         func() any
	HTTPConfigFns   []HTTPConfigFn
	EnableHTTP      bool
	GrpcDialOptions []grpc.DialOption

	// Rport Raft 监听端口值
	Rport int
	// Dport Discovery 端口值
	Dport int
	// Hport HTTP 端口值
	Hport int
	// Raft ServiceID
	ServerID string
	// ShutdownExit 集群停止时，直接退出，方便 systemctl 重启
	ShutdownExit bool
	// ShutdownExitCode 退出时的编码
	ShutdownExitCode int

	// HostIP 当前主机的IP
	HostIP string
}

// RaftID is the structure of node ID.
type RaftID struct {
	ID         string `json:"id,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	IP         string `json:"ip,omitempty"`
	Sqid       string `json:"sqid,omitempty"` // {RaftPort, Dport, Hport}
	ServerID   string `json:"serverID,omitempty"`
	ServerAddr string `json:"serverAddr,omitempty"`
}

func (i RaftID) NodeID() string {
	if i.ServerID != "" {
		return i.ServerID
	}
	return i.ID
}

// NewNode returns an BRaft node.
func NewNode(fns ...ConfigFn) (*Node, error) {
	node := &Node{fns: fns}
	if err := node.createNode(); err != nil {
		return nil, err
	}

	return node, nil
}

func UnmarshRaftID(serverID string) (*RaftID, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(serverID)
	if err != nil {
		return nil, err
	}

	var raftID RaftID
	if err := msgpack.Unmarshal(decoded, &raftID); err != nil {
		return nil, err
	}
	return &raftID, nil
}

var ErrNoLeader = errors.New("no leader found")

func (n *Node) Leader() (*RaftID, error) {
	leader, id := n.Raft.LeaderWithID()
	if id == "" {
		return nil, ErrNoLeader
	}

	raftID, err := UnmarshRaftID(string(id))
	if err != nil {
		return nil, err
	}

	raftID.ServerAddr = string(leader)
	return raftID, nil
}

func (n *Node) createNode() error {
	conf, err := createConfig(n.fns)
	if err != nil {
		return err
	}

	log.Printf("node data dir: %s", conf.DataDir)

	raftID := RaftID{
		ID: tsid.Fast().ToString(),
		Sqid: ss.Pick1(ss.Pick1(sqids.New()).Encode(
			[]uint64{uint64(conf.Rport), uint64(conf.Dport), uint64(conf.Hport)})),
		Hostname: ss.Pick1(os.Hostname()),
		IP:       conf.HostIP,
		ServerID: conf.ServerID,
	}

	raftIDMsg, _ := msgpack.Marshal(raftID)
	nodeID := base64.RawURLEncoding.EncodeToString(raftIDMsg)

	log.Printf("nodeID: %s", nodeID)

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(nodeID)
	raftConf.LogLevel = hclog.Info.String()
	raftConf.Logger = &logger{}

	stableStoreFile := filepath.Join(conf.DataDir, "store.boltdb")
	if util.FileExists(stableStoreFile) {
		if err := os.Remove(stableStoreFile); err != nil {
			return err
		}
	}
	// StableStore 稳定存储,存储Raft集群的节点信息
	stableStore, err := raftboltdb.NewBoltStore(stableStoreFile)
	if err != nil {
		return err
	}

	// LogStore 存储Raft的日志
	logStore, err := raft.NewLogCache(512, stableStore)
	if err != nil {
		return err
	}

	// SnapshotStore 快照存储,存储节点的快照信息
	snapshotStore := raft.NewDiscardSnapshotStore()

	// FSM 有限状态机
	sm := fsm.NewRoutingFSM(raftID.NodeID(), conf.Services, conf.TypeRegister)

	// default raft config
	addr := fmt.Sprintf("%s:%d", conf.HostIP, conf.Rport)
	// grpc transport, Transport Raft节点之间的通信通道
	t := transport.New(raft.ServerAddress(addr), conf.GrpcDialOptions)

	// raft server
	raftServer, err := raft.NewRaft(raftConf, sm, logStore, stableStore, snapshotStore, t.Transport())
	if err != nil {
		return err
	}

	n.ID = nodeID
	n.RaftID = raftID
	n.addr = fmt.Sprintf(":%d", conf.Rport)
	n.Raft = raftServer
	n.TransportManager = t
	n.Conf = conf
	n.memberConfig = func(nodeID string, dport, rport int) *memberlist.Config {
		c := memberlist.DefaultLocalConfig()

		if conf.HostIP != "" {
			c.AdvertiseAddr = conf.HostIP
			c.AdvertisePort = dport
		} else {
			// fix "get final advertise address: No private IP address found, and explicit IP not provided"
			if privateIP, _ := sockaddr.GetPrivateIP(); privateIP == "" {
				if allIPv4, _ := gnet.ListIPv4(); len(allIPv4) > 0 {
					c.AdvertiseAddr = allIPv4[0]
					c.AdvertisePort = dport
				}
			}
		}

		c.BindPort = dport
		c.Name = fmt.Sprintf("%s:%d", nodeID, rport)
		c.Logger = log.Default()
		return c
	}(nodeID, conf.Dport, conf.Rport)

	n.distributor = fsm.NewDistributor()
	n.raftLogSum = &sm.RaftLogSum
	n.addrQueue = util.NewUniqueQueue(100)
	n.notifyCh = make(chan NotifyEvent, 100)

	return nil
}

// Start starts the Node and returns a channel that indicates, that the node has been stopped properly
func (n *Node) Start() (err error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		go func() {
			n.Stop()
		}()

		// 等待10秒，等待 memberlist leave node
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}()

	for {
		if err := n.start(); err != nil {
			return err
		}

		n.wait()

		if err = n.createNode(); err != nil {
			log.Printf("restart failed: %v", err)
			return err
		}

		log.Printf("restart sucessfully")
	}
}

func (n *Node) start() (err error) {
	n.StartTime = time.Now()
	log.Printf("Node starting, rport: %d, dport: %d, hport: %d, discovery: %s",
		n.Conf.Rport, n.Conf.Dport, n.Conf.Hport, n.DiscoveryName())

	// 防止各个节点同时启动太快，随机休眠
	util.Think(ss.Or(util.Env("BRAFT_SLEEP", "BSL"), "10ms-15s"), "")

	// set stopped as false
	atomic.CompareAndSwapUint32(&n.stopped, 1, 0)

	f := n.Raft.BootstrapCluster(raft.Configuration{
		Servers: []raft.Server{{
			ID:      raft.ServerID(n.ID),
			Address: n.TransportManager.Transport().LocalAddr(),
		}},
	})
	if err := f.Error(); err != nil {
		return err
	}

	// memberlist discovery
	n.memberConfig.Events = n
	if n.mList, err = memberlist.Create(n.memberConfig); err != nil {
		return err
	}

	// grpc server
	grpcListen, err := net.Listen("tcp", n.addr)
	if err != nil {
		return err
	}

	n.GrpcListen = grpcListen
	n.GrpcServer = grpc.NewServer()
	// register management services
	n.TransportManager.Register(n.GrpcServer)

	// register client services
	proto.RegisterRaftServer(n.GrpcServer, NewClientGrpcService(n))

	if off, _ := ss.Parse[bool](util.Env("DISABLE_GRPC_REFLECTION", "DGR")); !off {
		reflection.Register(n.GrpcServer)
	}

	stdlog.RegisterCustomLevel("[DEBUG]", stdlog.DebugLevel)

	n.wg = &sync.WaitGroup{}
	n.ctx, n.cancelFunc = context.WithCancel(context.Background())

	// discovery method
	discoveryChan, err := n.Conf.Discovery.Start(n.ctx, n.ID, n.Conf.Dport)
	if err != nil {
		return err
	}

	n.goHandleDiscoveredNodes(discoveryChan)

	// serve grpc
	util.Go(n.wg, func() {
		err := n.GrpcServer.Serve(grpcListen)
		log.Printf("GrpcServer stopped: %v", err)
	})

	util.GoChan(n.ctx, n.wg, n.Raft.LeaderCh(), func(becameLeader bool) error {
		log.Printf("becameLeader: %v", n.Raft.State())
		n.Conf.LeaderChange(n, n.Raft.State())
		util.Go(n.wg, n.watchNodesDiff)
		return nil
	})

	n.goDealNotifyEvent()
	if n.Conf.EnableHTTP {
		n.runHTTP(n.Conf.HTTPConfigFns...)
	}

	log.Printf("Node started")

	return nil
}

// DiscoveryName returns the name of discovery.
func (n *Node) DiscoveryName() string { return n.Conf.Discovery.Name() }

// Stop stops the node and notifies on a stopped channel returned in Start.
func (n *Node) Stop() {
	if !atomic.CompareAndSwapUint32(&n.stopped, 0, 1) {
		return
	}

	log.Print("Stopping Node...")
	n.Conf.LeaderChange(n, raft.Shutdown)
	n.cancelFunc()

	if n.Conf.ShutdownExit {
		go func() {
			util.Think("5s", "prepare for ShutdownExit")
			log.Printf("ShutdownExit %d", n.Conf.ShutdownExitCode)
			os.Exit(n.Conf.ShutdownExitCode)
		}()
	}

	err := n.mList.Leave(10 * time.Second)
	log.Printf("mList leave: %v", err)

	err = n.mList.Shutdown()
	log.Printf("mList shutdown: %v", err)

	err = n.GrpcListen.Close()
	log.Printf("GrpcListen close: %v", err)

	n.GrpcServer.Stop()
	log.Print("GrpcServer Server stopped")

	go func() {
		err := n.Raft.Shutdown().Error()
		log.Printf("Raft shutdown Raft: %v", err)
	}()

	if n.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := n.httpServer.Shutdown(ctx)
		log.Printf("http server Shutdown: %v", err)
	}
}

// ExistsServer 根据 serverID 查看 Raft 集群，看是否存在.
func (n *Node) ExistsServer(serverID string) bool {
	for _, s := range n.GetRaftServers() {
		if string(s.ID) == serverID {
			return true
		}
	}
	return false
}

// goHandleDiscoveredNodes handles the discovered Node additions
func (n *Node) goHandleDiscoveredNodes(discoveryChan chan string) {
	util.GoChan(n.ctx, n.wg, discoveryChan, func(peer string) error {
		peerHost, port := util.Cut(peer, ":")
		if port == "" {
			peer = fmt.Sprintf("%s:%d", peerHost, n.Conf.Dport)
		}

		// format of peer should ip:port (the port is for discovery)
		log.Printf("start mlist joinn: %v", peer)
		_, err := n.mList.Join([]string{peer})
		log.Printf("%s joined memberlist error: %v", peer, err)

		return nil
	})
}

// NotifyType 定义通知类型
type NotifyType int

const (
	_ NotifyType = iota
	// NotifyJoin 通知加入 Raft 集群
	NotifyJoin
	// NotifyLeave 通知离开 Raft 集群
	NotifyLeave
	// NotifyUpdate 通知更新 Raft 集群
	NotifyUpdate
	// NotifyHeartbeat 心跳通知
	NotifyHeartbeat
)

func (t NotifyType) String() string {
	switch t {
	case NotifyJoin:
		return "NotifyJoin"
	case NotifyLeave:
		return "NotifyLeave"
	case NotifyUpdate:
		return "NotifyUpdate"
	case NotifyHeartbeat:
		return "NotifyHeartbeat"
	default:
		return "Unknown"
	}
}

// NotifyEvent 通知事件
type NotifyEvent struct {
	*memberlist.Node
	NotifyType
}

var waitLeaderTime = util.EnvDuration("BRAFT_RESTART_MIN", 90*time.Second)

func (n *Node) goDealNotifyEvent() {
	waitLeader := make(chan NotifyEvent, 100)

	util.GoChan(n.ctx, n.wg, n.notifyCh, func(e NotifyEvent) error {
		go n.processNotify(e, waitLeader)
		return nil
	})

	go n.notifyHeartbeat(n.ctx)

	util.GoChan(n.ctx, n.wg, waitLeader, func(e NotifyEvent) error {
		leaderAddr, leaderID, err := n.waitLeader(waitLeaderTime)
		if err != nil {
			return err
		}

		isLeader := n.IsLeader()
		log.Printf("leader waited, type: %s, leader: %s, leaderID: %s, isLeader: %t, node: %s",
			e.NotifyType, leaderAddr, leaderID, isLeader, ss.Json(e.Node))
		n.processNotifyAtLeader(isLeader, e)
		return nil
	})
}

func (n *Node) waitLeader(minWait time.Duration) (leaderAddr, leaderID string, err error) {
	start := time.Now()
	for {
		if addr, id := n.Raft.LeaderWithID(); addr != "" {
			log.Printf("waited leader: %s cost: %s", id, time.Since(start))
			return string(addr), string(id), nil
		}
		if time.Since(start) >= minWait {
			log.Printf("stop to wait leader, expired %s >= %s", time.Since(start), minWait)
			n.Stop()
			log.Printf("braft node stopped")
			return "", "", io.EOF
		}

		util.Think("10-20s", "wait for leader")
	}
}

func (n *Node) processNotify(e NotifyEvent, waitLeader chan NotifyEvent) {
	leader, _ := n.Raft.LeaderWithID() // return empty string if there is no current leader
	isLeader := n.IsLeader()
	log.Printf("received type: %s, leader: %s, isLeader: %t, node: %s",
		e.NotifyType, leader, isLeader, ss.Json(e.Node))
	if leader != "" {
		switch e.NotifyType {
		case NotifyJoin, NotifyLeave, NotifyUpdate:
			n.processNotifyAtLeader(isLeader, e)
		default:
			// ignore
		}
		return
	}

	select {
	case <-n.ctx.Done():
		return
	case waitLeader <- e:
		log.Printf("current no leader, to wait list, type: %s, leader: %s, isLeader: %t, node: %s",
			e.NotifyType, leader, isLeader, ss.Json(e.Node))
	default:
		log.Printf("too many waitLeaders")
	}
}

func (n *Node) processNotifyAtLeader(isLeader bool, e NotifyEvent) {
	leader, _ := n.Leader()
	log.Printf("processing type: %s, leader: %v, isLeader: %t, node: %s",
		e.NotifyType, leader, isLeader, ss.Json(e.Node))

	switch e.NotifyType {
	case NotifyJoin:
		n.join(e.Node)
	case NotifyLeave:
		n.leave(e.Node)
	default:
	}
}

func (n *Node) leave(node *memberlist.Node) {
	nodeID, _ := util.Cut(node.Name, ":")
	if r := n.Raft.RemoveServer(raft.ServerID(nodeID), 0, 0); r.Error() != nil {
		log.Printf("E! raft node left: %s, addr: %s, error: %v", node.Name, node.Addr, r.Error())
	} else {
		log.Printf("raft node left: %s, addr: %s sucessfully", node.Name, node.Addr)
	}
}

func (n *Node) join(node *memberlist.Node) {
	nodeID, _ := util.Cut(node.Name, ":")
	nodeAddr := fmt.Sprintf("%s:%d", node.Addr, node.Port-1)
	if r := n.Raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(nodeAddr), 0, 0); r.Error() != nil {
		log.Printf("E! raft node joined: %s, addr: %s, error: %v", node.Name, nodeAddr, r.Error())
	} else {
		log.Printf("raft node joined: %s, addr: %s sucessfully", node.Name, nodeAddr)
	}
}

// NotifyJoin triggered when a new Node has been joined to the cluster (discovery only)
// and capable of joining the Node to the raft cluster
func (n *Node) NotifyJoin(node *memberlist.Node) {
	n.notifyCh <- NotifyEvent{NotifyType: NotifyJoin, Node: node}
}

// NotifyLeave triggered when a Node becomes unavailable after a period of time
// it will remove the unavailable Node from the Raft cluster
func (n *Node) NotifyLeave(node *memberlist.Node) {
	n.notifyCh <- NotifyEvent{NotifyType: NotifyLeave, Node: node}
}

// NotifyUpdate responses the update of raft cluster member.
func (n *Node) NotifyUpdate(node *memberlist.Node) {
	n.notifyCh <- NotifyEvent{NotifyType: NotifyUpdate, Node: node}
}

// notifyHeartbeat 心跳.
func (n *Node) notifyHeartbeat(ctx context.Context) {
	t := time.NewTicker(15 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n.notifyCh <- NotifyEvent{NotifyType: NotifyHeartbeat}
		}
	}
}

// IsLeader tells whether the current node is the leader.
func (n *Node) IsLeader() bool { return n.Raft.VerifyLeader().Error() == nil }

// RaftApply is used to apply any new logs to the raft cluster
// this method will do automatic forwarding to the Leader Node
func (n *Node) RaftApply(request any, timeout time.Duration) (any, error) {
	payload, err := n.Conf.TypeRegister.Marshal(request)
	if err != nil {
		return nil, err
	}

	if n.IsLeader() {
		r := n.Raft.Apply(payload, timeout)
		if r.Error() != nil {
			return nil, r.Error()
		}

		rsp := r.Response()
		if err, ok := rsp.(error); ok {
			return nil, err
		}

		return rsp, nil
	}

	log.Printf("transfer to leader")
	return n.ApplyOnLeader(payload, 10*time.Second)
}

// ShortNodeIds returns a sorted list of short node IDs in the current raft cluster.
func (n *Node) ShortNodeIds() (nodeIds []string) {
	for _, server := range n.GetRaftServers() {
		rid := ParseRaftID(string(server.ID))
		nodeIds = append(nodeIds, rid.NodeID())
	}

	sort.Strings(nodeIds)
	return
}

// ParseRaftID parses the coded raft ID string a RaftID structure.
func ParseRaftID(s string) (rid RaftID) {
	data, _ := base64.RawURLEncoding.DecodeString(s)
	if err := msgpack.Unmarshal(data, &rid); err != nil {
		log.Printf("E! msgpack.Unmarshal raft id %s error:%v", s, err)
	}
	return rid
}

type DistributeOption struct {
	Key string // http /distribute/:key
}

type DistributeOptionFunc func(*DistributeOption)

func WithKey(key string) DistributeOptionFunc {
	return func(opt *DistributeOption) {
		opt.Key = key
	}
}

type DistributeResult struct {
	Items     any    `json:"items"`
	RaftApply any    `json:"raftApply"`
	Error     string `json:"error"`
}

// Distribute distributes the given bean to all the nodes in the cluster.
func (n *Node) Distribute(bean fsm.Distributable, fns ...DistributeOptionFunc) (any, error) {
	var opt DistributeOption
	for _, fn := range fns {
		fn(&opt)
	}

	var result DistributeResult

	if opt.Key != "" { // 如果 Key 不为空，则在 Leader 节点上，设置 KeyValue 缓存值
		if !n.IsLeader() {
			result.Error = "node is not leader"
			return result, nil
		}

		// 等待 HTTP GET /distribute/:key 取值
		n.DistributeCache.Store(opt.Key, bean)
		return nil, nil
	}

	items := bean.GetDistributableItems()
	dataLen := n.distributor.Distribute(n.ShortNodeIds(), items)
	result.Items = items

	log.Printf("distribute items(#%d): %s", dataLen, ss.Json(bean))
	applyResult, err := n.RaftApply(fsm.DistributeRequest{Data: bean}, time.Second)
	if err != nil {
		result.Error = err.Error()
	}

	result.RaftApply = applyResult
	return result, nil
}

func (n *Node) wait() {
	n.wg.Wait()
	log.Print("Node Stopped!")
}

func (n *Node) watchNodesDiff() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	nodesState := &nodesState{}

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			if !n.IsLeader() {
				return
			}

			if n.checkNodesDiff(nodesState) {
				n.Stop()
				return
			}
		}
	}
}

type nodesState struct {
	nodes           []string
	raftServerAddrs []string
}

func (s *nodesState) setState(nodes, raftServerAddrs []string) {
	equals := slices.Equal(nodes, s.nodes) &&
		slices.Equal(raftServerAddrs, s.raftServerAddrs)
	if !equals {
		s.nodes = nodes
		s.raftServerAddrs = raftServerAddrs

		log.Printf("dicoveried nodes: %v, raft nodes: %v", nodes, raftServerAddrs)
	}
}

// 检查当前集群发现的处于稳态的节点，是否与 Raft 集群节点不匹配
func (n *Node) checkNodesDiff(state *nodesState) bool {
	nodes, err := n.Conf.Discovery.Search()
	if err != nil {
		log.Printf("discovery search error: %v", err)
		return false
	}

	var raftServerAddrs []string
	servers := n.GetRaftServers()
	for _, server := range servers {
		id, err := UnmarshRaftID(string(server.ID))
		if err != nil {
			log.Printf("unmarsh raft id %s error: %v", server.ID, err)
			continue
		}
		ports := ss.Pick1(sqids.New()).Decode(id.Sqid) // {RaftPort, Dport, Hport}
		serverAddr := string(server.Address)
		if len(ports) >= 3 {
			host, _, _ := net.SplitHostPort(serverAddr)
			serverAddr = fmt.Sprintf("%s:%d", host, ports[1])
		}

		raftServerAddrs = append(raftServerAddrs, serverAddr)
	}

	state.setState(nodes, raftServerAddrs)

	// 检查稳态
	connectableNodes := make(map[string]bool)
	for _, node := range nodes {
		if CheckTCP(node) {
			connectableNodes[node] = true
		}
	}

	if len(servers) != len(connectableNodes) {
		return true
	}

	for _, raftServerAddr := range raftServerAddrs {
		if _, ok := connectableNodes[raftServerAddr]; !ok {
			return true
		}
	}

	return false
}

func CheckTCP(ipPort string) bool {
	c, err := net.DialTimeout("tcp", ipPort, 3*time.Second)
	if err != nil {
		return false
	}

	c.Close()
	return true
}

// logger adapters logger to LevelLogger.
type logger struct{}

func (l *logger) GetLevel() hclog.Level { return hclog.Debug }

// Log Emit a message and key/value pairs at a provided log level
func (l *logger) Log(level hclog.Level, msg string, args ...any) {
	for i, arg := range args {
		// Convert the field value to a string.
		switch st := arg.(type) {
		case hclog.Hex:
			args[i] = "0x" + strconv.FormatUint(uint64(st), 16)
		case hclog.Octal:
			args[i] = "0" + strconv.FormatUint(uint64(st), 8)
		case hclog.Binary:
			args[i] = "0b" + strconv.FormatUint(uint64(st), 2)
		case hclog.Format:
			args[i] = fmt.Sprintf(st[0].(string), st[1:]...)
		case hclog.Quote:
			args[i] = strconv.Quote(string(st))
		}
	}

	v := append([]any{"D!", msg}, args...)

	switch {
	case level <= hclog.Debug:
		v[0] = "D!"
	case level == hclog.Info:
		v[0] = "I!"
	case level == hclog.Warn:
		v[0] = "W!"
	case level >= hclog.Error:
		v[0] = "E!"
	}

	log.Print(logPrint(v))
}

func logPrint(a []any) string {
	var buf []byte
	for i, arg := range a {
		if i > 0 { // Add a space
			buf = append(buf, ' ')
		}
		buf = append(buf, []byte(fmt.Sprint(arg))...)
	}

	return string(buf)
}

func (l *logger) Trace(msg string, args ...any) { l.Log(hclog.Trace, msg, args...) }
func (l *logger) Debug(msg string, args ...any) { l.Log(hclog.Debug, msg, args...) }
func (l *logger) Info(msg string, args ...any)  { l.Log(hclog.Info, msg, args...) }
func (l *logger) Warn(msg string, args ...any)  { l.Log(hclog.Warn, msg, args...) }
func (l *logger) Error(msg string, args ...any) { l.Log(hclog.Error, msg, args...) }

func (l *logger) IsTrace() bool { return false }
func (l *logger) IsDebug() bool { return false }
func (l *logger) IsInfo() bool  { return false }
func (l *logger) IsWarn() bool  { return false }
func (l *logger) IsError() bool { return false }

func (l *logger) ImpliedArgs() []any             { return nil }
func (l *logger) With(...any) hclog.Logger       { return l }
func (l *logger) Name() string                   { return "" }
func (l *logger) Named(string) hclog.Logger      { return l }
func (l *logger) ResetNamed(string) hclog.Logger { return l }
func (l *logger) SetLevel(hclog.Level)           {}

func (l *logger) StandardLogger(*hclog.StandardLoggerOptions) *log.Logger { return nil }
func (l *logger) StandardWriter(*hclog.StandardLoggerOptions) io.Writer   { return nil }
