package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

const (
	defaultFilename      = "./output"
	defaultTimeInterval  = 1000 // ms
	defaultListenAddress = "0.0.0.0"
)

var csvFile *os.File
var csvWriter *csv.Writer

var writeMutex sync.Mutex

func writeHeader() {
	row := []string{
		"timeStamp", "millisSinceUnixEpoch", "quantity", "value", "unit",
	}

	writeMutex.Lock()
	defer writeMutex.Unlock()

	if csvFile != nil {
		csvWriter.Write(row)
		csvWriter.Flush()
		csvFile.Sync()
	}
}

func record(quantity, value, unit string) {
	recordWithPredefinedTimeStamp(time.Now(), quantity, value, unit)
}

func recordWithPredefinedTimeStamp(dt time.Time, quantity, value, unit string) {
	row := []string{
		dt.String(), strconv.FormatInt(dt.UnixNano()/1000000, 10), quantity, value, unit,
	}

	writeMutex.Lock()
	defer writeMutex.Unlock()

	if csvWriter != nil {
		csvWriter.Write(row)
		csvWriter.Flush()
		csvFile.Sync()
	}
}

func main() {
	var (
		filename     = flag.String("f", defaultFilename, "Filename prefix to store data to (timestamp and .csv will get appended automatically)")
		timeInterval = flag.Int64("t", defaultTimeInterval, "Time interval for measurements in ms")
	)
	flag.Parse()

	var err error
	csvFile, err = os.Create(*filename + time.Now().Format("20060102150405") + ".csv")
	if err != nil {
		log.Fatal(err)
	}

	csvWriter = csv.NewWriter(csvFile)

	cpuStat, err := cpu.Info()
	if err != nil {
		fmt.Println(err)
	}

	hostStat, err := host.Info()
	if err != nil {
		fmt.Println(err)
	}

	writeHeader()

	dt := time.Now()
	for idx, cpuinfo := range cpuStat {
		recordWithPredefinedTimeStamp(dt, "hostname", hostStat.Hostname, "-")
		recordWithPredefinedTimeStamp(dt, "CPU "+strconv.Itoa(idx)+" VendorID", cpuinfo.VendorID, "-")
		recordWithPredefinedTimeStamp(dt, "CPU "+strconv.Itoa(idx)+" Family", cpuinfo.Family, "-")
		recordWithPredefinedTimeStamp(dt, "CPU "+strconv.Itoa(idx)+" Number of cores", strconv.FormatInt(int64(cpuinfo.Cores), 10), "-")
		recordWithPredefinedTimeStamp(dt, "CPU "+strconv.Itoa(idx)+" Model Name", cpuinfo.ModelName, "-")
		recordWithPredefinedTimeStamp(dt, "CPU "+strconv.Itoa(idx)+" Speed", strconv.FormatFloat(cpuinfo.Mhz, 'f', 2, 64), "MHz")
	}

	record("Hostname", hostStat.Hostname, "-")
	record("Uptime", strconv.FormatUint(hostStat.Uptime, 10), "s")
	record("OS", hostStat.OS, "-")
	record("Platform", hostStat.Platform, "-")

	ticker := time.NewTicker(time.Duration(*timeInterval) * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case _ = <-ticker.C:
				percentage, err := cpu.Percent(0, true)
				if err != nil {
					fmt.Println(err)
				}

				vmStat, err := mem.VirtualMemory()
				if err != nil {
					fmt.Println(err)
				}

				dt := time.Now()
				for idx, cpupercent := range percentage {
					recordWithPredefinedTimeStamp(dt, "Current CPU utilization: ["+strconv.Itoa(idx)+"]", strconv.FormatFloat(cpupercent, 'f', 2, 64), "%")
				}

				if vmStat != nil {
					recordWithPredefinedTimeStamp(dt, "Total memory", strconv.FormatUint(vmStat.Total, 10), "Bytes")
					recordWithPredefinedTimeStamp(dt, "Available memory", strconv.FormatUint(vmStat.Available, 10), "Bytes")
					recordWithPredefinedTimeStamp(dt, "Percentage used memory", strconv.FormatFloat(vmStat.UsedPercent, 'f', 2, 64), "%")
				}
			}
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		case <-interrupt:
			ticker.Stop()
			done <- true
			fmt.Println("Ticker stopped")

			writeMutex.Lock()
			defer writeMutex.Unlock()

			csvWriter.Flush()
			csvFile.Sync()
			csvFile.Close()
			return
		}
	}

}
