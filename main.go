package main

import (
	"fmt"
	"go-redis/repo"
	"go-redis/server"
)

func main() {
	go func() {
		s := server.NewServer(server.Config{})
		fmt.Println("Server started.....")
		s.Start()
	}()
	// time.Sleep(time.Second)
	// client := client.NewClient("localhost:50001") // blockhere
	// for i := 0; i < 10; i++ {
	// 	err := client.Set(context.Background(), "user", "1")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }
	// time.Sleep(time.Second)
	// fmt.Println(repo.KvString)
	// client.Get(context.Background(), "user")
	select {}
}

func init() {
	// inititialize memory data
	repo.InitKV()
	repo.InitKVList()
}
