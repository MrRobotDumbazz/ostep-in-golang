package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type PrcoessState int

const (
	StateReady PrcoessState = iota
	StateRunning
	StateBlocked
	StateDone
)

func (s PrcoessState) String() string {
	switch s {
	case StateReady:
		return "READY"
	case StateRunning:
		return "RUNNING"
	case StateBlocked:
		return "BLOCKED"
	case StateDone:
		return "DONE"
	default:
		return "UNKNOWN"
	}
}

type Instruction struct {
	Type     string
	Duration int
}

type Process struct {
	ID           int
	Instructions []Instruction
	PC           int
	State        PrcoessState
	IOTimeLeft   int
}

type Simulator struct {
	Processes      []*Process
	CurrentTime    int
	IOLength       int
	SwitchBehavior string
	PrintState     bool
}

func NewSimulator() *Simulator {
	return &Simulator{
		Processes:      make([]*Process, 0),
		CurrentTime:    0,
		IOLength:       5,
		SwitchBehavior: "SWITCH_ON_IO",
		PrintState:     false,
	}
}

func (s *Simulator) AddProcess(processStr string) error {
	parts := strings.Split(processStr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format process: %s", processStr)
	}

	instructions := make([]Instruction, 0)
	instStr := parts[1]

	for _, char := range instStr {
		switch char {
		case 'c':
			instructions = append(instructions, Instruction{Type: "cpu", Duration: 1})
		case 'i':
			instructions = append(instructions, Instruction{Type: "io", Duration: s.IOLength})
		}
	}

	process := &Process{
		ID:           len(s.Processes),
		Instructions: instructions,
		PC:           0,
		State:        StateReady,
		IOTimeLeft:   0,
	}
	s.Processes = append(s.Processes, process)
	return nil
}

func (s *Simulator) GetReadyProcess() *Process {
	for _, p := range s.Processes {
		if p.State == StateReady {
			return p
		}
	}
	return nil
}
func (s *Simulator) AllProcessesDone() bool {
	for _, p := range s.Processes {
		if p.State != StateDone {
			return false
		}
	}
	return true
}

func (s *Simulator) UpdateIOProcesses() {
	for _, p := range s.Processes {
		if p.State == StateBlocked {
			p.IOTimeLeft--
			if p.IOTimeLeft <= 0 {
				p.State = StateReady
				p.PC++
			}
		}
	}
}

func (s *Simulator) PrintCurrentState(runningProcess *Process) {
	if !s.PrintState {
		return
	}

	fmt.Printf("Time %3d", s.CurrentTime)

	for i, p := range s.Processes {
		if i > 0 {
			fmt.Print(" ")
		}

		if runningProcess != nil && p.ID == runningProcess.ID {
			fmt.Printf("%s:%-7s", p.State.String(), "RUN")
		} else {
			fmt.Printf("%s:%-7s", p.State.String(), "")
		}
	}

	if runningProcess != nil && runningProcess.PC < len(runningProcess.Instructions) {
		inst := runningProcess.Instructions[runningProcess.PC]
		fmt.Printf(" [%s]", strings.ToUpper(inst.Type))
	}

	fmt.Println()
}

func (s *Simulator) Run() {
	fmt.Println("Симуляция выполнения процессов:")
	fmt.Println("================================")

	var currentProcess *Process

	for !s.AllProcessesDone() {
		// Обновляем процессы, выполняющие I/O
		s.UpdateIOProcesses()

		// Если нет текущего процесса или он заблокирован/завершен, выбираем новый
		if currentProcess == nil || currentProcess.State != StateRunning {
			currentProcess = s.GetReadyProcess()
		}

		// Печатаем текущее состояние
		s.PrintCurrentState(currentProcess)

		// Выполняем инструкцию
		if currentProcess != nil {
			currentProcess.State = StateRunning

			if currentProcess.PC < len(currentProcess.Instructions) {
				inst := currentProcess.Instructions[currentProcess.PC]

				switch inst.Type {
				case "cpu":
					// CPU инструкция выполняется за один такт
					currentProcess.PC++
					if currentProcess.PC >= len(currentProcess.Instructions) {
						currentProcess.State = StateDone
						currentProcess = nil // Переключаемся на следующий процесс
					} else if s.SwitchBehavior == "SWITCH_ON_END" {
						// Продолжаем выполнение того же процесса
						currentProcess.State = StateReady
					} else {
						currentProcess.State = StateReady
						currentProcess = nil // Переключаемся на следующий процесс
					}

				case "io":
					// I/O инструкция блокирует процесс
					currentProcess.State = StateBlocked
					currentProcess.IOTimeLeft = inst.Duration
					currentProcess = nil // Переключаемся на следующий процесс
				}
			}
		}

		s.CurrentTime++

		// Защита от бесконечного цикла
		if s.CurrentTime > 1000 {
			fmt.Println("Превышено максимальное время симуляции!")
			break
		}
	}

	fmt.Printf("\nСимуляция завершена за %d тактов\n", s.CurrentTime)
}

