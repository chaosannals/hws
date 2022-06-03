package main

import (
	"fmt"
	"os"

	"github.com/kardianos/service"
)

func IsExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func main() {
	svcConfig := &service.Config{
		Name:        "nws",
		DisplayName: "Hysteria Windows Service",
		Description: "Hysteria Windows Service.",
	}
	hs := NewHysteriaService()
	s, err := service.New(hs, svcConfig)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	if len(os.Args) > 1 {
		if os.Args[1] == "install" {
			// err = InitHysteria()
			// if err != nil {
			// 	fmt.Printf("%v\n", err)
			// 	return
			// }
			s.Install()
			s.Start()
			fmt.Println("服务安装成功")
			return
		}
		if os.Args[1] == "uninstall" {
			s.Stop()
			s.Uninstall()
			fmt.Println("服务卸载成功")
			return
		}
	}
	err = s.Run()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
