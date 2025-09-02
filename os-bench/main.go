package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
)

const (
	// Количество итераций для измерений
	SYSCALL_ITERATIONS = 1000000
	CONTEXT_ITERATIONS = 10000
)

func main() {
	fmt.Println("=== Измерение производительности операционной системы ===")
	fmt.Printf("GOOS: %s, GOARCH: %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Количество CPU: %d\n", runtime.NumCPU())
	fmt.Println()

	// Измерение стоимости системного вызова
	measureSystemCallCost()
	fmt.Println()

	// Измерение стоимости контекстного переключения
	measureContextSwitchCost()
}

// measureSystemCallCost измеряет стоимость простого системного вызова
func measureSystemCallCost() {
	fmt.Printf("=== Измерение стоимости системного вызова ===\n")
	fmt.Printf("Количество итераций: %d\n", SYSCALL_ITERATIONS)

	// Разогрев
	for i := 0; i < 1000; i++ {
		syscall.Syscall(syscall.SYS_GETPID, 0, 0, 0)
	}

	// Измерение времени выполнения getpid() системного вызова
	start := time.Now()
	for i := 0; i < SYSCALL_ITERATIONS; i++ {
		syscall.Syscall(syscall.SYS_GETPID, 0, 0, 0)
	}
	elapsed := time.Since(start)

	avgNs := elapsed.Nanoseconds() / SYSCALL_ITERATIONS
	fmt.Printf("Общее время: %v\n", elapsed)
	fmt.Printf("Среднее время на один системный вызов: %d нс\n", avgNs)
	fmt.Printf("Среднее время на один системный вызов: %.2f мкс\n", float64(avgNs)/1000.0)

	// Альтернативное измерение с использованием read() на /dev/null
	measureReadSyscall()
}

// measureReadSyscall измеряет стоимость системного вызова read()
func measureReadSyscall() {
	fmt.Println("\n--- Измерение read() системного вызова ---")

	file, err := os.OpenFile("/dev/null", os.O_RDONLY, 0)
	if err != nil {
		fmt.Printf("Ошибка открытия /dev/null: %v\n", err)
		return
	}
	defer file.Close()

	fd := int(file.Fd())
	buf := make([]byte, 0) // Читаем 0 байт

	// Разогрев
	for i := 0; i < 1000; i++ {
		syscall.Read(fd, buf)
	}

	start := time.Now()
	for i := 0; i < SYSCALL_ITERATIONS; i++ {
		syscall.Read(fd, buf)
	}
	elapsed := time.Since(start)

	avgNs := elapsed.Nanoseconds() / SYSCALL_ITERATIONS
	fmt.Printf("read() - Общее время: %v\n", elapsed)
	fmt.Printf("read() - Среднее время: %d нс (%.2f мкс)\n", avgNs, float64(avgNs)/1000.0)
}

// measureContextSwitchCost измеряет стоимость контекстного переключения
func measureContextSwitchCost() {
	fmt.Printf("=== Измерение стоимости контекстного переключения ===\n")
	fmt.Printf("Количество итераций: %d\n", CONTEXT_ITERATIONS)

	// Создаем pipe для коммуникации между процессами
	r1, w1, err := os.Pipe()
	if err != nil {
		fmt.Printf("Ошибка создания pipe1: %v\n", err)
		return
	}
	defer r1.Close()
	defer w1.Close()

	r2, w2, err := os.Pipe()
	if err != nil {
		fmt.Printf("Ошибка создания pipe2: %v\n", err)
		return
	}
	defer r2.Close()
	defer w2.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	// Канал для передачи времени измерения
	timeChan := make(chan time.Duration, 1)

	// Горутина 1 - отправитель
	go func() {
		defer wg.Done()
		defer w1.Close()
		defer r2.Close()

		data := []byte{1}
		response := make([]byte, 1)

		// Разогрев
		for i := 0; i < 100; i++ {
			w1.Write(data)
			r2.Read(response)
		}

		// Измерение
		start := time.Now()
		for i := 0; i < CONTEXT_ITERATIONS; i++ {
			w1.Write(data)
			r2.Read(response)
		}
		elapsed := time.Since(start)
		timeChan <- elapsed
	}()

	// Горутина 2 - получатель
	go func() {
		defer wg.Done()
		defer w2.Close()
		defer r1.Close()

		data := []byte{1}
		response := make([]byte, 1)

		// Разогрев
		for i := 0; i < 100; i++ {
			r1.Read(response)
			w2.Write(data)
		}

		// Измерение
		for i := 0; i < CONTEXT_ITERATIONS; i++ {
			r1.Read(response)
			w2.Write(data)
		}
	}()

	wg.Wait()
	elapsed := <-timeChan

	// Каждая итерация включает 2 контекстных переключения
	// (отправка -> получение -> отправка ответа -> получение ответа)
	avgNs := elapsed.Nanoseconds() / (CONTEXT_ITERATIONS * 2)
	fmt.Printf("Общее время: %v\n", elapsed)
	fmt.Printf("Среднее время контекстного переключения: %d нс\n", avgNs)
	fmt.Printf("Среднее время контекстного переключения: %.2f мкс\n", float64(avgNs)/1000.0)

	// Альтернативное измерение с использованием отдельных процессов
	measureProcessContextSwitch()
}

// measureProcessContextSwitch измеряет контекстное переключение между реальными процессами
func measureProcessContextSwitch() {
	fmt.Println("\n--- Измерение между процессами ---")

	// Создаем каналы для синхронизации
	ready := make(chan bool, 2)
	done := make(chan time.Duration, 1)

	// Запускаем два процесса, которые будут переключаться
	go processWorker(1, ready, done)
	go processWorker(2, ready, nil)

	// Ждем готовности обоих процессов
	<-ready
	<-ready

	// Получаем результат измерения
	elapsed := <-done

	avgNs := elapsed.Nanoseconds() / CONTEXT_ITERATIONS
	fmt.Printf("Процессы - Общее время: %v\n", elapsed)
	fmt.Printf("Процессы - Среднее время: %d нс (%.2f мкс)\n", avgNs, float64(avgNs)/1000.0)
}

// processWorker имитирует работу процесса для измерения переключений контекста
func processWorker(id int, ready chan bool, result chan time.Duration) {
	// Привязываем горутину к одному потоку ОС
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ready <- true

	if id == 1 {
		// Первый процесс измеряет время
		start := time.Now()
		for i := 0; i < CONTEXT_ITERATIONS; i++ {
			// Принудительно вызываем планировщик
			runtime.Gosched()
		}
		elapsed := time.Since(start)
		result <- elapsed
	} else {
		// Второй процесс просто участвует в переключениях
		for i := 0; i < CONTEXT_ITERATIONS; i++ {
			runtime.Gosched()
		}
	}
}

// Дополнительные функции для более точных измерений

// rdtsc возвращает значение Time Stamp Counter (только для x86/x64)
func rdtsc() uint64 {
	if runtime.GOARCH != "amd64" && runtime.GOARCH != "386" {
		return 0
	}

	// Простая реализация rdtsc через inline assembly недоступна в Go
	// Используем time.Now() как альтернативу
	return uint64(time.Now().UnixNano())
}

// getCPUInfo возвращает информацию о процессоре
func getCPUInfo() {
	fmt.Println("=== Информация о системе ===")

	if runtime.GOOS == "linux" {
		if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
			lines := string(data)
			fmt.Printf("CPU Info (первые 500 символов):\n%.500s...\n", lines)
		}
	}

	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
}
