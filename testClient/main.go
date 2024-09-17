package main

import "github.com/goPirateBay/client"

func main() {
	serverschave := &client.ServerCache{}

	client.ListServerCotainsFile(serverschave, "2d3e2c4412a753a2a60c1dd3574757adfc780d46")
}
