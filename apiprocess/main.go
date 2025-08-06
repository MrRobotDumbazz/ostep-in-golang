package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

func createTempGoFile(code string) (string, func(), error) {
	tmpFile, err := os.CreateTemp("", "ostep_*.go")
	if err != nil {
		return "", nil, err
	}

	_, err = tmpFile.WriteString(code)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", nil, err
	}
	tmpFile.Close()

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	return tmpFile.Name(), cleanup, nil
}

func ForkVariable() {
	fmt.Println("=== Задание 1: fork() и переменные ===")

	x := 100
	fmt.Printf("Родительский процесс: x = %d (PID: %d)\n", x, os.Getpid())

	childCode := `package main 
import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	parentX, _ := strconv.Atoi(os.Getenv("PARENT_X"))
	fmt.Printf("Дочерний процесс: x = %d (PID: %d)\n", parentX, os.Getpid())

	childX := parentX + 50 
	fmt.Printf("Дочерний процесс изменил x на: %d\n", childX)
}`

	tmpFile, cleanup, err := createTempGoFile(childCode)
	if err != nil {
		fmt.Printf("Ошибка создания временного файла: %v\n", err)
		return
	}
	defer cleanup()

	cmd := exec.Command("go", "run", tmpFile)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PARENT_X=%d", x))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Ошибка запуска дочернего процесса: %v\n", err)
	}

	x = x + 25
	fmt.Printf("Родительский процесс изменил x на: %d\n", x)

	fmt.Println("Вывод: Каждый процесс имеет свою копию переменной")
}

func FileDescriptors() {
	fmt.Println("=== Задание 2: Файловые дескрипторы и fork() ===")

	tempFile, err := os.CreateTemp("", "ostep_hw2_*.txt")
	if err != nil {
		fmt.Printf("Ошибка создания файла: %v\n", err)
		return
	}
	defer os.Remove(tempFile.Name())

	fmt.Printf("Создан файл: %s\n", tempFile.Name())

	tempFile.WriteString("Родительский процесс записал эту строку\n")
	tempFile.Sync()

	childCode := fmt.Sprintf(`package main
import (
	"fmt"
	"os"
	"time"
)

func main() {
	file, err := os.OpenFile("%s", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Дочерний процесс не может открыть файл: %%v\n", err)
		return
	}
	defer file.Close()
	
	file.WriteString("Дочерний процесс записал эту строку\n")
	fmt.Printf("Дочерний процесс (PID: %%d) записал в файл\n", os.Getpid())

	for i := 0; i < 3; i++ {
		file.WriteString(fmt.Sprintf("Дочерний процесс: запись %%d\n", i))
		time.Sleep(100 * time.Millisecond)
	}
}`, tempFile.Name())

	tmpFile, cleanup, err := createTempGoFile(childCode)
	if err != nil {
		fmt.Printf("Ошибка создания временного файла: %v\n", err)
		return
	}
	defer cleanup()

	cmd := exec.Command("go", "run", tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()

	for i := 0; i < 3; i++ {
		tempFile.WriteString(fmt.Sprintf("Родительский процесс: запись %d\n", i))
		time.Sleep(100 * time.Millisecond)
	}

	cmd.Wait()
	tempFile.Close()

	content, _ := os.ReadFile(tempFile.Name())
	fmt.Println("Содержимое файла после одновременной записи:")
	fmt.Printf("%s\n", content)

	fmt.Println("Вывод: Оба процесса могут использовать файл, но нужна синхронизация")
}

func OrderWithoutWait() {
	fmt.Println("=== Задание 3: Порядок выполнения без wait() ===")

	childCode := `package main
import (
	"fmt"
	"os"
	"time"
)

func main() {
	time.Sleep(50 * time.Millisecond) // Небольшая задержка
	fmt.Printf("hello (child PID: %d)\n", os.Getpid())
}`

	tmpFile, cleanup, err := createTempGoFile(childCode)
	if err != nil {
		fmt.Printf("Ошибка создания временного файла: %v\n", err)
		return
	}
	defer cleanup()

	cmd := exec.Command("go", "run", tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()

	// Без wait() - сразу выводим goodbye
	fmt.Printf("goodbye (parent PID: %d)\n", os.Getpid())

	// Все же дождемся завершения чтобы не оставлять зомби-процессы
	go func() {
		cmd.Wait()
	}()

	time.Sleep(200 * time.Millisecond) // Задержка чтобы увидеть результат дочернего процесса

	fmt.Println("Вывод: Без wait() порядок выполнения непредсказуем")
}

func OrderWithWait() {
	fmt.Println("=== Задание 3b: Порядок выполнения с wait() ===")

	childCode := `package main
import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("hello (child PID: %d)\n", os.Getpid())
}`

	tmpFile, cleanup, err := createTempGoFile(childCode)
	if err != nil {
		fmt.Printf("Ошибка создания временного файла: %v\n", err)
		return
	}
	defer cleanup()

	cmd := exec.Command("go", "run", tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()
	cmd.Wait() // Ждем завершения дочернего процесса

	fmt.Printf("goodbye (parent PID: %d)\n", os.Getpid())

	fmt.Println("Вывод: С wait() порядок выполнения предсказуем")
}

func ExecVariants() {
	fmt.Println("=== Задание 4: Варианты exec() ===")

	fmt.Println("1. Простой exec:")
	cmd1 := exec.Command("ls", "-l")
	output1, err := cmd1.CombinedOutput()
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Printf("Результат ls -l:\n%s\n", output1)
	}

	fmt.Println("2. Exec с переменными окружения:")
	cmd2 := exec.Command("printenv", "CUSTOM_VAR")
	cmd2.Env = []string{"CUSTOM_VAR=test_value", "PATH=" + os.Getenv("PATH")}
	output2, err := cmd2.CombinedOutput()
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Printf("Переменная CUSTOM_VAR: %s\n", output2)
	}

	fmt.Println("3. Exec с рабочей директорией:")
	cmd3 := exec.Command("pwd")
	cmd3.Dir = "/tmp"
	output3, err := cmd3.CombinedOutput()
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Printf("Рабочая директория: %s\n", output3)
	}

	fmt.Println("Вывод: Разные варианты exec() предоставляют различные способы настройки окружения")
}

