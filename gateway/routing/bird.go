package routing

import (
    "fmt"
    "os/exec"
)

type BIRDManager struct {
    birdSocket string
}

func NewBIRDManager(socketPath string) *BIRDManager {
    return &BIRDManager{birdSocket: socketPath}
}

// Запуск BIRDdaemon
func (b *BIRDManager) Start() error {
    cmd := exec.Command("bird", "-c", constBirdConfPath)
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start bird: %w", err)
    }
    return nil
}

// Выполнение команды через birdc
func (b *BIRDManager) ExecuteCommand(cmdArgs ...string) (string, error) {
    args := append([]string{"-s", b.birdSocket}, cmdArgs...)
    cmd := exec.Command("birdc", args...)
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("birdc error: %w, output: %s", err, output)
    }
    return string(output), nil
}

// Показать статус протоколов
func (b *BIRDManager) ShowProtocols() (string, error) {
    return b.ExecuteCommand("show", "protocols")
}

func RunBird() {
    bird := NewBIRDManager(constBirdSockPath)
    status, _ := bird.ShowProtocols()
    fmt.Println(status)
}


