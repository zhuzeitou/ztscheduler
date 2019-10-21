package main

import (
	"fmt"
	"time"
	"zzt/ztscheduler"
)

func main() {
	now := time.Now()
	scheduler := ztscheduler.NewScheduler(true)
	time.AfterFunc(3 * time.Second, func() {
		scheduler.AddTask(1000 * time.Millisecond, func() {
			fmt.Println("AddTask2", time.Now().Sub(now).Milliseconds())
		})
	})
	time.AfterFunc(30 * time.Second, func() {
		scheduler.Stop()
	})
	time.AfterFunc(5 * time.Second, func() {
		scheduler.RemoveTask(2)
	})
	time.AfterFunc(10 * time.Second, func() {
		scheduler.AddRecurringTask(0, 1000 * time.Millisecond, func() {
			fmt.Println("AddRecurringTask3", time.Now().Sub(now).Milliseconds())
		})
	})
	scheduler.Start()
	scheduler.AddRecurringTask(1000 * time.Millisecond, 1500 * time.Millisecond, func() {
		fmt.Println("AddRecurringTask", time.Now().Sub(now).Milliseconds())
	})
	scheduler.AddRecurringTask(2000 * time.Millisecond, 500 * time.Millisecond, func() {
		fmt.Println("AddRecurringTask2", time.Now().Sub(now).Milliseconds())
	})
	scheduler.AddTask(2000 * time.Millisecond, func() {
		fmt.Println("AddTask", time.Now().Sub(now).Milliseconds())
	})
	scheduler.AddTask(4000 * time.Millisecond, func() {
		fmt.Println("AddTask3", time.Now().Sub(now).Milliseconds())
	})
	for {
		select {
		case <-scheduler.Done():
			return
		}
	}
}