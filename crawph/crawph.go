package crawph

import (
	"fmt"
	"sync"
	"time"

	"github.com/alexmar07/crawler-go/graph"
	"github.com/alexmar07/crawler-go/queue"
	"github.com/alexmar07/crawler-go/scraper"
)

type Crawph struct {
	maxWorkers        int
	graph             *graph.Graph
	queue             *queue.Queue
	activeWorkers     int
	activeWorkersMu   sync.RWMutex
	shouldTerminate   bool
	shouldTerminateMu sync.RWMutex
	wg                sync.WaitGroup
}

func NewCrawph(maxWorkers int) *Crawph {
	return &Crawph{
		maxWorkers:      maxWorkers,
		graph:           graph.NewGraph(),
		queue:           queue.NewQueue(),
		wg:              sync.WaitGroup{},
		shouldTerminate: false,
	}
}

func (c *Crawph) IncreaseActiveWorkers() {

	c.activeWorkersMu.Lock()
	defer c.activeWorkersMu.Unlock()

	c.activeWorkers++
	fmt.Println("[Increment] Numero dei worker attivi: ", c.activeWorkers)
}

func (c *Crawph) GetActiveWorkers() int {
	c.activeWorkersMu.RLock()
	defer c.activeWorkersMu.RUnlock()

	return c.activeWorkers
}

func (c *Crawph) HasActiveWorkers() bool {
	return c.GetActiveWorkers() == 0
}

func (c *Crawph) DecreaseActiveWorkers() {

	c.activeWorkersMu.Lock()
	defer c.activeWorkersMu.Unlock()

	c.activeWorkers--

	fmt.Println("[Decrement] Numero dei worker attivi: ", c.activeWorkers)

}

func (c *Crawph) SetShouldTerminate(terminate bool) {

	c.shouldTerminateMu.Lock()
	defer c.shouldTerminateMu.Unlock()

	c.shouldTerminate = terminate
}

func (c *Crawph) GetShouldTerminate() bool {

	c.shouldTerminateMu.RLock()
	defer c.shouldTerminateMu.RUnlock()

	return c.shouldTerminate
}

func (c *Crawph) Start(urls []string) {

	for _, url := range urls {
		fmt.Println("Aggiunge in coda ", url)
		c.queue.Enqueue(url)
	}

	for i := 0; i < c.maxWorkers; i++ {
		fmt.Println("Crea il worker ", i)
		c.wg.Add(1)
		go c.worker(i)
	}

	go c.monitorWorkers()

	c.wg.Wait()
}

func (c *Crawph) worker(id int) {

	defer c.wg.Done()

	for {

		c.IncreaseActiveWorkers()

		shouldTerminate := c.GetShouldTerminate()

		fmt.Printf("[Worker %d] Terminazione: %s \n", id, shouldTerminate)

		url, ok := c.queue.Dequeue(&shouldTerminate)

		c.DecreaseActiveWorkers()

		if !ok {
			fmt.Printf("Il worker %d è terminato", id)
			return
		}

		fmt.Printf("[Worker %d] Ho recuperato l'url: %s \n", id, url)
		c.Crawling(url)
	}
}

func (c *Crawph) Crawling(url string) {

	startVertex := c.graph.AddVertex(url)

	fmt.Println("Vertice iniziale ", startVertex.FullUrl)

	scraper := scraper.NewScraper(url)

	_, err := scraper.StartDiscovered()

	if err != nil {
		fmt.Println(err)
		return
	}

	if scraper.CountDiscoveredLinks() == 0 {

		fmt.Println("Non sono stati scoperti ulteriori link.")

		return
	}

	links := scraper.GetDiscoveredLinks()

	for _, link := range links {

		c.queue.Enqueue(link)

		fmt.Println("Link scoperto da mettere in coda: ", link)

		newVertex := c.graph.AddVertex(link)

		c.graph.AddEdge(startVertex, newVertex)

		fmt.Printf("Crea arco tra %s e %s \n", startVertex.FullUrl, newVertex.FullUrl)
	}
}

func (c *Crawph) monitorWorkers() {

	fmt.Println("Start monitoring..")

	for {
		time.Sleep(100 * time.Millisecond)

		if c.HasActiveWorkers() && c.queue.IsEmpty() {
			c.SetShouldTerminate(true)
			c.queue.BroadcastTermination()
			fmt.Println("Terminazione dell'esecuzione.")
			return
		}
	}
}
