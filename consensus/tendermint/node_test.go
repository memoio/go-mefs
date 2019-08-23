package node

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

func TestNodeSingle(t *testing.T) {
	p2pPort := 3234
	rpcPort := 3235
	nodePath := path.Join(os.Getenv("HOME"), "test", "node")
	tim := time.Now()
	SingleNode(t, nodePath, p2pPort, rpcPort, tim)
}

func TestMultiNode(t *testing.T) {
	groupHome := path.Join(os.Getenv("HOME"), "TendermintTest")
	groupId := "777"
	tim := time.Now()
	nodes := []*Node{NewNode(groupHome+"/node1", groupId, tim),
		NewNode(groupHome+"/node2", groupId, tim),
		NewNode(groupHome+"/node3", groupId, tim),
		NewNode(groupHome+"/node4", groupId, tim)}
	//初始化目录
	for i := 0; i < len(nodes); i++ {
		if err := nodes[i].InitHome(); err != nil {
			log.Fatal(err)
		}
	}
	//获取实例id
	var ids [4]string
	for i := 0; i < len(nodes); i++ {
		id, err := nodes[i].GetID()
		fmt.Println("id-", i, ":", id)
		if err != nil {
			log.Fatal(err)
		}
		ids[i] = id
	}
	log.Println("ids:", ids)
	//获取公钥
	var validators []string
	for i := 0; i < len(nodes); i++ {
		validator, _ := nodes[i].GetPublicKey()
		validators = append(validators, validator)
	}
	var peers []string
	for i := 0; i < len(nodes); i++ {
		nodes[i].AlterValidator(validators)
		peers = append(peers, ids[i]+"@127.0.0.1:"+strconv.Itoa(30100+i+1))
	}
	fmt.Println(peers)
	//启动
	p2pPort := [5]int{30101, 30102, 30103, 30104, 30105}
	rpcPort := [5]int{30201, 30202, 30203, 30204, 30205}
	for i := 0; i < len(nodes); i++ {
		go nodes[i].Run(p2pPort[i], rpcPort[i], peers) //, ids[1] + "@127.0.0.1:30102"
		time.Sleep(time.Second)
	}
	fmt.Println("节点启动，5s后开始修改对等节点信息")
	time.Sleep(5 * time.Second)

	peers = []string{}
	for i := 0; i < len(nodes); i++ {
		nodes[i].AlterValidator(validators)
		peers = append(peers, ids[i]+"@127.0.0.1:"+strconv.Itoa(30100+i+2))
	}
	fmt.Println(peers)

	nodes[1].Stop()
	fmt.Println("重启node0")
	nodes[1].Node.Start()

	fmt.Println("start node0")
	time.Sleep(time.Second)

	fmt.Println("node0 重启成功，修改其他节点的对等节点信息")
	for i := 1; i < len(nodes); i++ {
		go nodes[i].ChangePeers(peers)
		time.Sleep(time.Second)
	}

	// select {}
}

func SingleNode(t *testing.T, nodePath string, p2pPort, rpcPort int, tim time.Time) {
	node := NewNode(nodePath, "666", tim)
	if err := node.InitHome(); err != nil {
		log.Fatal(err)
	}

	id, err := node.GetID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)

	err = node.Run(p2pPort, rpcPort, []string{})
	if err != nil {
		log.Fatal(err)
	}
}

//测试
//http://localhost:30201/broadcast_tx_commit?tx="{\"challengerPrivateKey\":\"0x1\",\"acceptChallengerAddress\":\"0x2\",\"dataPath\":\"3\"}"
//http://localhost:30203/abci_query?data="0x7f0429d28a8353ac14446ffb454c853d872791ff4d5f8b05e8ea37954fbd7422"
