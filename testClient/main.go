package main

import "github.com/goPirateBay/client"

func main() {
	serverschave := &client.ServerCache{}

	client.ListServerCotainsFile(serverschave, "5bf64a910e76804ecb697b156d77dd72c3661070")
}
