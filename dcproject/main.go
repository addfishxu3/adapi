package main

import routers "dcproject/routers"

func main() {
	server := routers.SetupRoute()
	server.Run(":8080") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
