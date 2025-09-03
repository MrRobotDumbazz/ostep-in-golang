package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Task представляет задачу в системе
type Task struct {
	ID         int // Идентификатор задачи
	Duration   int // Продолжительность выполнения
	Arrival    int // Время прибытия
	Start      int // Время начала выполнения
	Finish     int // Время завершения
	Response   int // Время отклика (Start - Arrival)
	Turnaround int // Оборотное время (Finish - Arrival)
	Waiting    int // Время ожидания (Turnaround - Duration)
}

// SchedulerResult содержит результаты планирования
type SchedulerResult struct {
	Tasks         []Task
	AvgResponse   float64
	AvgTurnaround float64
	AvgWaiting    float64
	TotalTime     int
	SchedulerType string
	TimeQuantum   int // Для RR
}

func main() {
	fmt.Println("=== Эмулятор планировщика процессов ===\n")

	// Задачи для тестирования (продолжительность 200с)
	tasks200 := []Task{
		{ID: 1, Duration: 80, Arrival: 0},
		{ID: 2, Duration: 60, Arrival: 10},
		{ID: 3, Duration: 60, Arrival: 20},
	}

	// Более показательный пример для демонстрации разницы SJF vs FIFO
	tasksDemonstration := []Task{
		{ID: 1, Duration: 100, Arrival: 0}, // Длинная задача приходит первой
		{ID: 2, Duration: 10, Arrival: 5},  // Короткая задача
		{ID: 3, Duration: 20, Arrival: 10}, // Средняя задача
		{ID: 4, Duration: 5, Arrival: 15},  // Очень короткая задача
	}

	// Различные продолжительности
	tasks100 := []Task{
		{ID: 1, Duration: 50, Arrival: 0},
		{ID: 2, Duration: 30, Arrival: 5},
		{ID: 3, Duration: 20, Arrival: 10},
	}

	tasks300 := []Task{
		{ID: 1, Duration: 120, Arrival: 0},
		{ID: 2, Duration: 90, Arrival: 15},
		{ID: 3, Duration: 90, Arrival: 30},
	}

	// RR задачи с временным квантом 1
	tasksRR := []Task{
		{ID: 1, Duration: 80, Arrival: 0},
		{ID: 2, Duration: 60, Arrival: 10},
		{ID: 3, Duration: 60, Arrival: 20},
	}

	fmt.Println("1. Анализ SJF и FIFO для задач продолжительностью 200с:")
	analyzeSJFvsFIFO(tasks200)

	fmt.Println("\n1a. Демонстрация разницы SJF vs FIFO (эффект конвоя):")
	analyzeSJFvsFIFO(tasksDemonstration)

	fmt.Println("\n2. Анализ различных продолжительностей:")
	fmt.Println("\n--- 100с общая продолжительность ---")
	analyzeSJFvsFIFO(tasks100)
	fmt.Println("\n--- 200с общая продолжительность ---")
	analyzeSJFvsFIFO(tasks200)
	fmt.Println("\n--- 300с общая продолжительность ---")
	analyzeSJFvsFIFO(tasks300)

	fmt.Println("\n3. Анализ RR с временным квантом 1:")
	resultRR := scheduleRR(tasksRR, 1)
	printResult(resultRR)

	fmt.Println("\n4. Сравнение всех алгоритмов:")
	compareAllSchedulers(tasks200)

	fmt.Println("\n5. Анализ рабочих нагрузок:")
	analyzeWorkloads()

	fmt.Println("\n6. Анализ RR с увеличением временного кванта:")
	analyzeRRTimeQuantum()

	fmt.Println("\n7. Формула времени отклика для RR:")
	deriveRRResponseTimeFormula()
}

