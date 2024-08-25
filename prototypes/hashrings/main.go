package main

import (
	"fmt"
	"prototypes/hashrings/basic"
)

func main() {
	// GENERATE HASH RING
	hr := basic.NewHashRing(3)

	hr.AddNode("NodeA")
	hr.AddNode("NodeB")
	hr.AddNode("NodeC")

	keys := []string{"Key1", "Key2", "Key3", "Key4", "Key5"}

	for _, key := range keys {
		node := hr.GetNode(key)
		fmt.Printf("Key %s is assigned to node %s\n", key, node)
	}

	hr.RemoveNode("NodeB")

	fmt.Println("After removing NodeB:")
	for _, key := range keys {
		node := hr.GetNode(key)
		fmt.Printf("Key %s is assigned to node %s\n", key, node)
	}
}