func WaitReturn() {
	fmt.Println("=== Задание 5: wait() и его возвращаемое значение ===")

	childCode := `package main
import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Printf("Дочерний процесс (PID: %d) работает...\n", os.Getpid())
	time.Sleep(1 * time.Second)
	fmt.Println("Дочерний процесс завершается с кодом 42")
	os.Exit(42)
}`

	tmpFile, cleanup, err := createTempGoFile(childCode)
	if err != nil {
		fmt.Printf("Ошибка создания временного файла: %v\n", err)
		return
	}
	defer cleanup()

	cmd := exec.Command("go", "run", tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Родительский процесс (PID: %d) запускает дочерний процесс\n", os.Getpid())
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Ошибка запуска: %v\n", err)
		return
	}

	fmt.Printf("Дочерний процесс запущен (PID: %d)\n", cmd.Process.Pid)
	fmt.Println("Родительский процесс ждет завершения дочернего...")

	err = cmd.Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Дочерний процесс завершился с кодом: %d\n", exitError.ExitCode())
		} else {
			fmt.Printf("Ошибка ожидания: %v\n", err)
		}
	} else {
		fmt.Println("Дочерний процесс завершился успешно")
	}

	fmt.Println("Примечание: В Go, exit код передается только если процесс завершился с ошибкой")
	fmt.Println("Вывод: wait() возвращает информацию о завершении дочернего процесса")
}

func WaitPid() {
	fmt.Println("=== Задание 6: waitpid() vs wait() ===")
	var cmds []*exec.Cmd
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		childCode := fmt.Sprintf(`package main
import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("Дочерний процесс %d запущен, PID:", os.Getpid())
	time.Sleep(%d * time.Second)
	fmt.Println("Дочерний процесс %d завершен")
}`, i, i+1, i)

		tmpFile, cleanup, err := createTempGoFile(childCode)
		if err != nil {
			fmt.Printf("Ошибка создания временного файла для процесса %d: %v\n", i, err)
			continue
		}
		defer cleanup()

		cmd := exec.Command("go", "run", tmpFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Start()
		cmds = append(cmds, cmd)
		fmt.Printf("Запущен дочерний процесс %d (PID: %d)\n", i, cmd.Process.Pid)
	}

	fmt.Println("Ожидание завершения всех дочерних процессов...")

	// Ждем каждый процесс в отдельной горутине
	for i, cmd := range cmds {
		wg.Add(1)
		go func(index int, c *exec.Cmd) {
			defer wg.Done()
			err := c.Wait()
			if err != nil {
				fmt.Printf("Процесс %d завершился с ошибкой: %v\n", index, err)
			} else {
				fmt.Printf("Процесс %d завершился успешно\n", index)
			}
		}(i, cmd)
	}

	wg.Wait()
	fmt.Println("Вывод: В Go можно ждать конкретные процессы используя горутины")
}

