package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct{
	Addr string
	Weight int
	Connections int64
}

type ServerPool struct {
	servers sync.Map
}

func (sp *ServerPool) AddServer(addr string , weight int){
	sp.servers.Store(addr, &Server{Addr: addr, Weight: weight})
	
}

func (sp *ServerPool) RemoveServer(addr string){
	sp.servers.Delete(addr)
}

func (sp *ServerPool) GetServers() []*Server{
	servers := make([]*Server, 0)
	sp.servers.Range(func(key, value interface{}) bool {
		servers = append(servers, value.(*Server))
		return true
	})
	return servers
}

// least connections algorithm
func LeastConnections(servers []*Server) *Server{
	var bestServer *Server
	leastConns := int64(1<<63 - 1)
	for _, server := range servers{
		if server.Connections < leastConns{
			leastConns = server.Connections
			bestServer = server
		}
	}
	return bestServer
}


func healthCheck(server *Server) bool{
	url, err:= url.Parse("http://"+server.Addr)
	if err != nil{
		return false
	}
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := client.Get(url.String())
	if err != nil{
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func runHealthCheck(sp *ServerPool, interval time.Duration){
	for {
		time.Sleep(interval)
		sp.servers.Range(func(addr, server interface{}) bool {
			healthy := healthCheck(server.(*Server))
			if !healthy{
				//! remove server from pool once removed, the server will not be added back to the pool unless the server is restarted
				//! this is to prevent the load balancer from sending requests to a server that is not healthy
				//! if the server is restarted, it will be added back to the pool
				//? to overcome this limitations you can have unhealthy servers in a separate pool
				//? and move them back to the main pool once they are healthy
				sp.RemoveServer(addr.(string))
			}
			return true
		})
	}
}

type LoadBalancer struct{
	serverPool *ServerPool
	algorithm func(servers []*Server) *Server
	interval time.Duration
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, req *http.Request){
	servers := lb.serverPool.GetServers()
	if len(servers) == 0{
		http.Error(w, "No servers available", http.StatusServiceUnavailable)
		return
	}

	bestServer := lb.algorithm(servers)
	if bestServer==nil{
		http.Error(w, "No server available", http.StatusServiceUnavailable)
		return
	}
	atomic.AddInt64(&bestServer.Connections, 1)
	defer atomic.AddInt64(&bestServer.Connections, -1)

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host: bestServer.Addr,
	})
	proxy.ServeHTTP(w, req)
	
}


func main(){
	serverPool := &ServerPool{}
	algorithm := LeastConnections
	interval := time.Second * 5

	serverPool.AddServer("localhost:8081", 1)
	serverPool.AddServer("localhost:8082", 2)
	serverPool.AddServer("localhost:8083", 3)

	lb := &LoadBalancer{
		serverPool: serverPool,
		algorithm: algorithm,
		interval: interval,
	}

	go runHealthCheck(serverPool, interval)

	http.Handle("/", lb)
	fmt.Println("Starting Load Balancer server on port 8080....")
	log.Fatal(http.ListenAndServe(":8080", nil))
}