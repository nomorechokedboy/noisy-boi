package main

import (
	"bufio"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/valyala/fasthttp"
)

var msgChan chan string

type SSEClient struct {
	channel    []chan string
	connection uint
}

type SSEPool struct {
	clients map[int]*SSEClient
	mu      sync.Mutex
}

func NewSSEClient() *SSEClient {
	return &SSEClient{channel: []chan string{make(chan string)}, connection: 1}
}

func NewSSEPool() *SSEPool {
	return &SSEPool{clients: make(map[int]*SSEClient)}
}

func (p *SSEPool) addClient(id int) (*chan string, uint) {
	p.mu.Lock()
	defer func() {
		fmt.Println("Clients in add: ", p.clients)
		for k, v := range p.clients {
			fmt.Printf("Key: %d, value: %d\n", k, v.connection)
		}
		p.mu.Unlock()
	}()

	c, ok := p.clients[id]
	if !ok {
		client := NewSSEClient()
		p.clients[id] = client
		return &client.channel[client.connection-1], client.connection
	}

	newCh := make(chan string)
	c.channel = append(c.channel, newCh)
	c.connection++
	return &newCh, c.connection
}

func (p *SSEPool) removeClient(id int, conn uint) {
	p.mu.Lock()
	defer func() {
		fmt.Println("Clients in remove: ", p.clients)
		for k, v := range p.clients {
			fmt.Printf("Key: %d, value: %d", k, v.connection)
		}
		p.mu.Unlock()
	}()

	c, ok := p.clients[id]
	if !ok {
		return
	}

	/* go func() {
		for range c.channel {
		}
	}() */
	// Check this logic again
	close(c.channel[conn])
	c.channel = append(c.channel[:conn], c.channel[conn+1:]...)
	c.connection--
	if c.connection <= 0 {
		delete(p.clients, id)
	}
}

func (p *SSEPool) broadcast(id int, data, event string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	c, ok := p.clients[id]
	if !ok {
		return
	}

	for _, ch := range c.channel {
		ch <- fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)
	}
}

func (p *SSEPool) sseHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.ErrBadRequest.Code, fiber.ErrBadRequest.Error())
	}

	ctx := c.Context()
	fmt.Println("Client connected")
	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set(fiber.HeaderConnection, "keep-alive")
	c.Set(fiber.HeaderTransferEncoding, "chunked")

	ctx.
		SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			msgChan, chId := p.addClient(id)
			fmt.Println("Clients:", p.clients, len(p.clients))

			defer func() {
				p.removeClient(id, chId-1)
				fmt.Println("Client closed connection")
				fmt.Println("Clients: ", p.clients)
			}()

			for {
				select {
				case message := <-*msgChan:
					fmt.Println("case message... sending message")
					fmt.Println(message)
					fmt.Fprintf(w, "data: %s\n\n", message)

					w.Flush()
				default:
					if err := w.Flush(); err != nil {
						fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)
						return
					}
				}
			}
		}))

	return nil
}

func (p *SSEPool) getTime(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.ErrBadRequest.Code, fiber.ErrBadRequest.Error())
	}

	p.broadcast(id, time.Now().Format("15:04:05"), "time")
	return c.JSON("Yeet")
}

func main() {
	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())
	ssePool := NewSSEPool()

	app.Get("/event/:id<int>", ssePool.sseHandler)
	app.Post("/time/:id<int>", ssePool.getTime)

	app.Listen(":3000")
}