func CloseStdout() {
	fmt.Println("=== Задание 7: Закрытие stdout ===")

	childCode := `package main
import (
	"fmt"
	"os"
	"syscall"
)

func main() {
	fmt.Printf("Дочерний процесс (PID: %d) перед закрытием stdout\n", os.Getpid())

	// Закрываем stdout (файловый дескриптор 1)
	syscall.Close(1)

	// Эта строка не должна появиться на экране
	fmt.Printf("Эта строка не должна появиться на экране\n")

	// Используем stderr вместо stdout
	fmt.Fprintf(os.Stderr, "Дочерний процесс: stdout закрыт, используем stderr\n")
}`

	tmpFile, cleanup, err := createTempGoFile(childCode)
	if err != nil {
		fmt.Printf("Ошибка создания временного файла: %v\n", err)
		return
	}
	defer cleanup()

	cmd := exec.Command("go", "run", tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Ошибка выполнения: %v\n", err)
	}

	fmt.Println("Вывод: После закрытия stdout, printf() не выводит на экран")
}

func Pipe() {
	fmt.Println("=== Задание 8: pipe() - соединение процессов ===")

	producerCode := `package main
import (
	"fmt"
	"time"
)

func main() {
	for i := 1; i <= 5; i++ {
		fmt.Printf("Данные от процесса-генератора: %d\n", i)
		time.Sleep(500 * time.Millisecond)
	}
}`

	consumerCode := `package main
import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Fprintf(os.Stderr, "Процесс-потребитель читает данные:\n")
	scanner := bufio.NewScanner(os.Stdin)
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		count++
		fmt.Fprintf(os.Stderr, "Обработано: %s (строка %d)\n", strings.TrimSpace(line), count)
	}
	fmt.Fprintf(os.Stderr, "Процесс-потребитель обработал %d строк\n", count)
}`

	// Создаем временные файлы
	producerFile, cleanupProducer, err := createTempGoFile(producerCode)
	if err != nil {
		fmt.Printf("Ошибка создания файла генератора: %v\n", err)
		return
	}
	defer cleanupProducer()

	consumerFile, cleanupConsumer, err := createTempGoFile(consumerCode)
	if err != nil {
		fmt.Printf("Ошибка создания файла потребителя: %v\n", err)
		return
	}
	defer cleanupConsumer()

	// Создаем pipe
	r, w, err := os.Pipe()
	if err != nil {
		fmt.Printf("Ошибка создания pipe: %v\n", err)
		return
	}

	// Процесс-генератор
	producer := exec.Command("go", "run", producerFile)
	producer.Stdout = w
	producer.Stderr = os.Stderr

	// Процесс-потребитель
	consumer := exec.Command("go", "run", consumerFile)
	consumer.Stdin = r
	consumer.Stdout = os.Stdout
	consumer.Stderr = os.Stderr

	fmt.Println("Запуск процесса-генератора...")
	producer.Start()

	fmt.Println("Запуск процесса-потребителя...")
	consumer.Start()

	// Ждем завершения генератора и закрываем pipe
	producer.Wait()
	w.Close()

	// Ждем завершения потребителя
	consumer.Wait()
	r.Close()

	fmt.Println("Вывод: pipe() позволяет соединить stdout одного процесса с stdin другого")
}

