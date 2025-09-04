package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Job представляет задачу в системе
type Job struct {
	ID            uint
	ArrivalTime   uint
	JobLength     uint
	IOFrequency   uint // 0 означает отсутствие I/O
	CurrentQueue  uint
	TimeLeft      uint
	StartTime     int // -1 если еще не запущена
	EndTime       uint
	TotalWait     uint
	LastRun       uint
	TimeSliceLeft uint
	IOEndTime     uint // Время завершения I/O операции
}

// MLFQ представляет многоуровневый планировщик
type MLFQ struct {
	Queues      [][]uint // ID задач в каждой очереди
	NumQueues   uint
	TimeSlice   []uint // Временной квант для каждой очереди
	BoostTime   uint   // Время для повышения приоритета
	Jobs        map[uint]*Job
	CurrentTime uint
	LastBoost   uint
	IOQueue     []uint // Задачи, ожидающие завершения I/O
	PendingJobs []*Job // Задачи, которые еще не прибыли
	IODuration  uint   // Длительность I/O операции
}

// NewMLFQ создает новый MLFQ планировщик
func NewMLFQ(numQueues uint, timeSlices []uint, boostTime uint, ioDuration uint) *MLFQ {
	queues := make([][]uint, numQueues)
	for i := range queues {
		queues[i] = make([]uint, 0)
	}

	return &MLFQ{
		Queues:      queues,
		NumQueues:   numQueues,
		TimeSlice:   timeSlices,
		BoostTime:   boostTime,
		Jobs:        make(map[uint]*Job),
		CurrentTime: 0,
		LastBoost:   0,
		IOQueue:     make([]uint, 0),
		PendingJobs: make([]*Job, 0),
		IODuration:  ioDuration,
	}
}

// AddPendingJob добавляет задачу, которая прибудет позже
func (m *MLFQ) AddPendingJob(job *Job) {
	job.StartTime = -1
	job.TimeLeft = job.JobLength
	m.PendingJobs = append(m.PendingJobs, job)
}

// AddJob добавляет новую задачу в планировщик (когда она прибывает)
func (m *MLFQ) AddJob(job *Job) {
	m.Jobs[job.ID] = job
	m.Queues[0] = append(m.Queues[0], job.ID) // Новые задачи идут в очередь с наивысшим приоритетом
	job.CurrentQueue = 0
	job.TimeLeft = job.JobLength
	job.TimeSliceLeft = m.TimeSlice[0]
}

// CheckArrivals проверяет прибывающие задачи
func (m *MLFQ) CheckArrivals() {
	newPending := make([]*Job, 0)
	for _, job := range m.PendingJobs {
		if job.ArrivalTime == m.CurrentTime {
			m.AddJob(job)
		} else if job.ArrivalTime > m.CurrentTime {
			newPending = append(newPending, job)
		}
	}
	m.PendingJobs = newPending
}

// MoveJobToLowerQueue перемещает задачу в очередь с более низким приоритетом
func (m *MLFQ) MoveJobToLowerQueue(jobID uint) {
	job := m.Jobs[jobID]
	currentQueue := job.CurrentQueue

	// Удаляем из текущей очереди
	for i, id := range m.Queues[currentQueue] {
		if id == jobID {
			m.Queues[currentQueue] = append(m.Queues[currentQueue][:i], m.Queues[currentQueue][i+1:]...)
			break
		}
	}

	// Добавляем в очередь с более низким приоритетом (если она существует)
	if currentQueue < m.NumQueues-1 {
		job.CurrentQueue = currentQueue + 1
		m.Queues[currentQueue+1] = append(m.Queues[currentQueue+1], jobID)
		job.TimeSliceLeft = m.TimeSlice[currentQueue+1]
	} else {
		// Остается в самой низкой очереди
		m.Queues[currentQueue] = append(m.Queues[currentQueue], jobID)
		job.TimeSliceLeft = m.TimeSlice[currentQueue]
	}
}

// BoostAllJobs повышает приоритет всех задач до наивысшей очереди
func (m *MLFQ) BoostAllJobs() {
	for i := uint(1); i < m.NumQueues; i++ {
		for _, jobID := range m.Queues[i] {
			job := m.Jobs[jobID]
			job.CurrentQueue = 0
			job.TimeSliceLeft = m.TimeSlice[0]
			m.Queues[0] = append(m.Queues[0], jobID)
		}
		m.Queues[i] = m.Queues[i][:0] // Очищаем очередь
	}
	m.LastBoost = m.CurrentTime
}

// GetNextJob возвращает следующую задачу для выполнения
func (m *MLFQ) GetNextJob() *Job {
	for i := uint(0); i < m.NumQueues; i++ {
		if len(m.Queues[i]) > 0 {
			jobID := m.Queues[i][0]
			return m.Jobs[jobID]
		}
	}
	return nil
}

