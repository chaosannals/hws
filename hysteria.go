package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cihub/seelog"
	"github.com/kardianos/service"
)

type HysteriaService struct {
	root    string
	logger  seelog.LoggerInterface
	process *os.Process
}

func NewHysteriaService() *HysteriaService {
	root, err := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Printf("当前路径: %v\n", root)
	if err != nil {
		return nil
	}
	// 设置当前工作路径为 Hysteria 程序所在路径
	err = os.Chdir(root)
	if err != nil {
		return nil
	}
	// 加载设置。
	path := filepath.Join(root, "seelog.xml")
	logger, err := seelog.LoggerFromConfigAsFile(path)
	if err != nil {
		fmt.Printf("seelog.xml 失败: %v\n", err)
		return nil
	}
	seelog.ReplaceLogger(logger)
	return &HysteriaService{
		root:    root,
		logger:  logger,
		process: nil,
	}
}

//Start 开始
func (p *HysteriaService) Start(s service.Service) error {
	p.logger.Info("服务启动")

	// 设置工作目录为 hysteria 目录
	// wkdir := filepath.Join(p.root, "hysteria")
	wkdir := p.root
	e, err := IsExists(wkdir)
	if err != nil {
		return err
	}
	if e {
		err := os.Chdir(wkdir)
		if err != nil {
			return err
		}
	}

	// 启动 Hysteria
	p.process, err = os.StartProcess("./hysteria.exe", []string{
		// "./hysteria.exe",
	}, &os.ProcAttr{
		Dir:   wkdir,
		Env:   os.Environ(),
		Files: []*os.File{nil, nil, nil},
	})
	if err != nil {
		p.logger.Error(err)
		return err
	}
	return nil
}

//Stop 停止
func (p *HysteriaService) Stop(s service.Service) error {
	p.logger.Info("服务关闭")
	// 设置工作目录为 hysteria 目录
	// wkdir := filepath.Join(p.root, "hysteria")
	wkdir := p.root
	e, err := IsExists(wkdir)
	if err != nil {
		return err
	}
	if e {
		err := os.Chdir(wkdir)
		if err != nil {
			return err
		}
	}
	// 关闭 Hysteria
	p.process.Kill()
	return nil
}