// Демонстрация создания множества процессов и zombie процессов
func ZombieDemo() {
	fmt.Println("=== Дополнительно: Демонстрация zombie процессов ===")

	childCode := `package main
import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("Дочерний процесс (PID: %d) завершается немедленно\n", os.Getpid())
}`

	tmpFile, cleanup, err := createTempGoFile(childCode)
	if err != nil {
		fmt.Printf("Ошибка создания временного файла: %v\n", err)
		return
	}
	defer cleanup()

	cmd := exec.Command("go", "run", tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Родительский процесс (PID: %d) запускает дочерний\n", os.Getpid())
	cmd.Start()

	fmt.Printf("Дочерний процесс запущен (PID: %d)\n", cmd.Process.Pid)
	fmt.Println("Родительский спит 2 секунды, не вызывая wait()...")

	time.Sleep(2 * time.Second)

	fmt.Println("Теперь вызываем wait()...")
	err = cmd.Wait()
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
	} else {
		fmt.Println("Дочерний процесс успешно завершен")
	}

	fmt.Println("Вывод: Без wait() дочерний процесс становится zombie")
}

// Дополнительная демонстрация гонки условий
func RaceConditionDemo() {
	fmt.Println("=== Дополнительно: Демонстрация гонки условий ===")

	// Создаем общий файл
	sharedFile, err := os.CreateTemp("", "race_*.txt")
	if err != nil {
		fmt.Printf("Ошибка создания файла: %v\n", err)
		return
	}
	defer os.Remove(sharedFile.Name())
	sharedFile.Close()

	var wg sync.WaitGroup
	numProcesses := 5

	for i := 0; i < numProcesses; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			childCode := fmt.Sprintf(`package main
			import (
				"fmt"
				"os"
				"time"
			)
			
			func main() {
				file, err := os.OpenFile("%s", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
				if err != nil {
					fmt.Println("Процесс %d: ошибка открытия файла:", err)
					return
				}
				defer file.Close()
				
				for j := 0; j < 3; j++ {
					file.WriteString("Процесс %d: запись " + fmt.Sprint(j) + "\n")
					time.Sleep(10 * time.Millisecond)
				}
				fmt.Println("Процесс %d завершен")
			}`, sharedFile.Name(), id, id, id)

			tmpFile, cleanup, err := createTempGoFile(childCode)
			if err != nil {
				fmt.Printf("Ошибка создания временного файла для процесса %d: %v\n", id, err)
				return
			}
			defer cleanup()

			cmd := exec.Command("go", "run", tmpFile)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}(i)
	}

	fmt.Printf("Запущено %d параллельных процессов, записывающих в общий файл...\n", numProcesses)
	wg.Wait()

	// Читаем результат
	content, _ := os.ReadFile(sharedFile.Name())
	fmt.Println("Содержимое файла после параллельной записи:")
	fmt.Printf("%s\n", content)

	fmt.Println("Вывод: Без синхронизации возможны гонки условий при записи в общий ресурс")
}

func main() {
	fmt.Println("OSTEP Глава 5: Process API - Домашнее задание (Исправленная версия)")
	fmt.Println("==================================================================")

	ForkVariable()
	fmt.Println()

	FileDescriptors()
	fmt.Println()

	OrderWithoutWait()
	fmt.Println()

	OrderWithWait()
	fmt.Println()

	ExecVariants()
	fmt.Println()

	WaitReturn()
	fmt.Println()

	WaitPid()
	fmt.Println()

	CloseStdout()
	fmt.Println()

	Pipe()
	fmt.Println()

	ZombieDemo()
	fmt.Println()

	RaceConditionDemo()

	fmt.Println("==================================================================")
	fmt.Println("Все задания выполнены!")
	fmt.Println()
	fmt.Println("Основные выводы:")
	fmt.Println("1. В Go нет прямого fork(), но exec.Command() создает новые процессы")
	fmt.Println("2. Файловые дескрипторы можно передавать через Stdin/Stdout/Stderr")
	fmt.Println("3. Wait() позволяет синхронизировать процессы и избегать zombie")
	fmt.Println("4. exec.Command() аналогичен exec() - запускает новую программу")
	fmt.Println("5. os.Pipe() позволяет соединять процессы в цепочки")
	fmt.Println("6. Горутины предоставляют более удобный способ параллелизма в Go")
	fmt.Println("7. Без синхронизации возможны гонки условий при работе с общими ресурсами")
}
