package main

import (
    "bytes"
    "encoding/json"
    "path/filepath"
    "log"
    "fmt"
    "hash/crc32"
    "io/ioutil"
    "net/http"
    "sort"
    "strconv"
    "sync"
    "time"
    "os"
    "context"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var isDebug = true

type HashRing struct {
    nodes       []int
    nodeMap     map[int]string
    replicas    int
    mu          sync.RWMutex
}

func NewHashRing(replicas int) *HashRing {
    return &HashRing{
        replicas: replicas,
        nodeMap:  make(map[int]string),
    } 
}

func (h *HashRing) AddNode(node string) {
    h.mu.Lock()
    defer h.mu.Unlock()
    for i := 0; i < h.replicas; i++ {
        hash := int(crc32.ChecksumIEEE([]byte(node + strconv.Itoa(i))))
        h.nodes = append(h.nodes, hash)
        h.nodeMap[hash] = node
    }
    sort.Ints(h.nodes)
    // print nodes
    fmt.Println(h.nodes)
}

func (h *HashRing) RemoveNode(node string) {
    h.mu.Lock()
    defer h.mu.Unlock()
    for i := 0; i < h.replicas; i++ {
        hash := int(crc32.ChecksumIEEE([]byte(node + strconv.Itoa(i))))
        idx := sort.Search(len(h.nodes), func(i int) bool {
            return h.nodes[i] == hash
        })
        if idx < len(h.nodes) && h.nodes[idx] == hash {
            h.nodes = append(h.nodes[:idx], h.nodes[idx+1:]...)
            delete(h.nodeMap, hash)
        }
    }
}

func (h *HashRing) GetNodes(key string, count int) []string {
    h.mu.RLock()
    defer h.mu.RUnlock()
    if len(h.nodes) == 0 {
        return nil
    }
    hash := int(crc32.ChecksumIEEE([]byte(key)))
    idx := sort.Search(len(h.nodes), func(i int) bool {
        return h.nodes[i] >= hash
    })
    if idx == len(h.nodes) {
        idx = 0
    }
    nodes := make([]string, 0, count)
    for i := 0; i < count; i++ {
        nodes = append(nodes, h.nodeMap[h.nodes[(idx+i)%len(h.nodes)]])
    }
    return nodes
}

func homeDir() string {
    if h := os.Getenv("HOME"); h != "" {
        return h
    }
    return os.Getenv("USERPROFILE") // windows
}

func k8setup() {
        // Load the kubeconfig file to connect to the cluster
        kubeconfig := filepath.Join(
            homeDir(), ".kube", "config",
        )
        // fmt.Println(kubeconfig)

        config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
            log.Fatalf("Error building kubeconfig: %v", err)
        }
    
        // Create the clientset
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
            log.Fatalf("Error creating Kubernetes client: %v", err)
        }
    
        // Discover pods of another deployment
        pods, err := discoverPods(clientset, "ch-demo", "app=cache-service")
        if err != nil {
            log.Fatalf("Error discovering pods: %v", err)
        }
    
        for _, pod := range pods {
            fmt.Println(pod)
        }
}

// discoverPods lists the pods in a given namespace with a specific label selector
func discoverPods(clientset *kubernetes.Clientset, namespace, labelSelector string) ([]string, error) {
    pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
        LabelSelector: labelSelector,
    })
    fmt.Println(pods)
    if err != nil {
        return nil, err
    }

    podIPs := make([]string, 0, len(pods.Items))
    for _, pod := range pods.Items {
        podIPs = append(podIPs, pod.Status.PodIP)
    }
    return podIPs, nil
}

func discoverNodes() []string {
    // Implement service discovery logic here
    // kubectl get pods --selector app=cache-node -o jsonpath='{.items[*].status.podIP}'
    // For example, using Kubernetes API to list pods with a specific label
    return []string{"http://localhost:8080", "http://localhost:8081", "http://localhost:8082"}
}