// HandleIO обрабатывает завершение I/O операций
func (m *MLFQ) HandleIO() {
	newIOQueue := make([]uint, 0)
	for _, jobID := range m.IOQueue {
		job := m.Jobs[jobID]
		if m.CurrentTime >= job.IOEndTime {
			// Возвращаем задачу в очередь с наивысшим приоритетом
			job.CurrentQueue = 0
			job.TimeSliceLeft = m.TimeSlice[0]
			m.Queues[0] = append(m.Queues[0], jobID)
		} else {
			newIOQueue = append(newIOQueue, jobID)
		}
	}
	m.IOQueue = newIOQueue
}

// Run запускает симуляцию планировщика
func (m *MLFQ) Run(maxTime uint) {
	fmt.Printf("Время: %4s | Выполняется: %4s | Очереди: %s\n", "T", "Job", "Q0:Q1:Q2")
	fmt.Println("------------------------------------------------------------")

	for m.CurrentTime < maxTime {
		// Проверяем прибывающие задачи
		m.CheckArrivals()

		// Проверяем повышение приоритета
		if m.BoostTime > 0 && m.CurrentTime >= m.LastBoost+m.BoostTime {
			m.BoostAllJobs()
		}

		// Обрабатываем завершение I/O операций
		m.HandleIO()

		// Получаем следующую задачу
		currentJob := m.GetNextJob()

		if currentJob == nil {
			fmt.Printf("%4d | %8s | %s\n", m.CurrentTime, "IDLE", m.getQueueStatus())
			m.CurrentTime++
			continue
		}

		// Если задача только начинается
		if currentJob.StartTime == -1 {
			currentJob.StartTime = int(m.CurrentTime)
		}

		fmt.Printf("%4d | %8d | %s\n", m.CurrentTime, currentJob.ID, m.getQueueStatus())

		// Выполняем задачу
		currentJob.TimeLeft--
		currentJob.TimeSliceLeft--
		currentJob.LastRun = m.CurrentTime
		m.CurrentTime++

		// Проверяем завершение задачи
		if currentJob.TimeLeft == 0 {
			currentJob.EndTime = m.CurrentTime - 1
			fmt.Printf("Задача %d завершена в время %d\n", currentJob.ID, currentJob.EndTime)
			// Удаляем из очереди
			queue := currentJob.CurrentQueue
			for i, id := range m.Queues[queue] {
				if id == currentJob.ID {
					m.Queues[queue] = append(m.Queues[queue][:i], m.Queues[queue][i+1:]...)
					break
				}
			}
			continue
		}

		// Проверяем I/O операцию
		if currentJob.IOFrequency > 0 {
			executed := currentJob.JobLength - currentJob.TimeLeft
			if executed > 0 && executed%currentJob.IOFrequency == 0 {
				// Задача выполняет I/O
				queue := currentJob.CurrentQueue
				for i, id := range m.Queues[queue] {
					if id == currentJob.ID {
						m.Queues[queue] = append(m.Queues[queue][:i], m.Queues[queue][i+1:]...)
						break
					}
				}
				currentJob.IOEndTime = m.CurrentTime + m.IODuration - 1
				m.IOQueue = append(m.IOQueue, currentJob.ID)
				continue
			}
		}

		// Проверяем истечение временного кванта
		if currentJob.TimeSliceLeft == 0 {
			m.MoveJobToLowerQueue(currentJob.ID)
		}
	}

	m.PrintStatistics()
}

// getQueueStatus возвращает статус всех очередей
func (m *MLFQ) getQueueStatus() string {
	status := ""
	for i, queue := range m.Queues {
		if i > 0 {
			status += ":"
		}
		if len(queue) == 0 {
			status += "[]"
		} else {
			status += fmt.Sprintf("%v", queue)
		}
	}
	return status
}

// PrintStatistics выводит статистику выполнения
func (m *MLFQ) PrintStatistics() {
	fmt.Println("\n=== Статистика ===")

	var completedJobs []*Job
	var totalTurnaround uint = 0
	var totalResponse int = 0

	for _, job := range m.Jobs {
		if job.EndTime > 0 && job.StartTime >= 0 {
			completedJobs = append(completedJobs, job)
			turnaround := job.EndTime - job.ArrivalTime + 1
			response := job.StartTime - int(job.ArrivalTime)
			totalTurnaround += turnaround
			totalResponse += response

			fmt.Printf("Задача %d: Время отклика = %d, Время выполнения = %d\n",
				job.ID, response, turnaround)
		}
	}

	if len(completedJobs) > 0 {
		avgTurnaround := float64(totalTurnaround) / float64(len(completedJobs))
		avgResponse := float64(totalResponse) / float64(len(completedJobs))

		fmt.Printf("\nСреднее время выполнения: %.2f\n", avgTurnaround)
		fmt.Printf("Среднее время отклика: %.2f\n", avgResponse)
	}
}

