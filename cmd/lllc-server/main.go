package main

import (
    "os"
    "flag"
    "fmt"
    "strconv"
    "encoding/hex"
    "github.com/project-douglas/lllc-server"
)



// simple lllc-server and cli
func main(){
    client := flag.Bool("c", false, "specify files to compile, separated by space. this must come last")
    host := flag.String("h", "", "specify the host and port")
    localOnly := flag.Bool("local", false, "only listen internally")
    port := flag.Int("port", 9999, "listen port")
    nonet := flag.Bool("no-net", false, "do you have lll locally?")

    flag.Parse()


    if *host != ""{
        lllcserver.URL = *host+"/"+"compile"
        fmt.Println("url:", lllcserver.URL)
    }

    if *client{
        lllcserver.CheckMakeDir(lllcserver.TMP)
        tocompile := flag.Args()
        fmt.Println("to compile:", tocompile)
        if *nonet{
            b, err := lllcserver.CompileLLLWrapper(tocompile[0])
            if err != nil{
                fmt.Println("failed to compile!", err)
                os.Exit(0)
            }
            fmt.Println("bytecode:", hex.EncodeToString(b))
        } else{
            lllcserver.RunClient(tocompile)
        }
    }else {
        lllcserver.CheckMakeDir(lllcserver.ServerTmp)
        addr := ""
        if *localOnly{
            addr = "localhost"
        }
        addr += ":"+strconv.Itoa(*port)
        lllcserver.StartServer(addr)
    }
}