func monitorNodes(ring *HashRing) {
    for {
        nodes := discoverNodes()
        fmt.Println(nodes)
        ring.mu.Lock()
        for _, node := range nodes {
            hash := int(crc32.ChecksumIEEE([]byte(node)))
            if _, ok := ring.nodeMap[hash]; !ok {
                ring.AddNode(node)
            }
        }
        for node := range ring.nodeMap {
            if !contains(nodes, ring.nodeMap[node]) {
                ring.RemoveNode(ring.nodeMap[node])
                redistributeData(ring, ring.nodeMap[node])
            }
        }
        ring.mu.Unlock()
        time.Sleep(10 * time.Second)
    }
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

func redistributeData(ring *HashRing, failedNode string) {
    // Fetch all keys from the remaining nodes
    keys := fetchKeysFromRemainingNodes(ring, failedNode)
    for _, key := range keys {
        value := fetchValueFromRemainingNodes(ring, failedNode, key)
        newNodes := ring.GetNodes(key, 3)
        for _, newNode := range newNodes {
            storeValueOnNode(newNode, key, value)
        }
    }
}

func fetchKeysFromRemainingNodes(ring *HashRing, failedNode string) []string {
    // Implement logic to fetch all keys from the remaining nodes
    return []string{}
}

func fetchValueFromRemainingNodes(ring *HashRing, failedNode, key string) string {
    // Implement logic to fetch the value of a key from the remaining nodes
    return ""
}

func storeValueOnNode(node, key, value string) {
    url := fmt.Sprintf("%s/set", node)
    data, _ := json.Marshal(map[string]string{key: value})
    http.Post(url, "application/json", bytes.NewBuffer(data))
}

func main() {
    ring := NewHashRing(3)
    go monitorNodes(ring)
    setHandlers(ring)
    k8setup()

    http.ListenAndServe(":8080", nil)
}

func setHandlers(ring *HashRing) {
        http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
        var kv map[string]string
        json.NewDecoder(r.Body).Decode(&kv)
        for k, v := range kv {
            nodes := ring.GetNodes(k, 3)
            for _, node := range nodes {
                url := fmt.Sprintf("%s/set", node)
                data, _ := json.Marshal(map[string]string{k: v})
                http.Post(url, "application/json", bytes.NewBuffer(data))
            }
        }
        w.WriteHeader(http.StatusOK)
    })

    http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
        key := r.URL.Query().Get("key")
        nodes := ring.GetNodes(key, 3)
        for _, node := range nodes {
            url := fmt.Sprintf("%s/get?key=%s", node, key)
            resp, err := http.Get(url)
            if err == nil && resp.StatusCode == http.StatusOK {
                body, _ := ioutil.ReadAll(resp.Body)
                w.Write(body)
                return
            }
        }
        http.NotFound(w, r)
    })

    http.HandleFunc("/nodes", func(w http.ResponseWriter, r *http.Request) {
        ring.mu.RLock()
        defer ring.mu.RUnlock()
        nodes := make([]string, 0, len(ring.nodeMap))
        for _, node := range ring.nodeMap {
            nodes = append(nodes, node)
        }
        json.NewEncoder(w).Encode(nodes)
    })

    http.HandleFunc("/addNode", func(w http.ResponseWriter, r *http.Request) {
        var node struct {
            Address string `json:"address"`
        }
        json.NewDecoder(r.Body).Decode(&node)
        ring.AddNode(node.Address)
        w.WriteHeader(http.StatusOK)
    })

    http.HandleFunc("/removeNode", func(w http.ResponseWriter, r *http.Request) {
        var node struct {
            Address string `json:"address"`
        }
        json.NewDecoder(r.Body).Decode(&node)
        ring.RemoveNode(node.Address)
        w.WriteHeader(http.StatusOK)
    })

    http.Handle("/", http.FileServer(http.Dir("./static")))
}