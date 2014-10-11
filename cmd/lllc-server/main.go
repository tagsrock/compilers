package main

import (
    "flag"
    "path"
    "strconv"
    "github.com/project-douglas/lllc-server"
)



// simple lllc-server and cli
func main(){
    client := flag.Bool("c", false, "specify files to compile, separated by space. this must come last")
    host := flag.String("h", "", "specify the host and port")
    localOnly := flag.Bool("local", false, "only listen internally")
    port := flag.Int("port", 9999, "listen port")

    flag.Parse()

    if *host != ""{
        lllcserver.URL = path.Join(*host, "compile")
    }

    if *client{
        tocompile := flag.Args()
        lllcserver.RunClient(tocompile)
    }else {
        addr := ""
        if *localOnly{
            addr = "localhost"
        }
        addr += ":"+strconv.Itoa(*port)
        lllcserver.StartServer(addr)
    }
}