// scheduleFIFO реализует планирование FIFO (First In, First Out)
func scheduleFIFO(tasks []Task) SchedulerResult {
	result := tasks // копируем задачи

	// Сортируем по времени прибытия
	sort.Slice(result, func(i, j int) bool {
		return result[i].Arrival < result[j].Arrival
	})

	currentTime := 0

	for i := range result {
		// Если процессор свободен и задача еще не прибыла, ждем
		if currentTime < result[i].Arrival {
			currentTime = result[i].Arrival
		}

		result[i].Start = currentTime
		result[i].Finish = currentTime + result[i].Duration
		result[i].Response = result[i].Start - result[i].Arrival
		result[i].Turnaround = result[i].Finish - result[i].Arrival
		result[i].Waiting = result[i].Turnaround - result[i].Duration

		currentTime = result[i].Finish
	}

	return SchedulerResult{
		Tasks:         result,
		SchedulerType: "FIFO",
		TotalTime:     currentTime,
		AvgResponse:   calculateAvgResponse(result),
		AvgTurnaround: calculateAvgTurnaround(result),
		AvgWaiting:    calculateAvgWaiting(result),
	}
}

// scheduleSJF реализует планирование SJF (Shortest Job First)
func scheduleSJF(tasks []Task) SchedulerResult {
	result := make([]Task, len(tasks))
	copy(result, tasks)

	var completed []Task
	currentTime := 0

	for len(result) > 0 {
		// Найти все задачи, которые уже прибыли
		var available []int
		for i, task := range result {
			if task.Arrival <= currentTime {
				available = append(available, i)
			}
		}

		if len(available) == 0 {
			// Нет доступных задач, переходим к следующему времени прибытия
			minArrival := math.MaxInt32
			for _, task := range result {
				if task.Arrival < minArrival {
					minArrival = task.Arrival
				}
			}
			currentTime = minArrival
			continue
		}

		// Выбрать самую короткую задачу среди доступных
		shortestIdx := available[0]
		for _, idx := range available {
			if result[idx].Duration < result[shortestIdx].Duration {
				shortestIdx = idx
			}
		}

		// Выполнить выбранную задачу
		task := result[shortestIdx]
		task.Start = currentTime
		task.Finish = currentTime + task.Duration
		task.Response = task.Start - task.Arrival
		task.Turnaround = task.Finish - task.Arrival
		task.Waiting = task.Turnaround - task.Duration

		completed = append(completed, task)
		currentTime = task.Finish

		// Удалить выполненную задачу из списка
		result = append(result[:shortestIdx], result[shortestIdx+1:]...)
	}

	return SchedulerResult{
		Tasks:         completed,
		SchedulerType: "SJF",
		TotalTime:     currentTime,
		AvgResponse:   calculateAvgResponse(completed),
		AvgTurnaround: calculateAvgTurnaround(completed),
		AvgWaiting:    calculateAvgWaiting(completed),
	}
}

// scheduleRR реализует планирование Round Robin
func scheduleRR(tasks []Task, timeQuantum int) SchedulerResult {
	type TaskState struct {
		Task
		RemainingTime int
		FirstRun      bool
	}

	taskStates := make([]TaskState, len(tasks))
	for i, task := range tasks {
		taskStates[i] = TaskState{
			Task:          task,
			RemainingTime: task.Duration,
			FirstRun:      true,
		}
	}

	var completed []Task
	var readyQueue []int // индексы задач в очереди
	currentTime := 0

	for len(completed) < len(tasks) {
		// Добавить новые задачи в очередь
		for i := range taskStates {
			if taskStates[i].Arrival <= currentTime && taskStates[i].RemainingTime > 0 {
				// Проверить, не находится ли уже в очереди
				inQueue := false
				for _, queueIdx := range readyQueue {
					if queueIdx == i {
						inQueue = true
						break
					}
				}
				if !inQueue {
					readyQueue = append(readyQueue, i)
				}
			}
		}

		if len(readyQueue) == 0 {
			// Найти следующее время прибытия
			minArrival := math.MaxInt32
			for i := range taskStates {
				if taskStates[i].RemainingTime > 0 && taskStates[i].Arrival > currentTime {
					if taskStates[i].Arrival < minArrival {
						minArrival = taskStates[i].Arrival
					}
				}
			}
			currentTime = minArrival
			continue
		}

		// Взять первую задачу из очереди
		currentTaskIdx := readyQueue[0]
		readyQueue = readyQueue[1:]

		task := &taskStates[currentTaskIdx]

		// Записать время начала при первом запуске
		if task.FirstRun {
			task.Start = currentTime
			task.FirstRun = false
		}

		// Выполнить задачу в течение кванта времени
		executionTime := timeQuantum
		if task.RemainingTime < timeQuantum {
			executionTime = task.RemainingTime
		}

		currentTime += executionTime
		task.RemainingTime -= executionTime

		// Если задача завершена
		if task.RemainingTime == 0 {
			task.Finish = currentTime
			task.Response = task.Start - task.Arrival
			task.Turnaround = task.Finish - task.Arrival
			task.Waiting = task.Turnaround - task.Duration
			completed = append(completed, task.Task)
		} else {
			// Добавить задачу обратно в конец очереди, если она не завершена
			// Но сначала добавим новые задачи, которые могли прийти
			for i := range taskStates {
				if taskStates[i].Arrival <= currentTime && taskStates[i].RemainingTime > 0 && i != currentTaskIdx {
					inQueue := false
					for _, queueIdx := range readyQueue {
						if queueIdx == i {
							inQueue = true
							break
						}
					}
					if !inQueue {
						readyQueue = append(readyQueue, i)
					}
				}
			}
			readyQueue = append(readyQueue, currentTaskIdx)
		}
	}

	return SchedulerResult{
		Tasks:         completed,
		SchedulerType: "RR",
		TimeQuantum:   timeQuantum,
		TotalTime:     currentTime,
		AvgResponse:   calculateAvgResponse(completed),
		AvgTurnaround: calculateAvgTurnaround(completed),
		AvgWaiting:    calculateAvgWaiting(completed),
	}
}

