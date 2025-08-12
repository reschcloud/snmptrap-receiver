package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/gosnmp/gosnmp"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
	"gopkg.in/ini.v1"
	"gopkg.in/natefinch/lumberjack.v2"
)

var version = "dev"

const serviceName = "SNMPTrapReceiver"

var elog debug.Log

type myservice struct{}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	s <- svc.Status{State: svc.StartPending}
	go runTrapListener()
	s <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				s <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			default:
			}
		}
	}

	s <- svc.Status{State: svc.StopPending}
	return false, 0
}

func runTrapListener() {
	exePath, _ := os.Executable()
	baseDir := filepath.Dir(exePath)

	// Logging aktivieren
	logFile, err := os.OpenFile(filepath.Join(baseDir, "service.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
	}
	log.Printf("Dienst gestartet.")
	log.Printf("EXE Pfad: %s", exePath)

	cfgPath := filepath.Join(baseDir, "config.ini")
	cfg, err := ini.Load(cfgPath)
	if err != nil {
		log.Printf("Konnte config.ini nicht laden: %v", err)
		return
	}

	port := cfg.Section("snmp").Key("port").MustInt(1162)
	csvPath := cfg.Section("snmp").Key("csv_file").MustString("snmp_traps.csv")
	if !filepath.IsAbs(csvPath) {
		csvPath = filepath.Join(baseDir, csvPath)
	}
	log.Printf("CSV Pfad: %s", csvPath)

	// CSV Rotation aus ini
	rotateSize := cfg.Section("snmp").Key("rotate_max_mb").MustInt(1)
	rotateBackups := cfg.Section("snmp").Key("rotate_max_files").MustInt(0)
	rotateAge := cfg.Section("snmp").Key("rotate_max_age").MustInt(0)

	log.Printf("CSV Rotation: max %d MB, max %d Dateien, max %d Tage",
		rotateSize, rotateBackups, rotateAge)

	lj := &lumberjack.Logger{
		Filename:   csvPath,
		MaxSize:    rotateSize,
		MaxBackups: rotateBackups,
		MaxAge:     rotateAge,
		Compress:   false,
	}
	writer := csv.NewWriter(lj)
	writer.Write([]string{"timestamp", "source", "oid", "type", "value"})
	writer.Flush()

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Printf("Fehler beim Oeffnen von Port %d: %v", port, err)
		return
	}
	defer pc.Close()

	buf := make([]byte, 2048)
	for {
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			log.Printf("Fehler beim Lesen: %v", err)
			continue
		}
		pkt := buf[:n]
		snmp := &gosnmp.GoSNMP{}
		trap, err := snmp.UnmarshalTrap(pkt, false)
		if err != nil {
			log.Printf("Fehler beim Parsen von Trap: %v", err)
			continue
		}
		for _, v := range trap.Variables {
			valStr := fmt.Sprintf("%v", v.Value)
			typ := v.Type.String()
			writer.Write([]string{
				time.Now().Format(time.RFC3339),
				addr.String(),
				v.Name,
				typ,
				valStr,
			})
		}
		writer.Flush()
	}
}

func isInteractive() (bool, error) {
	return svc.IsAnInteractiveSession()
}

func main() {
	interactive, err := isInteractive()
	if err != nil {
		log.Fatalf("Fehler: %v", err)
	}

	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
		case "install":
			m, err := mgr.Connect()
			if err != nil {
				log.Fatalf("mgr.Connect() error: %v", err)
			}
			defer m.Disconnect()
			exePath, err := os.Executable()
			if err != nil {
				log.Fatalf("os.Executable() error: %v", err)
			}
			s, err := m.CreateService(serviceName, exePath, mgr.Config{
				DisplayName: "SNMP Trap Receiver Service",
				StartType:   mgr.StartAutomatic,
			})
			if err != nil {
				log.Fatalf("CreateService() error: %v", err)
			}
			defer s.Close()
			log.Println("Dienst installiert.")
			return

		case "uninstall":
			m, err := mgr.Connect()
			if err != nil {
				log.Fatalf("mgr.Connect() error: %v", err)
			}
			defer m.Disconnect()
			s, err := m.OpenService(serviceName)
			if err != nil {
				log.Fatalf("OpenService() error: %v", err)
			}
			defer s.Close()
			err = s.Delete()
			if err != nil {
				log.Fatalf("Delete() error: %v", err)
			}
			log.Println("Dienst entfernt.")
			return
		}
	}

	if interactive {
		elog = debug.New(serviceName)
		runTrapListener()
	} else {
		elog, _ = eventlog.Open(serviceName)
		defer elog.Close()
		svc.Run(serviceName, &myservice{})
	}
}