func main() {
	// Определяем флаги командной строки
	var (
		processList    = flag.String("l", "0:cccc,1:cc", "список процессов (формат: id:инструкции)")
		ioLength       = flag.Int("L", 5, "длительность I/O операции")
		switchBehavior = flag.String("S", "SWITCH_ON_IO", "поведение переключения (SWITCH_ON_IO/SWITCH_ON_END)")
		printState     = flag.Bool("p", false, "печатать состояние каждого такта")
		seed           = flag.Int64("s", 0, "seed для генератора случайных чисел")
		help           = flag.Bool("h", false, "показать помощь")
	)

	flag.Parse()

	if *help {
		fmt.Println("Симулятор выполнения процессов")
		fmt.Println("===============================")
		fmt.Println("Использование:")
		fmt.Println("  -l строка    список процессов (по умолчанию \"0:cccc,1:cc\")")
		fmt.Println("  -L int       длительность I/O операции (по умолчанию 5)")
		fmt.Println("  -S строка    поведение переключения (по умолчанию \"SWITCH_ON_IO\")")
		fmt.Println("  -p           печатать состояние каждого такта")
		fmt.Println("  -s int       seed для генератора случайных чисел")
		fmt.Println("  -h           показать эту помощь")
		fmt.Println()
		fmt.Println("Формат процессов: id:инструкции")
		fmt.Println("  c = CPU инструкция")
		fmt.Println("  i = I/O инструкция")
		fmt.Println()
		fmt.Println("Примеры:")
		fmt.Println("  go run process-run.go -l \"0:cccc,1:cc\" -p")
		fmt.Println("  go run process-run.go -l \"0:cici,1:cc\" -L 3 -p")
		return
	}

	// Инициализируем генератор случайных чисел
	if *seed != 0 {
		rand.Seed(*seed)
	} else {
		rand.Seed(time.Now().UnixNano())
	}

	// Создаем симулятор
	sim := NewSimulator()
	sim.IOLength = *ioLength
	sim.SwitchBehavior = *switchBehavior
	sim.PrintState = *printState

	// Парсим список процессов
	processes := strings.Split(*processList, ",")
	for _, processStr := range processes {
		processStr = strings.TrimSpace(processStr)
		if processStr != "" {
			err := sim.AddProcess(processStr)
			if err != nil {
				fmt.Printf("Ошибка при добавлении процесса: %v\n", err)
				return
			}
		}
	}

	if len(sim.Processes) == 0 {
		fmt.Println("Не добавлено ни одного процесса!")
		return
	}

	// Печатаем информацию о процессах
	fmt.Printf("Параметры симуляции:\n")
	fmt.Printf("  I/O длительность: %d\n", sim.IOLength)
	fmt.Printf("  Поведение переключения: %s\n", sim.SwitchBehavior)
	fmt.Printf("  Процессов: %d\n", len(sim.Processes))
	fmt.Println()

	for _, p := range sim.Processes {
		fmt.Printf("Процесс %d: ", p.ID)
		for _, inst := range p.Instructions {
			fmt.Printf("%s ", strings.ToUpper(inst.Type))
		}
		fmt.Println()
	}
	fmt.Println()

	// Запускаем симуляцию
	sim.Run()
}
