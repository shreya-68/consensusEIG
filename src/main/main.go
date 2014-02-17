package main

import (
    "fmt"
    "time"
    "os"
//    "net"
    "peer"
    "strconv"
    )



type Graph struct {
    numNodes int
    nodes    []*peer.Node
    edges    map[*peer.Node]*peer.Node
}



var graph Graph

func checkErr(err error) {
    if err != nil {
        fmt.Printf("Fatal error: %s \n", err)
        os.Exit(1)
    }
}

func initGraph() {
    num, _ := strconv.Atoi(os.Args[1])
    graph = Graph{numNodes: num}
}



func main() {
    //Create a centralised monitor

    //Initialise graph
    initGraph()

    //Launch Client goroutines
    all := make([]string, graph.numNodes)
    for i := 0; i < graph.numNodes; i++{
         all[i] = ":" + strconv.Itoa(9000+i)
    }
    
    var byz int
    faults, _ := strconv.Atoi(os.Args[2])
    //conVal := make([]string, graph.numNodes)
    for i := 0; i < graph.numNodes; i++{
        port := all[i]
        nbrs := make([]string, graph.numNodes-1)
        for j, x := 0, 0; j < graph.numNodes; j++ {
            if j != i {
                nbrs[x] = all[j]
                x++
            }
        }
        switch {
            case i < faults: byz = 1
            default: byz = 0
        }
        go peer.Client(port, nbrs, byz)
    }
    time.Sleep(2000*time.Millisecond) 
    fmt.Printf("Done!")
    os.Exit(0)
}
