package main

import (
    "os"
    "flag"
    "fmt"
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
        lllcserver.URL = *host+"/"+"compile"
        fmt.Println("url:", lllcserver.URL)
    }

    if *client{
        CheckMakeDir(lllcserver.TMP)
        tocompile := flag.Args()
        fmt.Println("to compile:", tocompile)
        lllcserver.RunClient(tocompile)
    }else {
        CheckMakeDir(".tmp")
        addr := ""
        if *localOnly{
            addr = "localhost"
        }
        addr += ":"+strconv.Itoa(*port)
        lllcserver.StartServer(addr)
    }
}

func CheckMakeDir(dir string){
   _, err := os.Stat(dir)
   if err != nil{
       err := os.Mkdir(dir, 0777)  //wtf!
       if err != nil{
            fmt.Println("Could not make directory. Exiting", err)
            os.Exit(0)
       }
   }
}
