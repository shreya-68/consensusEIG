
package peer

import (
    "fmt"
    "time"
    "os"
    "net"
    "strconv"
    "math/rand"
    "strings"
    )

//Client Node with name, nbr are the first hop neighbours and status is current running status
type Node struct {
    name    string
    nbr     []*net.TCPAddr
    status  string
    addr    *net.TCPAddr
    val     int 
    setVal  []int
    byz     int
    root    *EIGNode
}

func checkErr(err error) {
    if err != nil {
        fmt.Printf("Fatal error: %s \n", err)
        os.Exit(1)
    }
}

type EIGNode struct {
    level int
    path  []int
    val   int
    child []*EIGNode
}

func initVal() int{
    return rand.Intn(2)
}

func (node *Node) handleClient(conn net.Conn) {
    var buf [256]byte
    n, err := conn.Read(buf[0:])
    checkErr(err)
    //val, err := strconv.Atoi(string(buf[0:n]))
    //checkErr(err)
    //node.setVal = append(node.setVal, val)
    msg := string(buf[0:n])
    eachNode := strings.Split(msg, ",")
    for _, each := range eachNode {
        pathVal := strings.Split(each, ":")
        pathStr := strings.Split(pathVal[0], ".")
        path := []int{}
        for _, eachInt := range pathStr {
            x, _ := strconv.Atoi(eachInt)
            path = append(path, x)
        }
        value, _  := strconv.Atoi(pathVal[1])
        depth := len(path)
        curr := node.root
        for i := 1; i < depth; i++ {
            for _, child := range curr.child {
                if child.path[i-1] == path[i-1] {
                    curr = child
                    break
                }
            }
        }
        newChild := &EIGNode{level:depth, val: value}
        newChild.path = make([]int, depth)
        copy(newChild.path, path)
        curr.child = append(curr.child, newChild)
    }
    fmt.Println("read")
    fmt.Println(msg)
    conn.Close() 
}

func (node *Node) accept(listener *net.TCPListener) {
    for {
            conn, err := listener.Accept()
            if err != nil {
                continue
            }
            go node.handleClient(conn)
            //node.connections[0] = &net.TCPConn(conn)
        }
}

func (node *Node) listen() {
    listener, err := net.ListenTCP("tcp", node.addr)
    checkErr(err)
    go node.accept(listener)
}

func (node *Node) openTCPconn(rcvr *net.TCPAddr) *net.TCPConn {
    conn, err := net.DialTCP("tcp", nil, rcvr)
    checkErr(err)
    return conn
}

func (node *Node) write(msg string, conn *net.TCPConn) {
        //fmt.Printf("Writing %s\n", msg)
        _, err := conn.Write([]byte(msg))
        checkErr(err)
}

func (node *Node) broadcast(msg string) {
    for _, nbr := range node.nbr {
        conn := node.openTCPconn(nbr)
        node.write(msg, conn)
        conn.Close()
        
    }
}


func traverseEIG(eigNode *EIGNode, depth int) []*EIGNode {
    if depth == 1 {
        return []*EIGNode{eigNode}
    }
    leaves := []*EIGNode{}
    for _, child := range eigNode.child {
        subLeaf := traverseEIG(child, depth-1)
        leaves = append(leaves, subLeaf...)
    }
    return leaves
}

//Message format ---> int1.int2.int3.currRoot:val,int1.int2.int3.currRoot:val, ...,
func (node *Node) initRound(roundNum int) {
    sendTo := traverseEIG(node.root, roundNum)
    msg := ""
    for _, each := range sendTo {
        for _, pathInt := range each.path {
            msg += strconv.Itoa(pathInt) + "."
        }
        val := strconv.Itoa(each.val)
        msg += node.name + ":" + val + ","
    }
    msg = strings.TrimRight(msg, ",")
    //msg = strconv.Itoa(node.val)
    node.broadcast(msg)
}


func (node *Node) getConsensus() {
    count := make(map[int]int)
    for _, val := range node.setVal {
        count[val] += 1
        //fmt.Printf("I am %s. Setval: %d\n", node.name, val)
    }
    maxVal := 0
    var maxKey int
    for key, _ := range count {
        if count[key] > maxVal {
            maxKey = key
            maxVal = count[key]
        }
    }
    fmt.Println("The consensus is value ", maxKey)
}

func Client(port string, nbrs []string, byz int) {
    node := Node{name: port, status: "Init", byz: byz}
    var err error
    port = ":" + port
    node.addr, err = net.ResolveTCPAddr("tcp", port)
    checkErr(err)
    tcpAddrNbr := make([]*net.TCPAddr, len(nbrs))
    for i, val := range nbrs {
        addr, err := net.ResolveTCPAddr("tcp", val)
        checkErr(err)
        tcpAddrNbr[i] = addr
    }
    node.nbr = tcpAddrNbr 
    rand.Seed(time.Now().UTC().UnixNano())
    node.val = initVal()
    node.setVal = append(node.setVal, node.val)
    msg := "My (" + strconv.Itoa(node.addr.Port) + ") initial value is " + strconv.Itoa(node.val)
    fmt.Println(msg)
    node.root = &EIGNode{level: 0, val: node.val}

    node.listen()
    time.Sleep(200*time.Millisecond) 
    node.initRound(1)
    time.Sleep(600*time.Millisecond) 
    node.initRound(2)
    time.Sleep(200*time.Millisecond) 
    //fmt.Printf("Hi, my port is %s. The set of values I have received are: \n", node.name)
    //node.getConsensus()
}
