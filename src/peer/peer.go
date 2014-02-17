
package peer

import (
    "fmt"
    "time"
    "os"
    "net"
    "strconv"
    "math/rand"
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
    rootEIG *EIGNode
}

func checkErr(err error) {
    if err != nil {
        fmt.Printf("Fatal error: %s \n", err)
        os.Exit(1)
    }
}

type EIGNode struct {
    path []int
    val  int
    child []*EIGNode
}

func (node *Node) handleClient(conn net.Conn) {
    var buf [256]byte
    n, err := conn.Read(buf[0:])
    checkErr(err)
    val, err := strconv.Atoi(string(buf[0:n]))
    checkErr(err)
    node.setVal = append(node.setVal, val)
    //fmt.Println("read")
    //fmt.Println(string(buf[0:n]))
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

func (node *Node) broadcast() {
    msg := strconv.Itoa(node.val)
    for _, nbr := range node.nbr {
        conn := node.openTCPconn(nbr)
        node.write(msg, conn)
        
    }
}

func initVal() int{
    return rand.Intn(2)
}

func Client(port string, nbrs []string, byz int) {
    node := Node{name: port, status: "Init", byz: byz}
    var err error
    node.addr, err = net.ResolveTCPAddr("tcp", port)
    checkErr(err)
    tcpAddrNbr := make([]*net.TCPAddr, len(nbrs))
    for i, val := range nbrs {
        addr, err := net.ResolveTCPAddr("tcp", val)
        checkErr(err)
        tcpAddrNbr[i] = addr
    }
    node.nbr = tcpAddrNbr 
    //fmt.Printf("Hi my name is %s\n", node.name)
    node.listen()
    time.Sleep(200*time.Millisecond) 
    rand.Seed(time.Now().UTC().UnixNano())
    node.val = initVal()
    node.setVal = append(node.setVal, node.val)
    msg := "My (" + strconv.Itoa(node.addr.Port) + ") initial value is " + strconv.Itoa(node.val)
    fmt.Println(msg)
    node.broadcast()
    time.Sleep(200*time.Millisecond) 
    //fmt.Printf("Hi, my port is %s. The set of values I have received are: \n", node.name)
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
