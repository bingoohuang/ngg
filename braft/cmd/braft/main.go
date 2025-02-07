package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/bingoohuang/ngg/braft"
	"github.com/bingoohuang/ngg/braft/fsm"
	"github.com/bingoohuang/ngg/braft/marshal"
	"github.com/bingoohuang/ngg/braft/ticker"
	_ "github.com/bingoohuang/ngg/rotatefile/stdlog/autoload"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/ver"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"github.com/thoas/go-funk"
	"github.com/vishal-bihani/go-tsid"
)

func main() {
	dh := &DemoPicker{}
	braft.DefaultMdnsService = "_braft._tcp,_demo"
	t := ticker.New(10 * time.Second)

	node, err := braft.NewNode(
		braft.WithServices(fsm.NewMemKvService(), fsm.NewDistributeService(dh)),
		braft.WithLeaderChange(func(n *braft.Node, s raft.RaftState) {
			log.Printf("nodeState: %s", s)
			if s == raft.Leader {
				t.Start(func() {
					log.Printf("ticker ticker, I'm %s, nodeIds: %v", n.Raft.State(), n.ShortNodeIds())
				})
			} else {
				t.Stop()
			}
		}),
		braft.WithHTTPFns(
			braft.WithHandler(http.MethodPost, "/distribute", dh.distributePost),
			braft.WithHandler(http.MethodGet, "/distribute", dh.distributeGet),
			braft.WithHandler(http.MethodGet, "/distribute2", dh.distributeGet2),
		))
	if err != nil {
		log.Fatalf("failed to new node, error: %v", err)
	}
	if err := node.Start(); err != nil {
		log.Fatalf("failed to start node, error: %v", err)
	}
}

type DemoItem struct {
	ID     string
	NodeID string
}

var _ fsm.DistributableItem = (*DemoItem)(nil)

func (d *DemoItem) GetItemID() string       { return d.ID }
func (d *DemoItem) SetNodeID(nodeID string) { d.NodeID = nodeID }

type DemoDist struct {
	Common string
	Items  []DemoItem
}

func (d *DemoDist) GetDistributableItems() any { return d.Items }

type DemoPicker struct{ DD *DemoDist }

func (d *DemoPicker) PickForNode(nodeID string, request any) {
	dd := request.(*DemoDist)
	dd.Items = funk.Filter(dd.Items, func(item DemoItem) bool {
		return item.NodeID == nodeID
	}).([]DemoItem)
	d.DD = dd
	log.Printf("got %d items: %s", len(dd.Items), ss.Json(dd))
}

func (d *DemoPicker) RegisterMarshalTypes(reg *marshal.TypeRegister) {
	reg.RegisterType(reflect.TypeOf(DemoDist{}))
}

func (d *DemoPicker) distributeGet2(ctx *gin.Context, n *braft.Node) {
	var dd DemoDist
	nodeID, err := n.GetDistribute("demo", &dd)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	dd.Items = funk.Filter(dd.Items, func(item DemoItem) bool {
		return item.NodeID == nodeID
	}).([]DemoItem)
	ctx.JSON(200, dd)
}

func (d *DemoPicker) distributeGet(ctx *gin.Context, _ *braft.Node) {
	ctx.JSON(http.StatusOK, d.DD)
}

func (d *DemoPicker) distributePost(ctx *gin.Context, n *braft.Node) {
	dd := &DemoDist{Items: makeRandItems(ctx.Query("n")), Common: tsid.Fast().ToString()}
	if result, err := n.Distribute(dd, braft.WithKey("demo")); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	} else {
		ctx.JSON(http.StatusOK, result)
	}
}

func makeRandItems(q string) (ret []DemoItem) {
	n, _ := strconv.Atoi(q)
	if n <= 0 {
		n = ss.Rand().Intn(20) + 1
	}

	for i := 0; i < n; i++ {
		ret = append(ret, DemoItem{ID: fmt.Sprintf("%d", i)})
	}

	return
}

func init() {
	ss.ParseArgs(&arg, os.Args)
}

var arg Arg

type Arg struct {
	Version bool `flag:",v"`
	Init    bool
}

// Usage is optional for customized show.
func (a Arg) Usage() string {
	return fmt.Sprintf(`
Usage of %s:
  -v    bool   show version
`, os.Args[0])
}

// VersionInfo is optional for a customized version.
func (a Arg) VersionInfo() string { return ver.Version() }
