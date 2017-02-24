package metrics

import (
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

// GetCPULoad - calculate CPU Load
func GetCPULoad() (average float64) {
	cpuload, _ := load.Avg()
	return cpuload.Load1
}

// GetMemFree - calculate free memory in MBytes
func GetMemFree() (average uint64) {
	virtualmem, _ := mem.VirtualMemory()
	return virtualmem.Free / 1024 / 1024
}
