package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	ping "https://github.com/Megabd/mockExam/grpc"
	"google.golang.org/grpc"
)

func main() {

	// If the file doesn't exist, create it or append to the file

	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()


	p := &peer{
		id:            ownPort,
		amountOfPings: make(map[int32]int32),
		clients:       make(map[int32]ping.PingClient),
		ctx:           ctx,
		wanted:        false,
		held:          false,
		timesAccessed: 0,
		amount: 		-1,
	}

	// Create listener tcp on port ownPort
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}
	grpcServer := grpc.NewServer()
	ping.RegisterPingServer(grpcServer, p)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	for i := 0; i < 3; i++ {
		port := int32(5000) + int32(i)

		if port == ownPort {
			continue
		}

		var conn *grpc.ClientConn
		fmt.Printf("Trying to dial: %v\n", port)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		c := ping.NewPingClient(conn)
		p.clients[port] = c
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		p.checkCommand(scanner.Text())
	}
}

type peer struct {
	ping.UnimplementedPingServer
	id            int32
	amountOfPings map[int32]int32
	clients       map[int32]ping.PingClient
	ctx           context.Context
	wanted        bool
	held          bool
	timesAccessed int32
	amount 			int32
}

func (p *peer) ReturnInfo(ctx context.Context, req *ping.Request) (*ping.ReturnInfoReply, error) {

	rep := &ping.ReturnInfoReply{Id: p.id, TimesAccessed: p.timesAccessed, Wanted: p.wanted, Held: p.held}
	return rep, nil
}

func (p *peer) checkCommand(command string){
	if (command == "Increment"){
		p.askPermission()
	}
	else {
		fmt.Println("Unknown command, try typing Increment")
	}

}

func (p *peer) askPermission() {
	p.wanted = true
	permission := false

	request := &ping.Request{Id: p.id}
	for id, client := range p.clients {


		returnInfoReply, err := client.ReturnInfo(p.ctx, request)
		if err != nil {
			fmt.Println("something went wrong")
		}
		if returnInfoReply.amount > p.amount{
			p.amount = returnInfoReply.amount
		}
		if returnInfoReply.Held {
			permission = false
			break
		} else if returnInfoReply.Wanted {
			if returnInfoReply.TimesAccessed < p.timesAccessed {
				permission = false
				break
			} else if returnInfoReply.TimesAccessed > p.timesAccessed {
				permission = true
			} else if returnInfoReply.TimesAccessed == p.timesAccessed {
				if returnInfoReply.Id < p.id {
					permission = false
					break
				} else if returnInfoReply.Id > p.id {
					permission = true
				}
			}
		} else if !returnInfoReply.Wanted {
			permission = true
		}
	}
	
	if permission {
		p.held = true
		amount = p.increment(p.amount)
		p.amount = amount
		fmt.Println(p.amount)

	} else if !permission {
		time.Sleep(1 * time.Second)
		p.askPermission()
		return
	}

	p.wanted = false
	p.held = false
	p.timesAccessed++

}

func (p *peer) increment (amount int) (newAmount int){
 
	newAmount = amount +1;
	return newAmount;


}
