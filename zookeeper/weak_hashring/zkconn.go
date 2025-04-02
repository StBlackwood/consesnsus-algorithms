package weak_hashring

import (
	"errors"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

type Cluster struct {
	ZKConn   *zk.Conn
	NodeID   string
	BasePath string
	OnChange func([]string)
}

func NewCluster(zkServers []string, nodeID string, onChange func([]string)) (*Cluster, error) {
	conn, _, err := zk.Connect(zkServers, time.Second*5)
	if err != nil {
		return nil, err
	}

	basePath := "/nodes"
	_, err = conn.Create(basePath, []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err != nil && !errors.Is(err, zk.ErrNodeExists) {
		return nil, err
	}

	nodePath := fmt.Sprintf("%s/%s", basePath, nodeID)
	_, err = conn.Create(nodePath, []byte{}, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		return nil, err
	}

	c := &Cluster{
		ZKConn:   conn,
		NodeID:   nodeID,
		BasePath: basePath,
		OnChange: onChange,
	}

	go c.watchNodes()

	return c, nil
}

func (c *Cluster) watchNodes() {
	for {
		children, _, ch, err := c.ZKConn.ChildrenW(c.BasePath)
		if err != nil {
			fmt.Println("Error watching zk nodes:", err)
			time.Sleep(1 * time.Second)
			continue
		}

		c.OnChange(children)

		// wait for the event
		evt := <-ch
		if evt.Type == zk.EventNodeChildrenChanged {
			fmt.Println("Node list changed, updating...")
		}
	}
}