// parseWorkload парсит пользовательскую рабочую нагрузку
func parseWorkload(workload string) ([]*Job, error) {
	var jobs []*Job
	jobStrings := strings.Split(workload, ";")

	for i, jobStr := range jobStrings {
		if jobStr == "" {
			continue
		}
		parts := strings.Split(jobStr, ",")
		if len(parts) < 2 {
			return nil, fmt.Errorf("неверный формат задачи: %s", jobStr)
		}

		arrival, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("неверное время прибытия: %s", parts[0])
		}

		length, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("неверная длительность: %s", parts[1])
		}

		var ioFreq uint = 0
		if len(parts) > 2 {
			freq, err := strconv.ParseUint(parts[2], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("неверная частота I/O: %s", parts[2])
			}
			ioFreq = uint(freq)
		}

		job := &Job{
			ID:          uint(i + 1),
			ArrivalTime: uint(arrival),
			JobLength:   uint(length),
			IOFrequency: ioFreq,
			StartTime:   -1,
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func main() {
	// Параметры командной строки
	numJobs := flag.Uint("j", 3, "Количество задач")
	maxTime := flag.Uint("t", 200, "Максимальное время симуляции")
	ioFreq := flag.Uint("i", 0, "Частота I/O операций (0 = нет I/O)")
	ioDuration := flag.Uint("I", 5, "Длительность I/O операции")
	boost := flag.Uint("B", 0, "Период повышения приоритета (0 = отключено)")
	seed := flag.Int64("s", time.Now().UnixNano(), "Семя для генератора случайных чисел")
	workload := flag.String("w", "", "Рабочая нагрузка (формат: время_прибытия,длительность,io_частота;...)")
	quantum0 := flag.Uint("q0", 10, "Временной квант для очереди 0")
	quantum1 := flag.Uint("q1", 20, "Временной квант для очереди 1")
	quantum2 := flag.Uint("q2", 40, "Временной квант для очереди 2")
	numQueues := flag.Uint("Q", 3, "Количество очередей")
	arrivalTime := flag.Uint("a", 20, "Максимальное время прибытия для случайных задач")
	jobLength := flag.Uint("l", 50, "Максимальная длительность для случайных задач")

	flag.Parse()

	rand.Seed(*seed)

	// Создаем временные кванты на основе флагов
	var timeSlices []uint
	if *numQueues >= 1 {
		timeSlices = append(timeSlices, *quantum0)
	}
	if *numQueues >= 2 {
		timeSlices = append(timeSlices, *quantum1)
	}
	if *numQueues >= 3 {
		timeSlices = append(timeSlices, *quantum2)
	}
	// Для дополнительных очередей используем последний квант
	for uint(len(timeSlices)) < *numQueues {
		timeSlices = append(timeSlices, timeSlices[len(timeSlices)-1])
	}

	// Создаем MLFQ планировщик
	scheduler := NewMLFQ(*numQueues, timeSlices, *boost, *ioDuration)

	// Добавляем задачи
	if *workload != "" {
		// Парсим пользовательскую рабочую нагрузку
		jobs, err := parseWorkload(*workload)
		if err != nil {
			fmt.Printf("Ошибка парсинга рабочей нагрузки: %v\n", err)
			return
		}

		for _, job := range jobs {
			scheduler.AddPendingJob(job)
			fmt.Printf("Добавлена задача %d: прибытие=%d, длительность=%d, I/O=%d\n",
				job.ID, job.ArrivalTime, job.JobLength, job.IOFrequency)
		}
	} else {
		// Генерируем случайные задачи
		for i := uint(0); i < *numJobs; i++ {
			job := &Job{
				ID:          i + 1,
				ArrivalTime: uint(rand.Intn(int(*arrivalTime))),
				JobLength:   uint(rand.Intn(int(*jobLength-10)) + 10), // От 10 до jobLength
				IOFrequency: *ioFreq,
				StartTime:   -1,
			}

			scheduler.AddPendingJob(job)
			fmt.Printf("Добавлена задача %d: прибытие=%d, длительность=%d\n",
				job.ID, job.ArrivalTime, job.JobLength)
		}
	}

	fmt.Printf("\nЗапуск MLFQ планировщика (повышение приоритета каждые %d единиц)\n", *boost)
	fmt.Printf("Количество очередей: %d\n", *numQueues)
	fmt.Printf("Временные кванты: %v\n", timeSlices)
	fmt.Printf("Длительность I/O: %d\n\n", *ioDuration)

	// Запускаем планировщик
	scheduler.Run(*maxTime)
}
