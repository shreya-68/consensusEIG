
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
    list    []int
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
    newval int
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
        for i := 0; i < depth; i++ {
            for _, child := range curr.child {
                if child.path[i] == path[i] {
                    curr = child
                    break
                }
            }
        }
        curr.val = value
    }
    //fmt.Println("read")
    //fmt.Println(msg)
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

func (node *Node) createChildren(eigNode *EIGNode) {
    list := node.list
    for _, i := range list {
        found := 0
        for _, j := range eigNode.path {
            if i == j {
                found = 1
                break
            }
        }
        if found == 0 {
            newChild := &EIGNode{level:eigNode.level+1}
            newChild.path = make([]int, eigNode.level)
            copy(newChild.path, eigNode.path)
            newChild.path = append(newChild.path, i)
            if i == node.addr.Port {
                newChild.val = eigNode.val
            }
            eigNode.child = append(eigNode.child, newChild)
        }
    }
}

func (node *Node) traverseEIG(eigNode *EIGNode, depth int) []*EIGNode {
    if depth == 0 {
        node.createChildren(eigNode)
        //fmt.Println(eigNode.child)
        found := 0
        for _, j := range eigNode.path {
            if node.addr.Port == j {
                found = 1
                break
            }
        }
        if found == 0 {
            //fmt.Println("To send: ", eigNode.val)
            return []*EIGNode{eigNode}
        }
        return []*EIGNode{}
    }
    leaves := []*EIGNode{}
    for _, child := range eigNode.child {
        subLeaf := node.traverseEIG(child, depth-1)
        leaves = append(leaves, subLeaf...)
    }
    return leaves
}

//Message format ---> int1.int2.int3.currRoot:val,int1.int2.int3.currRoot:val, ...,
func (node *Node) initRound(roundNum int) {
    sendThis := node.traverseEIG(node.root, roundNum-1)
    msg := ""
    for _, each := range sendThis {
        for _, pathInt := range each.path {
            msg += strconv.Itoa(pathInt) + "."
        }
        val := strconv.Itoa(each.val)
        if node.byz == 1{
            val = strconv.Itoa(rand.Intn(2))
        }
        msg += node.name + ":" + val + ","
    }
    msg = strings.TrimRight(msg, ",")
    //msg = strconv.Itoa(node.val)
    node.broadcast(msg)
}


//func (node *Node) getConsensus() {
//    count := make(map[int]int)
//    for _, val := range node.setVal {
//        count[val] += 1
//        //fmt.Printf("I am %s. Setval: %d\n", node.name, val)
//    }
//    maxVal := 0
//    var maxKey int
//    for key, _ := range count {
//        if count[key] > maxVal {
//            maxKey = key
//            maxVal = count[key]
//        }
//    }
//    fmt.Println("The consensus is value ", maxKey)
//}

func (node *Node) getConsensus(level int) {
    curr := node.root
    stack := []*EIGNode{}
    queue := []*EIGNode{curr}
    for len(queue) > 0{
        x := queue[0] 
        queue = queue[1:]
        if x.level <= level {
            for _, child := range x.child {
                queue = append(queue, child)
            }
        }
        stack = append(stack, x)
    }
    for len(stack) > 0 {
        x := stack[len(stack)-1]
        stack = stack[:len(stack)-1]
        switch {
            case x.level == level+1: x.newval = x.val
            default: 
                count := make(map[int]int)
                for _, child := range x.child {
                    count[child.val] += 1
                }
                maxVal := 0
                var maxKey int
                for key, _ := range count {
                    if count[key] > maxVal {
                        maxKey = key
                        maxVal = count[key]
                    }
                }
                x.newval = maxKey
        }
    }
}

func (node *Node) initConsensus(faults int){
    for i := 1; i <= faults + 1; i++ {
        node.initRound(i)
        time.Sleep(300*time.Millisecond) 
    }
    node.getConsensus(faults)
    fmt.Printf("My %s final value is %d \n", node.name, node.root.newval)

}

func Client(port string, nbrs []string, byz int, faults int) {
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
    for _, nbr := range node.nbr {
       node.list = append(node.list, nbr.Port)
    }
    node.list = append(node.list, node.addr.Port)

    node.listen()
    time.Sleep(400*time.Millisecond) 
    node.initConsensus(faults)
    //fmt.Printf("Hi, my port is %s. The set of values I have received are: \n", node.name)
    //node.getConsensus()
}
