package langs

import (
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/google/uuid"
)

type Volume struct {
	Name string
}

func CreateVolume() (Volume, error) {
	volumeName := "volume-" + uuid.New().String()

	args := []string{"volume", "create"}
	args = append(args, "--name", volumeName)

	cmd := exec.Command("docker", args...)

	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		log.Println("volume create failed:", err.Error())
		return Volume{}, err
	}

	return Volume{
		Name: volumeName,
	}, nil
}

func (v *Volume) CopyFile(srcPath string, dstPath string) error {
	log.Printf("Copy file to %v:%v", v.Name, dstPath)

	task := TaskInfo{
		VolumeMountInfo: []VolumeMountInfo{
			{
				Path:   "/workdir",
				Volume: v,
			},
		},
		Name: "ubuntu",
	}
	ci, err := task.create()
	if err != nil {
		return err
	}
	defer func() {
		if err := ci.Remove(); err != nil {
			log.Printf("Failed to remove container: %v", err)
		}
	}()

	return ci.CopyFile(srcPath, path.Join("/workdir", dstPath))
}

func (v *Volume) Remove() error {
	args := []string{"volume", "rm", v.Name}

	cmd := exec.Command("docker", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}