// Функции для расчета средних значений
func calculateAvgResponse(tasks []Task) float64 {
	total := 0
	for _, task := range tasks {
		total += task.Response
	}
	return float64(total) / float64(len(tasks))
}

func calculateAvgTurnaround(tasks []Task) float64 {
	total := 0
	for _, task := range tasks {
		total += task.Turnaround
	}
	return float64(total) / float64(len(tasks))
}

func calculateAvgWaiting(tasks []Task) float64 {
	total := 0
	for _, task := range tasks {
		total += task.Waiting
	}
	return float64(total) / float64(len(tasks))
}

// printResult выводит результаты планирования
func printResult(result SchedulerResult) {
	fmt.Printf("=== Результаты планирования %s", result.SchedulerType)
	if result.TimeQuantum > 0 {
		fmt.Printf(" (квант: %d)", result.TimeQuantum)
	}
	fmt.Println(" ===")

	fmt.Printf("%-5s %-10s %-8s %-8s %-8s %-10s %-12s %-10s\n",
		"ID", "Прибытие", "Длительн", "Начало", "Конец", "Отклик", "Оборотное", "Ожидание")
	fmt.Println(strings.Repeat("-", 75))

	for _, task := range result.Tasks {
		fmt.Printf("%-5d %-10d %-8d %-8d %-8d %-10d %-12d %-10d\n",
			task.ID, task.Arrival, task.Duration, task.Start, task.Finish,
			task.Response, task.Turnaround, task.Waiting)
	}

	fmt.Printf("\nСредние значения:\n")
	fmt.Printf("  Время отклика: %.2f\n", result.AvgResponse)
	fmt.Printf("  Оборотное время: %.2f\n", result.AvgTurnaround)
	fmt.Printf("  Время ожидания: %.2f\n", result.AvgWaiting)
	fmt.Printf("  Общее время: %d\n", result.TotalTime)
}

// analyzeSJFvsFIFO сравнивает SJF и FIFO
func analyzeSJFvsFIFO(tasks []Task) {
	resultSJF := scheduleSJF(tasks)
	resultFIFO := scheduleFIFO(tasks)

	printResult(resultSJF)
	fmt.Println()
	printResult(resultFIFO)

	fmt.Printf("\n--- Сравнение SJF vs FIFO ---\n")
	fmt.Printf("Время отклика - SJF: %.2f, FIFO: %.2f (разница: %.2f)\n",
		resultSJF.AvgResponse, resultFIFO.AvgResponse,
		resultFIFO.AvgResponse-resultSJF.AvgResponse)
	fmt.Printf("Оборотное время - SJF: %.2f, FIFO: %.2f (разница: %.2f)\n",
		resultSJF.AvgTurnaround, resultFIFO.AvgTurnaround,
		resultFIFO.AvgTurnaround-resultSJF.AvgTurnaround)
}

