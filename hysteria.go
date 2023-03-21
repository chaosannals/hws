package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

func (p *HysteriaService) startProcess() error {
	// 设置工作目录为 hysteria 目录
	// wkdir := filepath.Join(p.root, "hysteria")
	p.logger.Info("设置工作目录：")
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

	// 标准输出
	p.logger.Info("打开标准输出：")
	outr, outw, err := os.Pipe()
	if err != nil {
		p.logger.Error(err)
		return err
	}
	go func() {
		reader := bufio.NewReader(outr)
		b := make([]byte, 1024)
		for {
			time.Sleep(time.Second)
			n, err := reader.Read(b)
			if err != nil {
				p.logger.Error(err)
			} else {
				p.logger.Info(string(b[:n]))
			}
		}
	}()

	// 标准错误
	p.logger.Info("打开标准错误：")
	errr, errw, err := os.Pipe()
	if err != nil {
		p.logger.Error(err)
		return err
	}
	go func() {
		reader := bufio.NewReader(errr)
		b := make([]byte, 1024)
		for {
			time.Sleep(time.Second)
			n, err := reader.Read(b)
			if err != nil {
				p.logger.Error(err)
			} else {
				p.logger.Error(string(b[:n]))
			}
		}
	}()

	// 启动 Hysteria
	p.logger.Info("启动进程：")
	p.process, err = os.StartProcess("./hysteria.exe", []string{}, &os.ProcAttr{
		Dir:   wkdir,
		Env:   os.Environ(),
		Files: []*os.File{nil, outw, errw},
	})

	return nil
}

func (p *HysteriaService) checkProcess() (error, bool) {
	process, err := os.FindProcess(p.process.Pid)
	if err != nil {
		p.logger.Warn("查找进程错误：", p.process.Pid, err)
		return err, false
	}

	if process == nil {
		p.logger.Warn("没有找到进程：", p.process.Pid)
		return nil, false
	}

	// Windows 下判定进程存活，注：这个方法阻塞等待。
	ps, err := process.Wait()
	if err != nil {
		p.logger.Warn("进程状态错误：", p.process.Pid, err)
		return err, false
	}

	if ps.Exited() {
		p.logger.Warn("进程结束了：", p.process.Pid)
		return nil, false
	}

	// 以下的 linux 的判定进程存活的方式
	// err = process.Signal(syscall.Signal(0))
	// if err != nil {
	// 	   p.logger.Warn("进程接收信号错误：", err)
	//     return err, true
	// }

	return nil, true
}

//Start 开始
func (p *HysteriaService) Start(s service.Service) error {
	p.logger.Info("服务启动")

	err := p.startProcess()
	if err != nil {
		p.logger.Error(err)
		return err
	}

	// 进程常驻守护
	go func() {
		for {
			time.Sleep(time.Second * 4)

			if err, ok := p.checkProcess(); ok {
				p.logger.Info("进程存活：", p.process.Pid)
				time.Sleep(time.Second * 14)
			} else {
				p.logger.Error("进程不存在：", p.process.Pid, err)
				err = p.startProcess()
				if err != nil {
					p.logger.Error("重新启动进程失败：", err)
					time.Sleep(time.Second * 4)
				}
			}
		}
	}()

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
