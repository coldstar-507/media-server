package media_utils

import (
	"fmt"
	"syscall"

	"github.com/coldstar-507/media-server/internal/paths"
)

func GetStorageInfo() (uint64, uint64, error) {
	var stat syscall.Statfs_t

	path := paths.ServerFolder

	err := syscall.Statfs(path, &stat)
	if err != nil {
		fmt.Println("Error:", err)
		return .0, .0, nil
	}

	available := stat.Bavail * uint64(stat.Bsize)
	total := stat.Blocks * uint64(stat.Bsize)

	fmt.Printf("Total: %.2f GB\n", float64(total)/(1<<30))
	fmt.Printf("Available: %.2f GB\n", float64(available)/(1<<30))

	return total, available, nil
}