// compareAllSchedulers сравнивает все алгоритмы
func compareAllSchedulers(tasks []Task) {
	resultSJF := scheduleSJF(tasks)
	resultFIFO := scheduleFIFO(tasks)
	resultRR1 := scheduleRR(tasks, 1)
	resultRR10 := scheduleRR(tasks, 10)

	fmt.Printf("%-12s %-15s %-15s %-15s\n", "Алгоритм", "Время отклика", "Оборотное время", "Время ожидания")
	fmt.Println(strings.Repeat("-", 65))
	fmt.Printf("%-12s %-15.2f %-15.2f %-15.2f\n", "SJF", resultSJF.AvgResponse, resultSJF.AvgTurnaround, resultSJF.AvgWaiting)
	fmt.Printf("%-12s %-15.2f %-15.2f %-15.2f\n", "FIFO", resultFIFO.AvgResponse, resultFIFO.AvgTurnaround, resultFIFO.AvgWaiting)
	fmt.Printf("%-12s %-15.2f %-15.2f %-15.2f\n", "RR(q=1)", resultRR1.AvgResponse, resultRR1.AvgTurnaround, resultRR1.AvgWaiting)
	fmt.Printf("%-12s %-15.2f %-15.2f %-15.2f\n", "RR(q=10)", resultRR10.AvgResponse, resultRR10.AvgTurnaround, resultRR10.AvgWaiting)
}

// analyzeWorkloads анализирует разные типы рабочих нагрузок
func analyzeWorkloads() {
	// Короткие задачи
	shortTasks := []Task{
		{ID: 1, Duration: 1, Arrival: 0},
		{ID: 2, Duration: 1, Arrival: 1},
		{ID: 3, Duration: 1, Arrival: 2},
	}

	// Длинные задачи
	longTasks := []Task{
		{ID: 1, Duration: 100, Arrival: 0},
		{ID: 2, Duration: 100, Arrival: 10},
		{ID: 3, Duration: 100, Arrival: 20},
	}

	fmt.Println("Анализ коротких задач (длительность 1с):")
	analyzeSJFvsFIFO(shortTasks)

	fmt.Println("\nАнализ длинных задач (длительность 100с):")
	analyzeSJFvsFIFO(longTasks)
}

// analyzeRRTimeQuantum анализирует влияние размера временного кванта на RR
func analyzeRRTimeQuantum() {
	tasks := []Task{
		{ID: 1, Duration: 80, Arrival: 0},
		{ID: 2, Duration: 60, Arrival: 10},
		{ID: 3, Duration: 60, Arrival: 20},
	}

	quantums := []int{1, 5, 10, 20, 50}

	fmt.Printf("%-10s %-15s %-15s %-15s\n", "Квант", "Время отклика", "Оборотное время", "Время ожидания")
	fmt.Println(strings.Repeat("-", 60))

	for _, quantum := range quantums {
		result := scheduleRR(tasks, quantum)
		fmt.Printf("%-10d %-15.2f %-15.2f %-15.2f\n",
			quantum, result.AvgResponse, result.AvgTurnaround, result.AvgWaiting)
	}
}

// deriveRRResponseTimeFormula выводит формулу времени отклика для RR
func deriveRRResponseTimeFormula() {
	fmt.Println("Формула времени отклика в худшем случае для RR:")
	fmt.Println("Если есть N задач, каждая требует времени T, и квант времени равен Q:")
	fmt.Println("")
	fmt.Println("Время отклика = (N-1) * Q")
	fmt.Println("")
	fmt.Println("Объяснение:")
	fmt.Println("- Задача может ждать максимум (N-1) полных квантов времени")
	fmt.Println("- до того, как получит свой первый квант времени")
	fmt.Println("- Это происходит, когда задача прибывает последней")
	fmt.Println("- и все остальные задачи уже находятся в очереди")
	fmt.Println("")
	fmt.Println("Пример с 3 задачами и квантом 1:")
	fmt.Println("Максимальное время отклика = (3-1) * 1 = 2 единицы времени")
}
