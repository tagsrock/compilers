package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/project-douglas/lllc-server"
	"os"
	"strconv"
)

// simple lllc-server and cli
func main() {
	client := flag.Bool("c", false, "specify files to compile, separated by space. this must come last")
	host := flag.String("h", "", "specify the host and port")
	localOnly := flag.Bool("local", false, "only listen internally")
	port := flag.Int("port", 9999, "listen port")
	nonet := flag.Bool("no-net", false, "do you have lll locally?")
	lang := flag.String("lang", "lll", "language to compile")

	flag.Parse()

	langConfig := lllcserver.Languages[*lang]

	if *host != "" {
		langConfig.URL = *host + "/" + "compile"
		fmt.Println("url:", langConfig.URL)
	}

	if *client {
		lllcserver.CheckMakeDir(lllcserver.TMP)
		tocompile := flag.Args()[0]
		fmt.Println("to compile:", tocompile)
		if *nonet {
			b, err := lllcserver.CompileWrapper(tocompile, *lang)
			if err != nil {
				fmt.Println("failed to compile!", err)
				os.Exit(0)
			}
			fmt.Println("bytecode:", hex.EncodeToString(b))
		} else {
			lllcserver.Compile(tocompile)
		}
	} else {
		lllcserver.CheckMakeDir(lllcserver.ServerTmp)
		addr := ""
		if *localOnly {
			addr = "localhost"
		}
		addr += ":" + strconv.Itoa(*port)
		lllcserver.StartServer(addr)
	}
}
